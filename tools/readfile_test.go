package tools

import (
	"context"
	"os"
	"testing"
)

func TestReadFileTool_Name(t *testing.T) {
	tool := NewReadFileTool()
	if tool.Name() != "read_file" {
		t.Errorf("expected tool name 'read_file', got %q", tool.Name())
	}
}

func TestReadFileTool_Execute(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to change back to original directory: %v", err)
		}
	}()

	if err := os.WriteFile("readable.txt", []byte("file contents here"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewReadFileTool()

	t.Run("valid in-bounds path", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), []byte(`{"path":"readable.txt"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "file contents here" {
			t.Errorf("expected file contents, got %q", result)
		}
	})

	t.Run("path escaping working directory", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), []byte(`{"path":"../../../etc/passwd"}`))
		if err == nil {
			t.Error("expected an error for a path outside the working directory, got none")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), []byte(`{"path":"does-not-exist.txt"}`))
		if err == nil {
			t.Error("expected an error for a nonexistent file, got none")
		}
	})

	t.Run("malformed input JSON", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), []byte(`not json`))
		if err == nil {
			t.Error("expected an error for malformed input JSON, got none")
		}
	})
}
