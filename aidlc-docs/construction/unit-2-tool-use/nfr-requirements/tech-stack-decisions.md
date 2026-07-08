# Tech Stack Decisions — Unit 2 (Tool Use / Function Calling)

## Decision: No new dependencies
All required types (`types.ToolConfiguration`, `types.Tool`, `types.ToolMemberToolSpec`, `types.ToolUseBlock`, `types.ToolResultBlock`, `document.Interface`/`document.NewLazyDocument`) are already part of `github.com/aws/aws-sdk-go-v2/service/bedrockruntime`, already in `go.mod`. Confirms NFR5 from `requirements.md` continues to hold for this unit.

## Decision: New package `tools/` at the workspace root
Per Application Design (`components.md`), confirmed here: `tools/registry.go`, `tools/tool.go` (interface), `tools/readfile.go`. Kept separate from `cmd` so the registry and `read_file` tool are unit-testable without Cobra/Bedrock wiring, consistent with the pattern established in Unit 1 (`buildSystemContentBlocks` as a pure, separately-testable function).

## Decision: Standard library only for JSON handling
`encoding/json` for parsing accumulated tool-input fragments and building tool results — no schema-validation library is needed since the model-supplied JSON is only ever passed through to the tool's own `Execute` (which does its own targeted field extraction, e.g. `read_file` just needs a `path` string field) rather than being validated against the full schema client-side.
