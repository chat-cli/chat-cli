# Functional Design: Domain Entities - Unit 6, Confirmation and Sticky Approval Engine (#86)

## `tools/tool.go` (modified)
Adds to the existing `Tool` interface:
```go
RequiresConfirmation() bool
ConfirmationSummary(input json.RawMessage) (summary string, patternKey string, err error)
```

## `tools/permission.go` (new)
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

## `tools/approvalstore.go` (new)
```go
type ApprovalStore struct {
    repoRoot string             // "" if not inside a git repo (BR2)
    session  map[string]bool    // in-memory, this-process-only
    always   map[string][]string // loaded from disk, keyed by repo root
    path     string             // <fm.ConfigPath>/tool-approvals.yaml
}

func NewApprovalStore(configPath, repoRoot string) (*ApprovalStore, error)
func (s *ApprovalStore) IsApproved(toolName, patternKey string) bool
func (s *ApprovalStore) RecordSession(toolName, patternKey string)
func (s *ApprovalStore) RecordAlways(toolName, patternKey string) error // no-op-safe if repoRoot == ""
```
- `NewApprovalStore` loads the YAML file at construction (BR15: tolerant of missing/malformed file).
- The `key(toolName, patternKey string) string` helper (internal) joins with `"\x00"` per the business-logic-model's delimiter-collision note.

## `tools/registry.go` (modified)
```go
func (r *Registry) Dispatch(ctx context.Context, call ToolCall, gate PermissionGate) types.ToolResultBlock
```
- New logic inserted between the existing "look up tool" and "call Execute" steps (BR3-BR6).

## `cmd/permissionprompt.go` (new)
```go
type InteractivePermissionGate struct {
    store  *tools.ApprovalStore
    reader io.Reader
    writer io.Writer
}

func NewInteractivePermissionGate(store *tools.ApprovalStore, reader io.Reader, writer io.Writer) *InteractivePermissionGate
func (g *InteractivePermissionGate) Check(toolName, patternKey, summary string) tools.Decision
```
- Production call site in `cmd/chat.go` constructs it with `os.Stdin`/`os.Stdout`; tests inject `strings.Reader`/`bytes.Buffer`.

## `utils/utils.go` (modified - extraction)
```go
func FindGitBoundary(dir string) string
```
- Moved from `cmd/projectcontext.go`'s private `findGitBoundary` (identical body/behavior, 64-level cap and all) - `cmd/projectcontext.go` is refactored to call `utils.FindGitBoundary` instead of keeping its own copy.

## New Persisted File (not Go code, but a new artifact this unit introduces)
`<fm.ConfigPath>/tool-approvals.yaml` - see `business-logic-model.md`'s storage format. Created lazily on the first `RecordAlways` call; absent until then.
