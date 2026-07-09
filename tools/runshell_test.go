package tools

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunShellTool_RequiresConfirmation(t *testing.T) {
	tool := NewRunShellTool()
	if !tool.RequiresConfirmation() {
		t.Error("expected run_shell to require confirmation")
	}
}

func TestRunShellTool_ConfirmationSummary(t *testing.T) {
	tool := NewRunShellTool()

	t.Run("summary and pattern key", func(t *testing.T) {
		summary, patternKey, err := tool.ConfirmationSummary([]byte(`{"command":"git diff main"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary != "Run: git diff main" {
			t.Errorf("expected %q, got %q", "Run: git diff main", summary)
		}
		if patternKey != "git" {
			t.Errorf("expected pattern key %q, got %q", "git", patternKey)
		}
	})

	t.Run("empty command yields an empty pattern key", func(t *testing.T) {
		_, patternKey, err := tool.ConfirmationSummary([]byte(`{"command":""}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if patternKey != "" {
			t.Errorf("expected an empty pattern key, got %q", patternKey)
		}
	})
}

func TestRunShellTool_Execute(t *testing.T) {
	t.Run("returns command output on success", func(t *testing.T) {
		tool := NewRunShellTool()
		output, err := tool.Execute(context.Background(), []byte(`{"command":"echo hello"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(output, "hello") {
			t.Errorf("expected output to contain %q, got %q", "hello", output)
		}
	})

	t.Run("a non-zero exit code is embedded in output, not a Go error", func(t *testing.T) {
		tool := NewRunShellTool()
		output, err := tool.Execute(context.Background(), []byte(`{"command":"exit 1"}`))
		if err != nil {
			t.Fatalf("expected no Go error for a non-zero exit, got: %v", err)
		}
		if !strings.Contains(output, "exit code: 1") {
			t.Errorf("expected output to mention the exit code, got %q", output)
		}
	})

	t.Run("a timeout returns a Go error promptly, even for a grandchild process", func(t *testing.T) {
		tool := &RunShellTool{timeout: 50 * time.Millisecond}
		start := time.Now()
		// sh -c "sleep 5" makes `sleep` a grandchild - killing only the
		// direct child (sh) must not leave Execute blocked waiting on it.
		_, err := tool.Execute(context.Background(), []byte(`{"command":"sleep 5"}`))
		elapsed := time.Since(start)

		if err == nil {
			t.Fatal("expected an error when the command exceeds the timeout")
		}
		if elapsed > 2*time.Second {
			t.Errorf("expected Execute to return promptly after the timeout, took %s", elapsed)
		}
	})

	t.Run("output over 32KB is truncated", func(t *testing.T) {
		tool := NewRunShellTool()
		output, err := tool.Execute(context.Background(), []byte(`{"command":"yes | head -c 40000"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(output) > maxShellOutputSize+200 { // allow slack for the truncation marker text
			t.Errorf("expected output to be truncated near %d bytes, got %d", maxShellOutputSize, len(output))
		}
		if !strings.Contains(output, "truncated") {
			t.Errorf("expected a truncation marker, got output of length %d", len(output))
		}
	})
}
