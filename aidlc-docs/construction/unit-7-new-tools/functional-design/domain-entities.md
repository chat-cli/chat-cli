# Functional Design: Domain Entities - Unit 7, New Built-in Tools (#86)

## `utils/utils.go` (modified - additive)
```go
func ValidateLocalPathForWrite(filename string) (string, error)
```
Shares a new private `confineToWorkingDir(filename string) (string, error)` helper with the existing `ValidateLocalPath`, which is refactored internally to call it plus its own existence check. `ValidateLocalPath`'s public contract/behavior is unchanged.

## `tools/writefile.go` (new)
```go
type WriteFileTool struct{}
func NewWriteFileTool() *WriteFileTool
func (t *WriteFileTool) Name() string // "write_file"
func (t *WriteFileTool) Description() string
func (t *WriteFileTool) InputSchema() document.Interface // {"path": string, "content": string}
func (t *WriteFileTool) Execute(ctx context.Context, input json.RawMessage) (string, error)
func (t *WriteFileTool) RequiresConfirmation() bool // true
func (t *WriteFileTool) ConfirmationSummary(input json.RawMessage) (summary, patternKey string, err error)
```

## `tools/runshell.go` (new)
```go
const runShellTimeout = 30 * time.Second
const maxShellOutputSize = 32 * 1024

type RunShellTool struct{}
func NewRunShellTool() *RunShellTool
func (t *RunShellTool) Name() string // "run_shell"
func (t *RunShellTool) Description() string
func (t *RunShellTool) InputSchema() document.Interface // {"command": string}
func (t *RunShellTool) Execute(ctx context.Context, input json.RawMessage) (string, error)
func (t *RunShellTool) RequiresConfirmation() bool // true
func (t *RunShellTool) ConfirmationSummary(input json.RawMessage) (summary, patternKey string, err error)
```

## `tools/gitdiff.go` (new)
```go
type GitDiffTool struct{}
func NewGitDiffTool() *GitDiffTool
func (t *GitDiffTool) Name() string // "git_diff"
func (t *GitDiffTool) Description() string
func (t *GitDiffTool) InputSchema() document.Interface // {"arg": string, optional}
func (t *GitDiffTool) Execute(ctx context.Context, input json.RawMessage) (string, error)
func (t *GitDiffTool) RequiresConfirmation() bool // false
func (t *GitDiffTool) ConfirmationSummary(_ json.RawMessage) (string, string, error) // unused, never called
```

## `cmd/chat.go` (modified)
Registers the 3 new tools alongside the existing `tools.NewReadFileTool()` registration, still gated by `toolsEnabled` (Unit 8 removes that gate, not this unit).
