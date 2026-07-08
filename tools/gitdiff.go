package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
)

// GitDiffTool is a read-only built-in tool that runs `git diff` in the
// working directory. It never requires confirmation.
type GitDiffTool struct{}

// NewGitDiffTool creates a GitDiffTool.
func NewGitDiffTool() *GitDiffTool {
	return &GitDiffTool{}
}

func (t *GitDiffTool) Name() string {
	return "git_diff"
}

func (t *GitDiffTool) Description() string {
	return "Show the working tree diff (git diff) for the current repository, optionally scoped to a path or ref."
}

func (t *GitDiffTool) InputSchema() document.Interface {
	return document.NewLazyDocument(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"arg": map[string]interface{}{
				"type":        "string",
				"description": "Optional path or ref to pass to `git diff`, e.g. a filename or commit.",
			},
		},
	})
}

type gitDiffInput struct {
	Arg string `json:"arg"`
}

func (t *GitDiffTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params gitDiffInput
	// Empty input ({}) is valid - arg is optional - so only a genuine parse
	// error (malformed JSON) is treated as invalid input.
	if len(input) > 0 {
		if err := json.Unmarshal(input, &params); err != nil {
			return "", fmt.Errorf("invalid tool input: %w", err)
		}
	}

	args := []string{"diff"}
	if params.Arg != "" {
		args = append(args, params.Arg)
	}

	cmd := exec.CommandContext(ctx, "git", args...) // #nosec G204 - arg is passed as a single exec.Command argument, never shell-interpreted
	if cwd, err := os.Getwd(); err == nil {
		cmd.Dir = cwd
	}

	output, err := cmd.CombinedOutput()

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return "", fmt.Errorf("git diff failed: %s", string(output))
	}
	if err != nil {
		return "", fmt.Errorf("unable to run git diff: %w", err)
	}

	return string(output), nil
}

func (t *GitDiffTool) RequiresConfirmation() bool {
	return false
}

func (t *GitDiffTool) ConfirmationSummary(_ json.RawMessage) (string, string, error) {
	return "", "", nil
}
