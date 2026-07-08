/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveContextFilenames(t *testing.T) {
	tests := []struct {
		name        string
		configValue string
		want        []string
	}{
		{
			name:        "unset config value yields the default precedence list",
			configValue: "",
			want:        []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"},
		},
		{
			name:        "custom comma-separated list is parsed and trimmed",
			configValue: "CLAUDE.md, AGENTS.md ,README.md",
			want:        []string{"CLAUDE.md", "AGENTS.md", "README.md"},
		},
		{
			name:        "empty entries from stray commas are dropped",
			configValue: "AGENTS.md,,CLAUDE.md,",
			want:        []string{"AGENTS.md", "CLAUDE.md"},
		},
		{
			name:        "a value that trims down to nothing yields an empty list (disable case)",
			configValue: "  ,  ,",
			want:        []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveContextFilenames(tt.configValue)

			if len(got) != len(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("expected %v, got %v", tt.want, got)
					break
				}
			}
		})
	}
}

func TestFindProjectContextFile(t *testing.T) {
	t.Run("matches a candidate in cwd", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "AGENTS.md"), "hello")

		path, _, ok := findProjectContextFile(dir, []string{"AGENTS.md", "CLAUDE.md"})
		if !ok {
			t.Fatal("expected a match, got none")
		}
		if path != filepath.Join(dir, "AGENTS.md") {
			t.Errorf("expected match at %s, got %s", filepath.Join(dir, "AGENTS.md"), path)
		}
	})

	t.Run("no match anywhere yields ok=false", func(t *testing.T) {
		dir := t.TempDir()

		_, _, ok := findProjectContextFile(dir, []string{"AGENTS.md"})
		if ok {
			t.Fatal("expected no match")
		}
	})

	t.Run("a directory with a matching name is not treated as a match", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "AGENTS.md"), 0750); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		_, _, ok := findProjectContextFile(dir, []string{"AGENTS.md"})
		if ok {
			t.Fatal("expected no match - a directory should not count as a match")
		}
	})

	t.Run("matches at the git-boundary parent when cwd has no match", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, ".git"), 0750); err != nil {
			t.Fatalf("failed to create .git dir: %v", err)
		}
		writeFile(t, filepath.Join(root, "CLAUDE.md"), "hello")

		sub := filepath.Join(root, "a", "b", "c")
		if err := os.MkdirAll(sub, 0750); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}

		path, _, ok := findProjectContextFile(sub, []string{"AGENTS.md", "CLAUDE.md"})
		if !ok {
			t.Fatal("expected a match at the git-boundary parent")
		}
		if path != filepath.Join(root, "CLAUDE.md") {
			t.Errorf("expected match at %s, got %s", filepath.Join(root, "CLAUDE.md"), path)
		}
	})

	t.Run("cwd match takes precedence over a boundary match", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, ".git"), 0750); err != nil {
			t.Fatalf("failed to create .git dir: %v", err)
		}
		writeFile(t, filepath.Join(root, "AGENTS.md"), "root version")

		sub := filepath.Join(root, "a")
		if err := os.MkdirAll(sub, 0750); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}
		writeFile(t, filepath.Join(sub, "AGENTS.md"), "sub version")

		path, _, ok := findProjectContextFile(sub, []string{"AGENTS.md"})
		if !ok {
			t.Fatal("expected a match")
		}
		if path != filepath.Join(sub, "AGENTS.md") {
			t.Errorf("expected cwd match at %s, got %s", filepath.Join(sub, "AGENTS.md"), path)
		}
	})

	t.Run("no .git anywhere means only cwd is checked", func(t *testing.T) {
		// t.TempDir() is nested well outside any git repo boundary within itself.
		dir := t.TempDir()
		sub := filepath.Join(dir, "a", "b")
		if err := os.MkdirAll(sub, 0750); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}
		writeFile(t, filepath.Join(dir, "AGENTS.md"), "hello")

		_, _, ok := findProjectContextFile(sub, []string{"AGENTS.md"})
		if ok {
			t.Fatal("expected no match since no .git boundary exists between sub and dir")
		}
	})
}

func TestLoadProjectContext(t *testing.T) {
	t.Run("reads and trims content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "AGENTS.md")
		writeFile(t, path, "  hello world  \n")

		content, truncated, originalSize, err := loadProjectContext(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content != "hello world" {
			t.Errorf("expected trimmed content %q, got %q", "hello world", content)
		}
		if truncated {
			t.Error("did not expect truncation")
		}
		if originalSize != len("  hello world  \n") {
			t.Errorf("expected originalSize %d, got %d", len("  hello world  \n"), originalSize)
		}
	})

	t.Run("truncates content over 32KB", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "AGENTS.md")
		big := strings.Repeat("a", 32*1024+100)
		writeFile(t, path, big)

		content, truncated, originalSize, err := loadProjectContext(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !truncated {
			t.Error("expected truncation")
		}
		if len(content) != 32*1024 {
			t.Errorf("expected truncated content of exactly 32KB, got %d bytes", len(content))
		}
		if originalSize != len(big) {
			t.Errorf("expected originalSize %d, got %d", len(big), originalSize)
		}
	})

	t.Run("content under 32KB is not truncated", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "AGENTS.md")
		writeFile(t, path, strings.Repeat("a", 100))

		_, truncated, _, err := loadProjectContext(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if truncated {
			t.Error("did not expect truncation")
		}
	})

	t.Run("an unreadable path returns an error", func(t *testing.T) {
		dir := t.TempDir()
		// A directory is not a readable "file" in the sense this function needs.
		_, _, _, err := loadProjectContext(dir)
		if err == nil {
			t.Fatal("expected an error reading a directory as a file")
		}
	})
}

func TestResolveAndLoadProjectContext(t *testing.T) {
	t.Run("finds and loads a match", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "AGENTS.md"), "you are helpful")

		content, sourcePath, truncated, found := resolveAndLoadProjectContext(dir, []string{"AGENTS.md", "CLAUDE.md"})
		if !found {
			t.Fatal("expected a match")
		}
		if content != "you are helpful" {
			t.Errorf("expected content %q, got %q", "you are helpful", content)
		}
		if sourcePath != filepath.Join(dir, "AGENTS.md") {
			t.Errorf("expected sourcePath %s, got %s", filepath.Join(dir, "AGENTS.md"), sourcePath)
		}
		if truncated {
			t.Error("did not expect truncation")
		}
	})

	t.Run("an empty-after-trim file is skipped in favor of the next candidate", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "AGENTS.md"), "   \n  ")
		writeFile(t, filepath.Join(dir, "CLAUDE.md"), "real content")

		content, sourcePath, _, found := resolveAndLoadProjectContext(dir, []string{"AGENTS.md", "CLAUDE.md"})
		if !found {
			t.Fatal("expected a match (should have skipped the vacuous AGENTS.md)")
		}
		if content != "real content" {
			t.Errorf("expected content %q, got %q", "real content", content)
		}
		if sourcePath != filepath.Join(dir, "CLAUDE.md") {
			t.Errorf("expected sourcePath %s, got %s", filepath.Join(dir, "CLAUDE.md"), sourcePath)
		}
	})

	t.Run("no candidates match anywhere yields found=false", func(t *testing.T) {
		dir := t.TempDir()

		_, _, _, found := resolveAndLoadProjectContext(dir, []string{"AGENTS.md", "CLAUDE.md"})
		if found {
			t.Fatal("expected no match")
		}
	})

	t.Run("empty candidate list yields found=false without touching the filesystem", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "AGENTS.md"), "content")

		_, _, _, found := resolveAndLoadProjectContext(dir, []string{})
		if found {
			t.Fatal("expected no match when the candidate list is empty")
		}
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
