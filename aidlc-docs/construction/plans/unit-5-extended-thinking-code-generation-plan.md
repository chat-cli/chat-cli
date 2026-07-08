# Unit 5 — Extended Thinking / Reasoning Mode — Code Generation Plan

**Unit**: Unit 5, issue [#85](https://github.com/chat-cli/chat-cli/issues/85) | **Story**: 5.1 (final unit)
**FR/NFR coverage**: FR5.1-FR5.4 (NFR Requirements/Design skipped: no new security surface)
**Dependencies**: Unit 2's `accumulateStream`/`blockAccumulator` (extended, not duplicated)

## ⚠️ Carried Forward From Functional Design
The `reasoning_config`/`budget_tokens` request shape is unverified (untyped SDK field, no live source available in this environment). Implemented as designed; flagged again in the unit summary as the top thing to confirm with real credentials.

## TDD Order

### Step 1-2 — `buildReasoningConfig` helper
- [ ] Test (`cmd/reasoning_test.go`, new): `enabled=false` → `nil` (NFR1, unchanged request shape); `enabled=true, budget=1024` → a `document.Interface` that round-trips (via `document.Interface`'s marshal) to `{"reasoning_config":{"type":"enabled","budget_tokens":1024}}`.
- [ ] Implement `buildReasoningConfig(enabled bool, budgetTokens int32) document.Interface` in `cmd/reasoning.go` (new file), using `document.NewLazyDocument(map[string]interface{}{...})`.

### Step 3-4 — Extend `blockAccumulator`/`accumulateStream` for reasoning blocks
- [ ] Test (extend `cmd/streamaccumulate_test.go`): a stream with `ContentBlockDeltaMemberReasoningContent` text deltas at one index and text deltas at another → `accumulateStream` returns a finalized message containing both a `ContentBlockMemberReasoningContent` block (with accumulated text + signature, if a signature delta was present) and the regular text block, in original index order; reasoning text is passed to a **new**, separate callback (not the existing `onText`) so callers can render it distinctly (Rule 2).
- [ ] Implement: extend `blockKind` with `blockKindReasoning`; extend `blockAccumulator` with a `reasoningText`/`reasoningSignature` pair; extend `accumulateStream`'s delta switch to handle `ContentBlockDeltaMemberReasoningContent` (itself wrapping `ReasoningContentBlockDeltaMemberText`/`...MemberSignature`/`...MemberRedactedContent`); add an `onReasoning utils.StreamingOutputHandler`-shaped parameter to `accumulateStream`'s signature (redacted content is preserved in the finalized block but not passed to `onReasoning`, per Rule 5).
- [ ] Update `runChatTurnWithTools`'s signature to accept and forward the new `onReasoning` callback.

### Step 5 — Wire `chat.go`
- [ ] Add `--thinking` (bool, default `false`) and `--thinking-budget` (int32, default `1024`) persistent flags on `rootCmd` (same dual-registration pattern as `--system`/`--tools`).
- [ ] Set `converseStreamInput.AdditionalModelRequestFields = buildReasoningConfig(thinking, thinkingBudget)`.
- [ ] Pass an `onReasoning` callback to `runChatTurnWithTools` that prints reasoning text dimmed/prefixed (`\033[90m[thinking] ...\033[0m`, matching the existing user-echo ANSI convention), distinct from the final answer.

### Step 6 — Wire `prompt.go`
- [ ] Add `--thinking`/`--thinking-budget` flags (mirrors `prompt`'s own flag-duplication pattern).
- [ ] Set `AdditionalModelRequestFields` the same way on both `converseInput` and `converseStreamInput`.
- [ ] **No-stream path**: after getting the response, check `response.Value.Content` for a `*types.ContentBlockMemberReasoningContent` block; if present, print its text dimmed/prefixed before the final answer text.
- [ ] **Streaming path**: `prompt`'s streaming path uses `utils.ProcessStreamingOutput` (not `accumulateStream` — that's `chat`-only per Application Design). Extend `ProcessStreamingOutput` minimally to also invoke a callback for reasoning text deltas (`ContentBlockDeltaMemberReasoningContent`), or add a small parallel handling branch — decide the smaller, less invasive change while writing the code, since `ProcessStreamingOutput` has its own existing test (`TestProcessStreamingOutput`) that must keep passing unmodified.

### Step 7 — Full test suite, lint, coverage, integration
- [ ] `make test`, `make lint`, `make test-coverage`, `make cli && go test -tags=integration -v .`.

### Step 8 — Documentation
- [ ] `README.md`/`docs/usage.md`: document `--thinking`/`--thinking-budget`, the max-tokens interaction, and the explicit caveat that the reasoning-config shape is best-effort and may need adjustment for some providers.

### Step 9 — Unit Documentation Summary
- [ ] `aidlc-docs/construction/unit-5-extended-thinking/code/summary.md` — re-flag the unverified request shape prominently.

## Story Traceability
- Story 5.1 (all 3 acceptance criteria) → Steps 1-6 implement and test them; Step 7 verifies; Step 8 satisfies NFR6.
