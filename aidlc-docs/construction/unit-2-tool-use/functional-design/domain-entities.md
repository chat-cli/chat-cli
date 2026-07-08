# Domain Entities — Unit 2 (Tool Use / Function Calling)

These are internal Go types (package `tools`, plus one accumulator type local to `cmd`), not persisted entities — no database schema changes in this unit (confirmed in `application-design/component-dependency.md`).

## `tools.Tool` (interface, from Application Design, unchanged)
```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() document.Interface // JSON schema, matches types.ToolInputSchemaMemberJson's Value type
    Execute(ctx context.Context, input json.RawMessage) (result string, err error)
}
```
Correction from `application-design/component-methods.md`: `InputSchema()` returns `document.Interface` (from `github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document`, already part of the AWS SDK already in `go.mod` — NFR5 still holds, no new dependency), built via that package's `document.NewLazyDocument(goValue)` from a plain `map[string]interface{}` JSON-schema literal — not `*types.ToolInputSchema`, to match what `types.ToolInputSchemaMemberJson.Value` actually expects.

## `tools.Registry`
```go
type Registry struct { tools map[string]Tool }

func NewRegistry() *Registry
func (r *Registry) Register(tool Tool)
func (r *Registry) ToolConfiguration() *types.ToolConfiguration // nil if no tools registered
func (r *Registry) Dispatch(ctx context.Context, call ToolCall) types.ToolResultBlock
```

## `tools.ToolCall` (new — the finalized, parsed form of an accumulated `ToolUseBlock`)
```go
type ToolCall struct {
    Name      string
    ToolUseID string
    Input     json.RawMessage
}
```
Relationship: produced by `cmd`'s streaming accumulator (see below) once a tool-use content block is finalized; consumed by `Registry.Dispatch`.

## `tools.ReadFileTool` (unchanged from Application Design)
```go
type ReadFileTool struct{}
func NewReadFileTool() *ReadFileTool
func (t *ReadFileTool) Name() string          // "read_file"
func (t *ReadFileTool) Description() string
func (t *ReadFileTool) InputSchema() document.Interface // {"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}
func (t *ReadFileTool) Execute(ctx context.Context, input json.RawMessage) (string, error)
```

## `cmd` package: streaming content-block accumulator (new, internal to `chat.go`/a new small file)
```go
// blockAccumulator tracks one in-progress content block by its stream index.
type blockAccumulator struct {
    kind      blockKind // text | toolUse
    text      strings.Builder
    toolName  string
    toolUseID string
    toolInput strings.Builder // raw JSON fragments, concatenated
}
```
Not exported outside `cmd` — this is streaming-protocol bookkeeping, not a reusable domain concept, so it doesn't belong in the `tools` package per Application Design's component boundaries.

## Relationships
```
Registry *-- Tool          (Registry holds many Tools, keyed by Name())
Registry ..> ToolCall       (Dispatch takes a ToolCall as input)
Registry ..> types.ToolResultBlock (Dispatch produces this SDK type as output)
ReadFileTool ..|> Tool      (implements)
blockAccumulator ..> ToolCall (cmd finalizes an accumulator into a ToolCall before calling Dispatch)
```
