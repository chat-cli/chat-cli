# Unit of Work

chat-cli is a monolith — these 5 units are logical modules within the single `chat-cli` binary, not independently deployable services. All 5 ship together in the same release.

## Unit 1 — System Prompt Support
- **Scope**: FR1.1-FR1.4, Story 1.1, issue #81
- **Files touched**: `cmd/chat.go`, `cmd/prompt.go`, `config/config.go`
- **New components**: None — direct additions to existing request-building code and `FileManager`'s supported-keys list
- **Definition of done**: `--system` flag and `system-prompt` config key both work with correct precedence; `SystemContentBlocks` sent only when set; existing behavior unchanged when unset

## Unit 2 — Tool Use / Function Calling
- **Scope**: FR2.1-FR2.5, Stories 2.1 + 2.2, issue #82
- **Files touched**: `cmd/chat.go`; new `tools/` package
- **New components**: `tools.Tool` (interface), `tools.Registry`, `tools.ReadFileTool`, `utils.ValidateLocalPath`
- **Definition of done**: `chat` builds a `ToolConfig` when tools are registered; tool-use round trips work end-to-end via the built-in `read_file` tool; unknown-tool and execution-failure cases return error results to the model instead of crashing; tool-call turns persist through the existing `ChatRepository`

## Unit 3 — Prompt Caching
- **Scope**: FR3.1-FR3.4, Story 3.1, issue #83
- **Files touched**: `cmd/chat.go`, `cmd/prompt.go`; new cache-point helper in `utils`
- **New components**: Cache-point helper (`utils`)
- **Definition of done**: Cache checkpoints inserted after system prompt / piped document content; graceful one-time fallback retry on rejection; no user-visible failure when caching isn't supported by the selected model

## Unit 4 — Native Document Input
- **Scope**: FR4.1-FR4.4, Story 4.1, issue #84
- **Files touched**: `cmd/prompt.go`; new `utils.ReadDocument` (reuses `utils.ValidateLocalPath` from Unit 2, or introduces it if Unit 4 lands first)
- **New components**: `utils.ReadDocument`
- **Definition of done**: `--document` flag attaches supported file types as `ContentBlockMemberDocument`; unsupported extensions/out-of-bounds paths rejected with a clear error before calling Bedrock; `--image` behavior completely unchanged; both flags usable together

## Unit 5 — Extended Thinking / Reasoning Mode
- **Scope**: FR5.1-FR5.4, Story 5.1, issue #85
- **Files touched**: `cmd/chat.go`, `cmd/prompt.go`
- **New components**: None — direct additions to existing request-building code and streaming-output rendering
- **Definition of done**: `--thinking` flag sets `AdditionalModelRequestFields`; reasoning content rendered distinctly from the final answer; unsupported-model rejection surfaced clearly; behavior unchanged when flag is unset

## Code Organization Note (Brownfield)
No new top-level directory structure beyond the one new `tools/` package introduced in Unit 2 (see `application-design/components.md`). All other changes extend existing files in `cmd/`, `config/`, and `utils/` in place, per the "Application Code: Workspace root" rule in `aidlc-state.md`.
