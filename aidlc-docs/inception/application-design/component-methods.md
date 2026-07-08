# Component Methods

Method signatures only — business rules and error-path details are defined in per-unit Functional Design during Construction, per this stage's scope.

## `tools.Tool` (interface)

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() *types.ToolInputSchema // Bedrock's tool input schema type
    Execute(ctx context.Context, input json.RawMessage) (result string, err error)
}
```

## `tools.Registry`

```go
type Registry struct { /* unexported set of registered Tools */ }

func NewRegistry() *Registry
func (r *Registry) Register(tool Tool)
func (r *Registry) ToolConfig() *types.ToolConfig
func (r *Registry) Dispatch(ctx context.Context, toolUse types.ToolUseBlock) types.ContentBlock // returns a ToolResultBlock (success or error variant)
```

- **`ToolConfig()`**: builds Bedrock's `ToolConfig` from all registered tools' `Name()`/`Description()`/`InputSchema()`. Returns `nil` if no tools are registered (so chat-cli behaves exactly as today when the registry is empty).
- **`Dispatch()`**: looks up the tool by `toolUse.Name`; if not found, returns an error `ToolResultBlock`; if found, calls `Execute` and wraps a success or error `ToolResultBlock` around the result.

## `tools.ReadFileTool`

```go
type ReadFileTool struct{}

func NewReadFileTool() *ReadFileTool
func (t *ReadFileTool) Name() string          // "read_file"
func (t *ReadFileTool) Description() string
func (t *ReadFileTool) InputSchema() *types.ToolInputSchema // { "path": string }
func (t *ReadFileTool) Execute(ctx context.Context, input json.RawMessage) (string, error)
```

- **`Execute()`**: unmarshals `{"path": "..."}` from `input`, calls `utils.ValidateLocalPath`, reads and returns file contents, or returns the validation/read error.

## `utils.ValidateLocalPath` (new exported function)

```go
func ValidateLocalPath(filename string) (fullPath string, err error)
```

- Confines resolution to the current working directory (same rule as today's inline logic in `ReadImage`). `ReadImage` is refactored to call this instead of duplicating the checks.

## Cache-point helper (new function(s), exact package location TBD in Code Generation — likely `utils`)

```go
func WithCachePoint(blocks []types.ContentBlock) []types.ContentBlock
// appends a cachePoint content block after the given content

func SendWithCacheFallback(
    ctx context.Context,
    send func(ctx context.Context) (*bedrockruntime.ConverseStreamOutput, error),
    sendWithoutCache func(ctx context.Context) (*bedrockruntime.ConverseStreamOutput, error),
) (*bedrockruntime.ConverseStreamOutput, error)
// calls send(); on a cache-point-rejection error, logs a warning and calls sendWithoutCache() once
```

- Exact function shape (streaming vs. non-streaming variants) is refined during Functional Design for the Prompt Caching unit — this is the high-level contract only.

## `config.FileManager` (existing, extended)

```go
// No new methods. GetConfigValue(key, flagValue, defaultValue) is reused as-is
// for the new "system-prompt" key, exactly like "model-id"/"custom-arn" today.
```

## `cmd/prompt.go` document-attachment path (existing file, extended — not a new component)

```go
// New helper, likely in utils, parallel to ReadImage:
func ReadDocument(filename string) (data []byte, format string, err error)
// validates via ValidateLocalPath, checks against the supported-format allow-list
// (pdf, csv, doc, docx, xls, xlsx, html, txt, md), returns bytes + detected format
```

---

# Initiative 3 Component Methods (#86)

## `tools.Tool` (interface, extended)

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() document.Interface
    Execute(ctx context.Context, input json.RawMessage) (result string, err error)

    // New:
    RequiresConfirmation() bool
    // ConfirmationSummary returns human-readable text describing what this
    // specific call will do (for the prompt), and a coarse pattern key used
    // for sticky-approval matching (base command for run_shell, directory
    // for write_file). Only meaningful when RequiresConfirmation() is true.
    ConfirmationSummary(input json.RawMessage) (summary string, patternKey string, err error)
}
```

- `read_file`/`git_diff`: `RequiresConfirmation()` returns `false`; `ConfirmationSummary` is unused (never called by `Dispatch` for a non-destructive tool).

## `tools.Registry` (extended)

```go
func (r *Registry) Dispatch(ctx context.Context, call ToolCall, gate PermissionGate) types.ToolResultBlock
```

- Before calling `Execute`, if `tool.RequiresConfirmation()`, calls `tool.ConfirmationSummary(call.Input)` then `gate.Check(tool.Name(), patternKey, summary)`. A `Deny` decision short-circuits to a declined-action error `ToolResultBlock` (FR5.3) without calling `Execute`.

## `tools.PermissionGate` (new interface)

```go
type Decision int
const (
    DecisionAllowOnce Decision = iota
    DecisionAllowSession
    DecisionAllowAlways
    DecisionDeny
)

type PermissionGate interface {
    // Check returns a decision for this specific call. Implementations may
    // block on user input; a prior session/always approval matching
    // toolName+patternKey short-circuits without prompting.
    Check(toolName, patternKey, summary string) Decision
}
```

## `tools.ApprovalStore` (new)

```go
type ApprovalStore struct { /* unexported: in-memory session set + persisted-tier backing */ }

func NewApprovalStore(repoRoot string) (*ApprovalStore, error)
func (s *ApprovalStore) IsApproved(toolName, patternKey string) bool
func (s *ApprovalStore) RecordSession(toolName, patternKey string)
func (s *ApprovalStore) RecordAlways(toolName, patternKey string) error // persists
```

## `tools.WriteFileTool`, `tools.RunShellTool`, `tools.GitDiffTool` (new)

```go
type WriteFileTool struct{}
func NewWriteFileTool() *WriteFileTool
func (t *WriteFileTool) Execute(ctx context.Context, input json.RawMessage) (string, error)
// unmarshals {"path": "...", "content": "..."}, validates via utils.ValidateLocalPath, writes

type RunShellTool struct{}
func NewRunShellTool() *RunShellTool
func (t *RunShellTool) Execute(ctx context.Context, input json.RawMessage) (string, error)
// unmarshals {"command": "..."}, runs via sh -c with a timeout, returns combined truncated output

type GitDiffTool struct{}
func NewGitDiffTool() *GitDiffTool
func (t *GitDiffTool) Execute(ctx context.Context, input json.RawMessage) (string, error)
// unmarshals {"arg": "..."} (optional), runs `git diff [arg]`, returns raw output or a clear error
```

## `utils.FindGitBoundary` (new exported function, extracted from `cmd/projectcontext.go`)

```go
func FindGitBoundary(dir string) string
// identical behavior to #88's private findGitBoundary: walks upward from dir
// (stat-only, capped at 64 levels) looking for the first ancestor containing
// a .git entry; returns "" if none found. cmd/projectcontext.go is refactored
// to call this instead of its own private copy.
```

## `cmd.InteractivePermissionGate` (new, implements `tools.PermissionGate`)

```go
type InteractivePermissionGate struct { /* unexported: *tools.ApprovalStore */ }

func NewInteractivePermissionGate(store *tools.ApprovalStore) *InteractivePermissionGate
func (g *InteractivePermissionGate) Check(toolName, patternKey, summary string) tools.Decision
// prints the summary, blocks reading a once/session/always/deny choice,
// records session/always decisions into the store, returns the decision
```
