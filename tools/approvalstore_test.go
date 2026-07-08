package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApprovalStore_SessionTier(t *testing.T) {
	t.Run("not approved before any record", func(t *testing.T) {
		s, err := NewApprovalStore(t.TempDir(), "/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.IsApproved("run_shell", "git") {
			t.Error("expected no approval before any record")
		}
	})

	t.Run("approved after RecordSession", func(t *testing.T) {
		s, err := NewApprovalStore(t.TempDir(), "/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s.RecordSession("run_shell", "git")
		if !s.IsApproved("run_shell", "git") {
			t.Error("expected approval after RecordSession")
		}
	})

	t.Run("approvals are independent per tool+pattern pair", func(t *testing.T) {
		s, err := NewApprovalStore(t.TempDir(), "/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s.RecordSession("run_shell", "git")

		if s.IsApproved("run_shell", "npm") {
			t.Error("did not expect a different pattern key to be approved")
		}
		if s.IsApproved("write_file", "git") {
			t.Error("did not expect a different tool with the same pattern key to be approved")
		}
	})
}

func TestApprovalStore_AlwaysTier(t *testing.T) {
	t.Run("RecordAlways persists with 0600 permissions and is loaded by a fresh store", func(t *testing.T) {
		configPath := t.TempDir()

		s1, err := NewApprovalStore(configPath, "/repo/a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := s1.RecordAlways("run_shell", "git"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		storePath := filepath.Join(configPath, "tool-approvals.yaml")
		info, err := os.Stat(storePath)
		if err != nil {
			t.Fatalf("expected %s to exist: %v", storePath, err)
		}
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Errorf("expected 0600 permissions, got %o", perm)
		}

		s2, err := NewApprovalStore(configPath, "/repo/a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !s2.IsApproved("run_shell", "git") {
			t.Error("expected a fresh store pointed at the same configPath/repoRoot to see the persisted approval")
		}
	})

	t.Run("always approvals do not leak across different repository roots", func(t *testing.T) {
		configPath := t.TempDir()

		s1, err := NewApprovalStore(configPath, "/repo/a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := s1.RecordAlways("run_shell", "git"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		s2, err := NewApprovalStore(configPath, "/repo/b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s2.IsApproved("run_shell", "git") {
			t.Error("expected an always-approval in one repo root to not apply in a different one")
		}
	})

	t.Run("a missing store file yields zero approvals, not an error", func(t *testing.T) {
		s, err := NewApprovalStore(t.TempDir(), "/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.IsApproved("run_shell", "git") {
			t.Error("expected no approvals from a missing store file")
		}
	})

	t.Run("a malformed store file yields zero approvals, not a fatal error", func(t *testing.T) {
		configPath := t.TempDir()
		storePath := filepath.Join(configPath, "tool-approvals.yaml")
		if err := os.WriteFile(storePath, []byte("not: valid: yaml: [["), 0600); err != nil {
			t.Fatal(err)
		}

		s, err := NewApprovalStore(configPath, "/repo")
		if err != nil {
			t.Fatalf("expected malformed store file to degrade gracefully, got error: %v", err)
		}
		if s.IsApproved("run_shell", "git") {
			t.Error("expected no approvals from a malformed store file")
		}
	})

	t.Run("RecordAlways is a safe no-op when repoRoot is empty", func(t *testing.T) {
		configPath := t.TempDir()
		s, err := NewApprovalStore(configPath, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := s.RecordAlways("run_shell", "git"); err != nil {
			t.Fatalf("expected RecordAlways to be a safe no-op outside a repo, got error: %v", err)
		}
		if s.IsApproved("run_shell", "git") {
			t.Error("expected no approval to be recorded when repoRoot is empty")
		}
		if _, err := os.Stat(filepath.Join(configPath, "tool-approvals.yaml")); err == nil {
			t.Error("expected no store file to be created when repoRoot is empty")
		}
	})
}
