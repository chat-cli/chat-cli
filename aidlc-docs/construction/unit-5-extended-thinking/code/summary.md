# Unit 5 — Extended Thinking / Reasoning Mode — Code Generation Summary

**Issue**: [#85](https://github.com/chat-cli/chat-cli/issues/85) | **Story**: 5.1 (final unit) | **Plan**: `aidlc-docs/construction/plans/unit-5-extended-thinking-code-generation-plan.md`

## ⚠️ Top Item To Verify With Real Credentials
The `reasoning_config`/`budget_tokens` request shape (`cmd/reasoning.go`'s `buildReasoningConfig`) is **unverified** — `AdditionalModelRequestFields` is an untyped SDK field with no way to confirm the correct JSON shape by static inspection. If `--thinking` doesn't trigger reasoning output against a real model, this is the first place to check. Everything else in this unit (response-side parsing, rendering, multi-turn signature preservation) is built against fully-typed, SDK-confirmed structures.

## Files Created
- `cmd/reasoning.go` — `buildReasoningConfig` (100% covered), `printReasoningBlock` (used by `prompt`'s no-stream path; not unit-tested, consistent with this codebase's existing convention of not testing terminal-print side effects)
- `cmd/reasoning_test.go`

## Files Modified
- `cmd/toolloop.go` — extended `blockKind`/`blockAccumulator`/`accumulateStream` (Unit 2) to handle `ContentBlockDeltaMemberReasoningContent` (text + signature deltas), rather than duplicating the accumulation machinery; `runChatTurnWithTools` now takes and forwards an `onReasoning` callback
- `utils/utils.go` — `ProcessStreamingOutput` now takes a second `reasoningHandler` callback and safely type-switches on delta kind. **Found and fixed a latent bug while extending this**: the previous code did an unchecked type assertion (`v.Value.Delta.(*types.ContentBlockDeltaMemberText)`) that would have panicked the moment any non-text delta (e.g. reasoning) appeared in a `prompt` stream — now a safe type switch.
- `cmd/root.go` — `--thinking`/`--thinking-budget` persistent flags
- `cmd/chat.go` — sets `AdditionalModelRequestFields`; renders reasoning dimmed/prefixed via a new `onReasoning` callback, with a reset before the final answer text starts
- `cmd/prompt.go` — same flags; renders reasoning in both the no-stream (`printReasoningBlock`) and streaming (`onReasoning` callback) paths
- All existing `accumulateStream`/`runChatTurnWithTools` test call sites updated for the new callback parameter
- `README.md`, `docs/usage.md` — documented `--thinking`/`--thinking-budget`, the `--max-tokens` interaction, and the request-shape caveat

## Verification
- `make test`: all green, no regressions (existing Unit 2 tests updated for the new signature, still pass)
- `make lint`: clean
- `go test -tags=integration -v .`: all 7 pass; `--thinking`/`--thinking-budget` confirmed in `--help` on both commands
- Coverage: total statement coverage held at 66.3% (steady, `printReasoningBlock` untested per existing print-function convention)
- **Not verified in this environment**: an actual extended-thinking request/response against real Bedrock — no AWS credentials available. This is the same category of gap noted in every prior unit's summary, but carries more weight here given the unverified request shape above.

## Story 5.1 Acceptance Criteria Status
- [x] No `--thinking` → no `AdditionalModelRequestFields` sent, behavior unchanged
- [x] `--thinking` on a supported model → reasoning rendered distinct from the final answer (both `chat` and `prompt`, streaming and non-streaming)
- [x] Rejection surfaces a clear error (existing `log.Fatalf` pattern, no retry, no silent failure)
- [x] Reasoning text + signature preserved across multi-turn history in `chat` (via the extended `accumulateStream`), per the SDK's documented multi-turn requirement

## Initiative Status
This completes Code Generation for all 5 units (#81-#85). Next: **Build and Test** (the cross-unit integration phase per `core-workflow.md`), then INCEPTION/CONSTRUCTION for this initiative is complete.
