# Functional Design: Business Logic Model - Unit 6, Confirmation and Sticky Approval Engine (#86)

## Context
Verified against the actual current codebase rather than assumed:
- `tools/tool.go`/`tools/registry.go` (Initiative 1 Unit 2, reviewed in Application Design) - the interface/`Dispatch` extension points this unit builds on.
- `utils/utils.go`'s `StringPrompt` - **deliberately not reused**. It launches a full `bubbletea` text-input widget for free-form entry (chat messages), which is the wrong shape for a discrete once/session/always/deny choice and harder to unit-test (a full TUI program lifecycle vs. a plain `io.Reader`). This unit implements its own minimal prompt instead.
- `config.FileManager.ConfigPath` (exported, `config/config.go:16`) - reused as the base directory for the new persisted approvals file, consistent with where chat-cli already keeps its config/DB.

## Core Components (from Application Design, detailed here)

### 1. Extended `tools.Tool` Interface
```go
RequiresConfirmation() bool
ConfirmationSummary(input json.RawMessage) (summary string, patternKey string, err error)
```
- `ConfirmationSummary` is called with the **same raw input** `Execute` will receive - it must parse it itself (duplicating the unmarshal `Execute` also does). This is accepted duplication, not a design flaw: keeping the two methods independent means a summary-generation bug can never accidentally skip validation that `Execute` would have caught, and vice versa - each method is a complete, independently-testable unit.
- If `ConfirmationSummary` itself fails to parse the input (malformed JSON from the model), that's treated as a **denial-equivalent**: `Registry.Dispatch` returns an error `ToolResultBlock` immediately, without ever calling the gate or `Execute` - a call that can't even be summarized can't be meaningfully confirmed.

### 2. `PermissionGate` / `Decision`
```go
type Decision int
const (
    DecisionAllowOnce Decision = iota
    DecisionAllowSession
    DecisionAllowAlways
    DecisionDeny
)
type PermissionGate interface {
    Check(toolName, patternKey, summary string) Decision
}
```

### 3. `ApprovalStore` - Two Tiers
- **Session tier**: `map[string]bool` keyed by `toolName + "\x00" + patternKey` (a byte unlikely to appear in either, avoiding a delimiter-collision edge case that a simple `":"` join could hit if a pattern key ever contained a colon). Lives only in memory, created fresh per `chat` process, discarded on exit.
- **Always tier**: persisted at `<fm.ConfigPath>/tool-approvals.yaml`, structure:
  ```yaml
  repos:
    /absolute/path/to/repo-root:
      - "run_shell:git"
      - "write_file:src"
  ```
  Keyed by the **absolute** repo-root path (from `utils.FindGitBoundary`) at the top level - this is the per-repository scoping FR7.1 requires. Within a repo's list, entries are `"toolName:patternKey"` strings (simpler to serialize/diff than nested YAML, and human-readable/editable if someone wants to hand-edit the file).
  - File permissions: `0600` (owner read/write only) on creation, matching NFR1.
  - Loaded once at `ApprovalStore` construction (one read), only written to on a `RecordAlways` call - `IsApproved` never touches disk.
  - A repo root that isn't yet a key in the file, or a missing file entirely, is treated as "no always-approvals for this repo" - never an error.

### 4. Pattern Key Derivation (business rules detailed in `business-rules.md`)
- `run_shell`: base command = the first whitespace-separated token of the command string.
- `write_file`: containing directory, **relative to the repo root** (not absolute) - keeps the persisted file portable if the repo is later cloned to a different absolute path, and keeps entries human-readable.

### 5. `InteractivePermissionGate`
```go
type InteractivePermissionGate struct {
    store  *tools.ApprovalStore
    reader io.Reader // defaults to os.Stdin in production
    writer io.Writer // defaults to os.Stdout in production
}
```
- Constructor takes the reader/writer explicitly (not defaulted internally) so Code Generation's tests can inject a `strings.Reader`/`bytes.Buffer` and exercise the full prompt-parsing logic without a real terminal - resolves the "testability of the interactive prompt" question Application Design deferred.
- `Check` flow: look up `store.IsApproved(toolName, patternKey)` first (covers both tiers - `ApprovalStore` itself checks session then always internally) → if approved, return `DecisionAllowOnce`-equivalent (no prompt) → else print the summary + choice prompt to `writer`, read one line from `reader`, parse the first non-whitespace character case-insensitively (`o`/`s`/`a`/`n`, default to deny on empty/EOF/unrecognized input - a fail-closed default per NFR1) → on session/always, call the matching `store.Record*` method → return the decision.

## Data Flow

```
Registry.Dispatch(call, gate)
  -> tool := lookup(call.Name)
  -> if !tool.RequiresConfirmation(): Execute directly
  -> else:
       summary, patternKey, err := tool.ConfirmationSummary(call.Input)
       if err != nil: return error ToolResultBlock (never reaches the gate)
       decision := gate.Check(tool.Name(), patternKey, summary)
       if decision == Deny: return declined-error ToolResultBlock
       else: Execute(call.Input)  // Once/Session/Always all mean "proceed now"

gate.Check(toolName, patternKey, summary):
  -> if store.IsApproved(toolName, patternKey): return AllowOnce (skip prompt)
  -> print summary + prompt to writer
  -> read one line from reader, parse choice
  -> on Session: store.RecordSession(...)
  -> on Always: store.RecordAlways(...) (writes to disk)
  -> return the parsed Decision
```
