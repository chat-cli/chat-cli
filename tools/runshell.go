package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
)

// defaultRunShellTimeout bounds how long a run_shell command may execute
// before being killed (BR8) - no way for a single call to hang chat forever.
const defaultRunShellTimeout = 30 * time.Second

// maxShellOutputSize is the size (in bytes) at which run_shell's combined
// stdout+stderr is truncated before being returned to the model (BR9).
const maxShellOutputSize = 32 * 1024

// RunShellTool is a destructive built-in tool that executes an arbitrary
// shell command in chat-cli's working directory. Every call must pass
// through a PermissionGate (RequiresConfirmation returns true) before
// Execute runs - there is no command allowlist, the gate is the control.
type RunShellTool struct {
	timeout time.Duration
}

// NewRunShellTool creates a RunShellTool with the default 30s timeout.
func NewRunShellTool() *RunShellTool {
	return &RunShellTool{timeout: defaultRunShellTimeout}
}

func (t *RunShellTool) Name() string {
	return "run_shell"
}

func (t *RunShellTool) Description() string {
	return "Run a shell command in the current working directory."
}

func (t *RunShellTool) InputSchema() document.Interface {
	return document.NewLazyDocument(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to run.",
			},
		},
		"required": []interface{}{"command"},
	})
}

type runShellInput struct {
	Command string `json:"command"`
}

func (t *RunShellTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params runShellInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid tool input: %w", err)
	}

	timeout := t.timeout
	if timeout <= 0 {
		timeout = defaultRunShellTimeout
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.Command("sh", "-c", params.Command) // #nosec G204 - arbitrary command execution is this tool's purpose, gated by PermissionGate
	if cwd, err := os.Getwd(); err == nil {
		cmd.Dir = cwd
	}
	// Run in its own process group so a timeout can kill the whole group
	// (e.g. a grandchild like `sleep` spawned by `sh -c`), not just the
	// shell itself - killing only the shell leaves such children running
	// and holding the output pipe open, which would otherwise block
	// CombinedOutput() below well past the intended timeout.
	setProcessGroup(cmd)

	type result struct {
		output []byte
		err    error
	}
	done := make(chan result, 1)
	go func() {
		output, err := cmd.CombinedOutput()
		done <- result{output: output, err: err}
	}()

	select {
	case <-runCtx.Done():
		killProcessGroup(cmd)
		<-done // wait for the goroutine to unblock now that the group is dead
		return "", fmt.Errorf("command timed out after %s", timeout)

	case res := <-done:
		output := truncateShellOutput(string(res.output))

		var exitErr *exec.ExitError
		if errors.As(res.err, &exitErr) {
			return fmt.Sprintf("[exit code: %d]\n%s", exitErr.ExitCode(), output), nil
		}
		if res.err != nil {
			// The shell itself couldn't be started - a real tool-level
			// failure, distinct from the command it was asked to run failing.
			return "", fmt.Errorf("unable to run command: %w", res.err)
		}

		return output, nil
	}
}

func truncateShellOutput(output string) string {
	if len(output) <= maxShellOutputSize {
		return output
	}
	return output[:maxShellOutputSize] + "\n... (output truncated)"
}

func (t *RunShellTool) RequiresConfirmation() bool {
	return true
}

func (t *RunShellTool) ConfirmationSummary(input json.RawMessage) (string, string, error) {
	var params runShellInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", "", fmt.Errorf("invalid tool input: %w", err)
	}

	summary := "Run: " + params.Command

	patternKey := ""
	if fields := strings.Fields(params.Command); len(fields) > 0 {
		patternKey = fields[0]
	}

	return summary, patternKey, nil
}
