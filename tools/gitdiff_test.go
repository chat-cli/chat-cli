package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitDiffTool_RequiresConfirmation(t *testing.T) {
	tool := NewGitDiffTool()
	if tool.RequiresConfirmation() {
		t.Error("expected git_diff to not require confirmation - it's read-only")
	}
}

func initTestGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	runGit("init", "-q")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test")

	filePath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filePath, []byte("line1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runGit("add", "file.txt")
	runGit("commit", "-q", "-m", "initial")

	if err := os.WriteFile(filePath, []byte("line1\nline2\n"), 0600); err != nil {
		t.Fatal(err)
	}

	return dir
}

func chdirToRepo(t *testing.T, dir string) {
	t.Helper()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to change back to original directory: %v", err)
		}
	})
}

func TestGitDiffTool_Execute(t *testing.T) {
	t.Run("no arg runs a plain git diff", func(t *testing.T) {
		chdirToRepo(t, initTestGitRepo(t))
		tool := NewGitDiffTool()

		output, err := tool.Execute(context.Background(), []byte(`{}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(output, "line2") {
			t.Errorf("expected diff output to contain the new line, got %q", output)
		}
	})

	t.Run("an arg is passed through", func(t *testing.T) {
		chdirToRepo(t, initTestGitRepo(t))
		tool := NewGitDiffTool()

		output, err := tool.Execute(context.Background(), []byte(`{"arg":"file.txt"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(output, "line2") {
			t.Errorf("expected diff output to contain the new line, got %q", output)
		}
	})

	t.Run("outside a git repository returns git's own error", func(t *testing.T) {
		chdirToRepo(t, t.TempDir())
		tool := NewGitDiffTool()

		_, err := tool.Execute(context.Background(), []byte(`{}`))
		if err == nil {
			t.Fatal("expected an error when not inside a git repository")
		}
	})
}
