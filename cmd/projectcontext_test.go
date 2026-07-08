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
		configSet   bool
		want        []string
	}{
		{
			name:        "unset config key yields the default precedence list",
			configValue: "",
			configSet:   false,
			want:        []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"},
		},
		{
			name:        "explicitly set empty config value disables discovery",
			configValue: "",
			configSet:   true,
			want:        []string{},
		},
		{
			name:        "custom comma-separated list is parsed and trimmed",
			configValue: "CLAUDE.md, AGENTS.md ,README.md",
			configSet:   true,
			want:        []string{"CLAUDE.md", "AGENTS.md", "README.md"},
		},
		{
			name:        "empty entries from stray commas are dropped",
			configValue: "AGENTS.md,,CLAUDE.md,",
			configSet:   true,
			want:        []string{"AGENTS.md", "CLAUDE.md"},
		},
		{
			name:        "a value that trims down to nothing yields an empty list (disable case)",
			configValue: "  ,  ,",
			configSet:   true,
			want:        []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveContextFilenames(tt.configValue, tt.configSet)

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

	t.Run("a symlink to a regular file is treated as a match", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "instructions.md")
		writeFile(t, target, "hello")
		link := filepath.Join(dir, "AGENTS.md")
		if err := os.Symlink(target, link); err != nil {
			t.Skip("symlinks not supported in this environment")
		}

		path, _, ok := findProjectContextFile(dir, []string{"AGENTS.md"})
		if !ok {
			t.Fatal("expected a match via symlink")
		}
		if path != link {
			t.Errorf("expected match at %s, got %s", link, path)
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

func TestFormatProjectContextDisplayPath(t *testing.T) {
	t.Run("returns basename when file is in cwd", func(t *testing.T) {
		cwd := "/repo"
		source := "/repo/AGENTS.md"
		got := formatProjectContextDisplayPath(cwd, source)
		if got != "AGENTS.md" {
			t.Errorf("expected AGENTS.md, got %q", got)
		}
	})

	t.Run("returns relative path when file is at repo root from nested cwd", func(t *testing.T) {
		cwd := "/repo/a/b"
		source := "/repo/AGENTS.md"
		want, err := filepath.Rel(cwd, source)
		if err != nil {
			t.Fatalf("unexpected rel error: %v", err)
		}
		got := formatProjectContextDisplayPath(cwd, source)
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
