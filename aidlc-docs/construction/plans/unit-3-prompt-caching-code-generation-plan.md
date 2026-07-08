# Unit 3 — Prompt Caching — Code Generation Plan

**Unit**: Unit 3, issue [#83](https://github.com/chat-cli/chat-cli/issues/83) | **Story**: 3.1
**FR/NFR coverage**: FR3.1-FR3.4, NFR1, NFR2 (NFR Requirements/Design skipped for this unit)
**Dependencies**: Unit 1 (system prompt) complete; SDK upgrade (prerequisite) complete

## Approach (from Functional Design)
- `stripSystemCachePoints`/`stripContentCachePoints` remove cache-point blocks from an already-built request, rather than maintaining two parallel "with cache"/"without cache" variants — works uniformly whether system caching, document caching, or both are active.
- Retry wraps the **outer** call only (`runChatTurnWithTools` in `chat.go`; `svc.Converse`/`svc.ConverseStream` in `prompt.go`), not each internal tool round-trip within `chat`'s loop, per Functional Design's accepted edge-case tradeoff (a cache-point rejection is a request-shape problem that manifests on the first call of a turn in practice).

## TDD Order

### Step 1-2 — `withSystemCachePoint` helper
- [x] Test (`cmd/promptcache_test.go`, new): empty/nil input → nil output (no system prompt → nothing to cache, NFR1); non-empty input → same blocks plus one `*types.SystemContentBlockMemberCachePoint{Value: CachePointBlock{Type: CachePointTypeDefault}}` appended.
- [x] Implement `withSystemCachePoint(blocks []types.SystemContentBlock) []types.SystemContentBlock` in `cmd/promptcache.go` (new file).

### Step 3-4 — `stripSystemCachePoints` / `stripContentCachePoints` helpers
- [x] Test: a `[]types.SystemContentBlock` containing a text block and a cache-point block → strip returns only the text block; a `[]types.ContentBlock` containing text + cache-point + text → strip returns the two text blocks in original order. Also test the no-cache-point-present case returns the input unchanged (idempotent).
- [x] Implement both strip functions in `cmd/promptcache.go`.

### Step 5-6 — `hasSystemCachePoint` guard (avoid a pointless retry when nothing was cached)
- [x] Test: nil/text-only `System` → `false`; `System` containing a cache-point block → `true`.
- [x] Implement `hasSystemCachePoint(blocks []types.SystemContentBlock) bool`. Used to skip the retry entirely when there was nothing to strip (e.g., no system prompt was set), avoiding a wasted duplicate call.

### Step 7-8 — `buildQuestionContent` (Design Decision 2: split document from question)
- [x] Test (`cmd/promptcache_test.go`): empty document → single text block with the question, unchanged from today's shape; non-empty document → `[documentTextBlock, cachePointBlock, questionTextBlock]` in that order.
- [x] Implement `buildQuestionContent(document, question string) []types.ContentBlock` in `cmd/promptcache.go`.

### Step 9 — Wire into `cmd/prompt.go`
- [x] Stop doing `prompt += document`; keep `document` and `prompt` (the question) separate.
- [x] Build `userMsg.Content` via `buildQuestionContent(document, prompt)` instead of the single hardcoded text block (image block, if any, still appended after, unchanged).
- [x] Set `converseInput.System`/`converseStreamInput.System` via `withSystemCachePoint(buildSystemContentBlocks(systemPrompt))` instead of the bare `buildSystemContentBlocks(...)` call from Unit 1.
- [x] Wrap both the `--no-stream` (`svc.Converse`) and streaming (`svc.ConverseStream`) calls: on error, if `hasSystemCachePoint(...)` was true OR the message content had a cache-point block, log a warning, strip cache points from both `System` and the user message's `Content`, and retry once; surface the second attempt's error normally (existing `log.Fatalf` pattern) if it also fails.

### Step 10 — Wire into `cmd/chat.go`
- [x] Set `converseStreamInput.System` via `withSystemCachePoint(buildSystemContentBlocks(systemPrompt))`.
- [x] Wrap the `runChatTurnWithTools` call: on error, if a cache point was present, log a warning, strip it from `converseStreamInput.System`, and call `runChatTurnWithTools` once more with the same (already-mutated-in-place, but not yet appended-to on a first-call failure) `converseStreamInput`; surface the second error normally if it also fails.
- [x] `chat.go` does not get document caching — no document-input capability exists in `chat` yet (issue #46), consistent with Functional Design's Rule 5 scope note.

### Step 11 — Full test suite, lint, coverage, integration
- [x] `make test`, `make lint`, `make test-coverage` (expect improvement, all new helpers are pure and fully tested), `make cli && go test -tags=integration -v .`.

### Step 12 — Documentation
- [x] `README.md`/`docs/usage.md`: note that system prompts and piped documents are now cached automatically where the model supports it, with no flag needed and no behavior change when caching isn't supported (graceful fallback).

### Step 13 — Unit Documentation Summary
- [x] `aidlc-docs/construction/unit-3-prompt-caching/code/summary.md`.

## Story Traceability
- Story 3.1 (all 4 acceptance criteria) → Steps 1-10 implement and test them; Step 11 verifies; Step 12 satisfies NFR6.
