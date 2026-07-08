# Unit 2 — Tool Use / Function Calling — Code Generation Summary

**Issue**: [#82](https://github.com/chat-cli/chat-cli/issues/82) | **Stories**: 2.1, 2.2 | **Plan**: `aidlc-docs/construction/plans/unit-2-tool-use-code-generation-plan.md`

## Product Decision Made During Planning
Tool use is **opt-in via `--tools`** (default off), not always-on, because Bedrock exposes no "supports tool use" capability check and enabling it unconditionally would risk breaking `chat` for models that reject the `ToolConfig` field. Confirmed with the user before generation began.

## Files Created
- `tools/tool.go` — `Tool` interface
- `tools/registry.go` — `Registry`, `ToolCall`, `Dispatch` (fail-closed: unknown tool and execution errors both become error `ToolResultBlock`s, never a panic or crash)
- `tools/readfile.go` — the built-in `read_file` tool (read-only, path-confined via `utils.ValidateLocalPath`)
- `tools/registry_test.go`, `tools/readfile_test.go` — 90% coverage on the new package
- `cmd/toolloop.go` — `accumulateStream` (pure function over a `<-chan types.ConverseStreamOutput`, testable without mocking SDK internals), `runChatTurnWithTools` (the round-trip orchestration loop, injectable `converseStreamFunc`), `finalizeToolCall`, the round-trip cap constant
- `cmd/streamaccumulate_test.go`, `cmd/toolturn_test.go`, `cmd/toolloop_test.go` — cover text-only streams, tool-use streams, the no-tool-use fast path, and the round-trip cap

## Files Modified
- `utils/utils.go` — extracted `ValidateLocalPath` from `ReadImage`'s inline logic (pure refactor; `TestReadImage` passes unmodified, proving no behavior change)
- `utils/utils_test.go` — added `TestValidateLocalPath`
- `cmd/root.go` — registered `--tools` bool flag (default `false`)
- `cmd/chat.go` — builds a `tools.Registry` (populated only when `--tools` is set), replaces the direct `svc.ConverseStream` + `utils.ProcessStreamingOutput` call with `runChatTurnWithTools`
- `README.md`, `docs/usage.md` — documented `--tools` and the `read_file` tool

## Verification
- `make test`: all packages pass, no regressions
- `make lint`: clean
- `go test -tags=integration -v .`: all 7 integration tests pass; `--tools` confirmed present in `--help` output on both `chat-cli` and `chat-cli chat`
- Coverage: `cmd` 8.0% → 18.7%; new `tools` package at 90.0%; total statement coverage 52.6% → 62.4%
- **Not verified in this environment**: an actual live tool-call round-trip against real Bedrock — no AWS credentials are available in this sandbox. The full protocol (streaming accumulation, dispatch, round-trip loop, cap) is covered by unit tests using hand-constructed SDK event values on real Go channels; the only untested seam is the thin `sendFn` closure wrapping the real `svc.ConverseStream` call itself, which has no logic beyond forwarding to `GetStream().Events()`.

## Story Acceptance Criteria Status
- **Story 2.1** (tool round-trip loop): all criteria covered by `accumulateStream`/`runChatTurnWithTools` tests — unknown tool, execution failure, and persistence (only final text + prompt saved, verified by inspection since `chatRepo.Create` call sites are unchanged from Unit 1) all hold. Requires `--tools` to be passed, per the flagged decision above.
- **Story 2.2** (`read_file` tool): covered by `tools/readfile_test.go` — valid read, out-of-bounds rejection, nonexistent file, malformed input all pass.
