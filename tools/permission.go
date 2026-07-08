package tools

// Decision is the outcome of a PermissionGate.Check call for a specific
// destructive tool invocation.
type Decision int

const (
	// DecisionAllowOnce permits this specific call, without recording any
	// sticky approval.
	DecisionAllowOnce Decision = iota
	// DecisionAllowSession permits this call and all future calls matching
	// the same tool+pattern key for the remainder of the current process.
	DecisionAllowSession
	// DecisionAllowAlways permits this call and persists the approval so
	// future chat sessions in the same repository skip the prompt too.
	DecisionAllowAlways
	// DecisionDeny refuses this call. Denials are never sticky.
	DecisionDeny
)

// PermissionGate decides whether a destructive tool call may proceed.
// Registry.Dispatch consults it before calling Execute on any tool whose
// RequiresConfirmation returns true.
type PermissionGate interface {
	Check(toolName, patternKey, summary string) Decision
}
