/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/chat-cli/chat-cli/tools"
)

// InteractivePermissionGate is the concrete, terminal-facing PermissionGate
// implementation: it shows the confirmation prompt, reads the user's
// once/session/always/deny choice, and records the decision via an
// ApprovalStore.
type InteractivePermissionGate struct {
	store  *tools.ApprovalStore
	reader io.Reader
	writer io.Writer
}

// NewInteractivePermissionGate creates a gate backed by store, reading
// prompt responses from reader and writing prompt text to writer. In
// production these are os.Stdin/os.Stdout; tests inject a strings.Reader/
// bytes.Buffer so the parsing logic is testable without a real terminal.
func NewInteractivePermissionGate(store *tools.ApprovalStore, reader io.Reader, writer io.Writer) *InteractivePermissionGate {
	return &InteractivePermissionGate{store: store, reader: reader, writer: writer}
}

// Check implements tools.PermissionGate. A prior session/always approval
// (BR6) short-circuits without printing anything or reading any input.
// Otherwise it prints the summary and choice prompt, reads one line, and
// parses the first non-whitespace character (BR12): o/s/a/n, case
// insensitive. "a" (always) is only offered/accepted when the store's
// approvals can actually be persisted (BR10 - inside a git repository);
// any unrecognized or empty/EOF input defaults to deny, fail-closed (BR12).
func (g *InteractivePermissionGate) Check(toolName, patternKey, summary string) tools.Decision {
	if g.store.IsApproved(toolName, patternKey) {
		return tools.DecisionAllowOnce
	}

	offerAlways := g.store.CanRecordAlways()

	fmt.Fprintf(g.writer, "%s\n", summary)
	if offerAlways {
		fmt.Fprint(g.writer, "Allow this action? [o]nce / [s]ession / [a]lways / [n]o: ")
	} else {
		fmt.Fprint(g.writer, "Allow this action? [o]nce / [s]ession / [n]o: ")
	}

	line, _ := bufio.NewReader(g.reader).ReadString('\n')
	choice := strings.ToLower(strings.TrimSpace(line))

	var decision tools.Decision
	switch {
	case choice == "o":
		decision = tools.DecisionAllowOnce
	case choice == "s":
		decision = tools.DecisionAllowSession
	case choice == "a" && offerAlways:
		decision = tools.DecisionAllowAlways
	case choice == "n":
		decision = tools.DecisionDeny
	default:
		decision = tools.DecisionDeny
	}

	switch decision {
	case tools.DecisionAllowSession:
		g.store.RecordSession(toolName, patternKey)
	case tools.DecisionAllowAlways:
		if err := g.store.RecordAlways(toolName, patternKey); err != nil {
			fmt.Fprintf(g.writer, "warning: failed to persist approval: %v\n", err)
		}
	}

	return decision
}
