# Unit 2 — Tool Use / Function Calling — Code Generation Plan

**Unit**: Unit 2, issue [#82](https://github.com/chat-cli/chat-cli/issues/82) | **Stories**: 2.1, 2.2
**FR/NFR coverage**: FR2.1-FR2.5, NFR1-NFR5, SEC-1..SEC-4, REL-1
**Dependencies**: Unit 1 (complete) — no hard blockers, but reuses `buildSystemContentBlocks`-style extraction pattern

## Key SDK Facts Verified by Direct Inspection (not guessed)
- `ConverseStreamInput.ToolConfig *types.ToolConfiguration{Tools []types.Tool, ToolChoice}`
- `types.Tool` union satisfied by `types.ToolMemberToolSpec{Value: types.ToolSpecification{Name *string, Description *string, InputSchema types.ToolInputSchema}}`
- `types.ToolInputSchema` union satisfied by `types.ToolInputSchemaMemberJson{Value: document.Interface}` (build via `document.NewLazyDocument(goValue)`, package `.../bedrockruntime/document`)
- Streaming tool-use arrives as `ContentBlockStartEvent{ContentBlockIndex *int32, Start: ContentBlockStartMemberToolUse{ToolUseBlockStart{Name, ToolUseId}}}`, then `ContentBlockDeltaEvent{ContentBlockIndex *int32, Delta: ContentBlockDeltaMemberToolUse{ToolUseBlockDelta{Input *string}}}` (fragments to concatenate), then `ContentBlockStopEvent{ContentBlockIndex *int32}`
- `MessageStopEvent.StopReason` — `types.StopReasonToolUse = "tool_use"` signals a tool round-trip is needed
- `output.GetStream().Events()` returns `<-chan types.ConverseStreamOutput` — a plain channel of an interface I can construct test values for directly; the SDK explicitly documents `ConverseStreamEventStream`/`ConverseStreamOutputReader` as mockable for testing, but `ConverseStreamOutput.eventStream` is unexported with no public constructor, so **this plan abstracts the "get the next event channel" step behind a local function type** rather than fighting SDK internals — same spirit as Unit 1's pure-function extraction
- `types.ToolResultBlock{Content []types.ToolResultContentBlock, ToolUseId *string, Status types.ToolResultStatus}`; `types.ToolResultContentBlockMemberText{Value string}`; `types.ToolResultStatusSuccess`/`types.ToolResultStatusError`

## TDD Order

### Step 1-2 — `tools.Tool` interface + empty `Registry` (failing test, then implementation)
- [ ] Test (`tools/registry_test.go`): `NewRegistry().ToolConfiguration()` returns `nil` when no tools registered (proves NFR1 — chat behaves exactly as today until a tool exists).
- [ ] Implement `tools/tool.go` (interface) and `tools/registry.go` (`NewRegistry`, `Register`, `ToolConfiguration`, empty `Dispatch` stub). Confirm test passes.

### Step 3-4 — `Registry.Dispatch`: unknown tool (Rule 2 / SEC-2)
- [ ] Test: `Dispatch` with a `ToolCall{Name: "nonexistent"}` returns `types.ToolResultBlock{Status: ToolResultStatusError, ...}` mentioning the unknown name.
- [ ] Implement the unknown-tool branch of `Dispatch`. Confirm test passes.

### Step 5-6 — `Registry.Dispatch`: successful execution
- [ ] Test: register a fake `Tool` (test double returning fixed output), `Dispatch` a matching `ToolCall`, assert `Status: ToolResultStatusSuccess` and the correct text content.
- [ ] Implement the success branch. Confirm test passes.

### Step 7-8 — `Registry.Dispatch`: tool execution error (Rule 3 / SEC-3)
- [ ] Test: fake `Tool` whose `Execute` returns an error; assert `Dispatch` returns `Status: ToolResultStatusError` with the error text, no panic.
- [ ] Implement the error-wrapping branch. Confirm test passes.

### Step 9-10 — `utils.ValidateLocalPath` (extracted from `ReadImage`, SEC-1)
- [ ] Test (`utils/utils_test.go`, new cases): valid relative path within CWD → returns absolute path, no error; path escaping CWD (`../../etc/passwd`-style) → error; matches the existing `TestReadImage/path_traversal_attempt` case in spirit but calling the new function directly.
- [ ] Implement `ValidateLocalPath(filename string) (string, error)` by extracting the existing inline logic from `ReadImage`; refactor `ReadImage` to call it. **Re-run the full existing `TestReadImage` suite unmodified — it must still pass without any test changes**, proving this is a pure refactor (no behavior change).

### Step 11-12 — `tools.ReadFileTool` (Story 2.2, FR2.5)
- [ ] Test (`tools/readfile_test.go`): `Name()` returns `"read_file"`; `Execute` with a valid in-CWD path returns file contents; `Execute` with a path escaping the CWD returns an error (not a panic, not a successful read); `Execute` with a nonexistent path returns an error.
- [ ] Implement `tools/readfile.go` using `utils.ValidateLocalPath`. Confirm tests pass.

### Step 13-14 — Tool-call finalization from accumulated JSON (Rule 4)
- [ ] Test (`cmd/toolloop_test.go`): a helper that takes accumulated Name/ToolUseId/raw-JSON-string and returns a `tools.ToolCall` — valid JSON parses correctly; malformed JSON returns an error (not a panic).
- [ ] Implement `finalizeToolCall(name, toolUseID, rawInput string) (tools.ToolCall, error)` in `cmd/toolloop.go`.

### Step 15-16 — Stream accumulation (the core algorithm, `functional-design/business-logic-model.md`)
- [ ] Test: build a `chan types.ConverseStreamOutput`, push hand-constructed events onto it in order (`ContentBlockStart` for text, `ContentBlockDelta` text fragments, `ContentBlockStart` for tool-use, `ContentBlockDelta` tool-use fragments, `ContentBlockStop` x2, `MessageStop` with `StopReasonToolUse`), close it, and assert `accumulateStream` returns the right finalized text, the right `[]tools.ToolCall`, and the right `StopReason`. Add a second case with `StopReasonEndTurn` and no tool-use blocks (normal turn, proves NFR1-style "nothing changes when no tools are involved").
- [ ] Implement `accumulateStream(events <-chan types.ConverseStreamOutput, onText utils.StreamingOutputHandler) (types.Message, []tools.ToolCall, types.StopReason, error)` tracking blocks by `ContentBlockIndex` in a `map[int32]*blockAccumulator`. Confirm tests pass.

### Step 17-18 — Round-trip orchestration loop (Rule 5 / REL-1, the cap)
- [ ] Test: define a local `converseStreamFunc func(ctx, *bedrockruntime.ConverseStreamInput) (<-chan types.ConverseStreamOutput, error)` type; supply a fake that returns a channel signaling `StopReasonToolUse` every time (simulating a runaway model); assert `runChatTurnWithTools` stops after exactly 10 round-trips with a clear, non-panicking error. Add a second case: a fake returning `StopReasonEndTurn` on the first call, no tool round-trip at all — assert it returns immediately with the right text.
- [ ] Implement `runChatTurnWithTools(ctx, send converseStreamFunc, input *bedrockruntime.ConverseStreamInput, registry *tools.Registry, onText utils.StreamingOutputHandler) (string, error)` in `cmd/toolloop.go`, using `accumulateStream` + `Registry.Dispatch` + the Rule 5 cap. Confirm tests pass.

### Step 19 — Wire into `cmd/chat.go`

**⚠️ Decision surfaced during planning, not resolved in Functional Design — flagging before generation rather than silently picking one:**

Bedrock's `GetFoundationModel` doesn't expose a "supports tool use" capability flag the way it does for text/image/streaming (confirmed: only `OutputModalities`, `InputModalities`, `ResponseStreamingSupported` exist on `ModelDetails`, per Unit 1's SDK inspection). If tool use were made **unconditionally active** for every `chat` session, a model that doesn't support tools would likely reject the request outright (Bedrock validation error) — that would **break existing usage for non-tool-capable models**, directly violating NFR1 (backward compatibility), the one hard constraint from `requirements.md`.

**Plan (pending your confirmation)**: make tool use **opt-in** via a new `--tools` boolean flag on `chat` (default `false`), consistent with how `--no-stream` is opt-in/out elsewhere in this codebase. Only when `--tools` is set does the registry get built and `ToolConfig` get attached — with the flag unset, `chat` behaves exactly as before this unit, for every model, which is the safest default and keeps NFR1 airtight. This does mean Story 2.1's "the model calls a tool" scenario requires `--tools` to be passed — a small deviation from the story's phrasing (which didn't specify a flag), made here for safety; flag if you'd rather tool use be always-on and accept the compatibility risk for non-tool models.

- [ ] Add `--tools` bool flag to `chat` (root + chat command, same dual-registration pattern as `--system` in Unit 1).
- [ ] Build a `tools.Registry` and register `tools.NewReadFileTool()` **only if `--tools` is set**.
- [ ] Set `converseStreamInput.ToolConfig = registry.ToolConfiguration()` only in that case (nil otherwise, unchanged from today).
- [ ] Replace the direct `svc.ConverseStream` + `utils.ProcessStreamingOutput` call in the tty-loop with a call to `runChatTurnWithTools`, passing a real `converseStreamFunc` closure that wraps `svc.ConverseStream` and returns `output.GetStream().Events()`. When `--tools` is unset, `runChatTurnWithTools` is still used (with a `nil`/empty registry) so there's only one code path to maintain, but it behaves identically to the old direct call since `ToolConfig` is nil and no `StopReasonToolUse` can occur.
- [ ] `prompt.go` and `image.go` are **not** touched — tool use is `chat`-only per Application Design/Functional Design.

### Step 20 — Full test suite, lint, coverage, integration
- [ ] `make test` — all green, no regressions.
- [ ] `make lint` — clean.
- [ ] `make test-coverage` — `cmd` and new `tools` package coverage recorded; expect `tools` package to start high given the pure-logic-first design.
- [ ] `make cli && go test -tags=integration -v` — all pass; manually smoke-test `chat-cli` with a prompt that should trigger `read_file` (e.g. "what's in go.mod?") against a real or represents a dry-run check of flag wiring only if AWS credentials aren't available in this environment — see summary.md for what was actually exercised.

### Step 21 — Documentation
- [ ] `README.md`/`docs/usage.md`: brief mention that `chat` can now use tools, starting with a built-in `read_file` tool, no flag needed to enable it (it's automatic once at least one tool is registered — there's no `--tools` opt-out in this pass, consistent with the stories; note this explicitly since it's a slight product-shape decision worth surfacing to the user in the summary).

### Step 22 — Unit Documentation Summary
- [ ] `aidlc-docs/construction/unit-2-tool-use/code/summary.md`.

## Story Traceability
- Story 2.1 (tool round-trip loop) → Steps 1-10, 13-19
- Story 2.2 (`read_file` tool) → Steps 9-12
