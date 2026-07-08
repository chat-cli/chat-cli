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
