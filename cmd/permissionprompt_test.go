/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chat-cli/chat-cli/tools"
)

func newTestApprovalStore(t *testing.T, repoRoot string) *tools.ApprovalStore {
	t.Helper()
	store, err := tools.NewApprovalStore(t.TempDir(), repoRoot)
	if err != nil {
		t.Fatalf("unexpected error creating ApprovalStore: %v", err)
	}
	return store
}

func TestInteractivePermissionGate_ChoiceParsing(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  tools.Decision
	}{
		{"once, lowercase", "o\n", tools.DecisionAllowOnce},
		{"once, uppercase", "O\n", tools.DecisionAllowOnce},
		{"session, lowercase", "s\n", tools.DecisionAllowSession},
		{"session, uppercase", "S\n", tools.DecisionAllowSession},
		{"always, lowercase", "a\n", tools.DecisionAllowAlways},
		{"always, uppercase", "A\n", tools.DecisionAllowAlways},
		{"deny, lowercase", "n\n", tools.DecisionDeny},
		{"deny, uppercase", "N\n", tools.DecisionDeny},
		{"empty input defaults to deny", "\n", tools.DecisionDeny},
		{"EOF (no newline at all) defaults to deny", "", tools.DecisionDeny},
		{"unrecognized character defaults to deny", "x\n", tools.DecisionDeny},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestApprovalStore(t, "/repo")
			var out bytes.Buffer
			gate := NewInteractivePermissionGate(store, strings.NewReader(tt.input), &out)

			got := gate.Check("run_shell", "git", "run `git diff`")
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestInteractivePermissionGate_SessionChoiceIsRemembered(t *testing.T) {
	store := newTestApprovalStore(t, "/repo")
	var out bytes.Buffer
	gate := NewInteractivePermissionGate(store, strings.NewReader("s\n"), &out)

	gate.Check("run_shell", "git", "run `git diff`")

	if !store.IsApproved("run_shell", "git") {
		t.Error("expected a session choice to record a session approval")
	}
}

func TestInteractivePermissionGate_AlwaysChoiceIsPersisted(t *testing.T) {
	store := newTestApprovalStore(t, "/repo")
	var out bytes.Buffer
	gate := NewInteractivePermissionGate(store, strings.NewReader("a\n"), &out)

	gate.Check("run_shell", "git", "run `git diff`")

	if !store.IsApproved("run_shell", "git") {
		t.Error("expected an always choice to record an approval")
	}
}

func TestInteractivePermissionGate_AlwaysNotOfferedOutsideARepo(t *testing.T) {
	store := newTestApprovalStore(t, "") // no repo root
	var out bytes.Buffer
	gate := NewInteractivePermissionGate(store, strings.NewReader("a\n"), &out)

	got := gate.Check("run_shell", "git", "run `git diff`")

	if got != tools.DecisionDeny {
		t.Errorf("expected 'a' to be treated as unrecognized (deny) when always isn't offered, got %v", got)
	}
	if strings.Contains(strings.ToLower(out.String()), "always") {
		t.Errorf("expected the prompt to not mention 'always' when outside a repo, got: %s", out.String())
	}
}

func TestInteractivePermissionGate_PriorApprovalSkipsThePrompt(t *testing.T) {
	store := newTestApprovalStore(t, "/repo")
	store.RecordSession("run_shell", "git")
	var out bytes.Buffer
	// No input available - if the gate tried to read a prompt response, it
	// would get EOF and deny. A pre-existing approval must short-circuit
	// before any read happens.
	gate := NewInteractivePermissionGate(store, strings.NewReader(""), &out)

	got := gate.Check("run_shell", "git", "run `git diff`")

	if got == tools.DecisionDeny {
		t.Error("expected a pre-existing approval to skip the prompt and allow the call")
	}
	if out.Len() != 0 {
		t.Errorf("expected no prompt to be printed when an approval already exists, got: %s", out.String())
	}
}
