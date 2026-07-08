package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chdirForTest(t *testing.T, dir string) {
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

func TestWriteFileTool_RequiresConfirmation(t *testing.T) {
	tool := NewWriteFileTool()
	if !tool.RequiresConfirmation() {
		t.Error("expected write_file to require confirmation")
	}
}

func TestWriteFileTool_ConfirmationSummary(t *testing.T) {
	chdirForTest(t, t.TempDir())

	t.Run("summary and pattern key without a repo", func(t *testing.T) {
		tool := NewWriteFileTool()
		summary, patternKey, err := tool.ConfirmationSummary([]byte(`{"path":"sub/foo.txt","content":"hello"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(summary, "sub/foo.txt") || !strings.Contains(summary, "hello") {
			t.Errorf("expected summary to mention path and content, got %q", summary)
		}
		if patternKey != "sub" {
			t.Errorf("expected pattern key %q, got %q", "sub", patternKey)
		}
	})

	t.Run("content over 4KB is truncated in the summary", func(t *testing.T) {
		tool := NewWriteFileTool()
		bigContent := strings.Repeat("a", 4*1024+100)
		input := []byte(`{"path":"big.txt","content":"` + bigContent + `"}`)
		summary, _, err := tool.ConfirmationSummary(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(summary, bigContent) {
			t.Error("expected content over 4KB to be truncated in the summary")
		}
		if !strings.Contains(summary, "truncated") {
			t.Errorf("expected a truncation note, got %q", summary)
		}
	})

	t.Run("path traversal attempt returns an error", func(t *testing.T) {
		tool := NewWriteFileTool()
		_, _, err := tool.ConfirmationSummary([]byte(`{"path":"../../../etc/passwd","content":"x"}`))
		if err == nil {
			t.Error("expected an error for a path outside the working directory")
		}
	})
}

func TestWriteFileTool_Execute(t *testing.T) {
	chdirForTest(t, t.TempDir())
	tool := NewWriteFileTool()

	t.Run("creates a new file", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), []byte(`{"path":"new.txt","content":"hello world"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, readErr := os.ReadFile("new.txt")
		if readErr != nil {
			t.Fatalf("expected file to exist: %v", readErr)
		}
		if string(data) != "hello world" {
			t.Errorf("expected file contents %q, got %q", "hello world", string(data))
		}
	})

	t.Run("overwrites an existing file", func(t *testing.T) {
		if err := os.WriteFile("existing.txt", []byte("old content"), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := tool.Execute(context.Background(), []byte(`{"path":"existing.txt","content":"new content"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, readErr := os.ReadFile("existing.txt")
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(data) != "new content" {
			t.Errorf("expected file contents %q, got %q", "new content", string(data))
		}
	})

	t.Run("rejects a path outside the working directory", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), []byte(`{"path":"../../../etc/passwd","content":"x"}`))
		if err == nil {
			t.Error("expected an error for a path outside the working directory")
		}
	})

	t.Run("supports a nested new directory-free path", func(t *testing.T) {
		dir := t.TempDir()
		chdirForTest(t, dir)
		_, err := tool.Execute(context.Background(), []byte(`{"path":"nested.txt","content":"x"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, statErr := os.Stat(filepath.Join(dir, "nested.txt")); statErr != nil {
			t.Errorf("expected file to exist: %v", statErr)
		}
	})
}
