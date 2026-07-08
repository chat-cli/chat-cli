# Unit 3 — Prompt Caching — Code Generation Summary

**Issue**: [#83](https://github.com/chat-cli/chat-cli/issues/83) | **Story**: 3.1 | **Plan**: `aidlc-docs/construction/plans/unit-3-prompt-caching-code-generation-plan.md`

## Files Created
- `cmd/promptcache.go` — `withSystemCachePoint`, `hasSystemCachePoint`, `hasContentCachePoint`, `stripSystemCachePoints`, `stripContentCachePoints`, `buildQuestionContent` — all pure functions, all at 100% test coverage
- `cmd/promptcache_test.go`

## Files Modified
- `cmd/prompt.go` — no longer merges piped document into the prompt string (`prompt += document` removed); builds message content via `buildQuestionContent(document, prompt)` so a document and the question are separate, cacheable content blocks; both the `--no-stream` and streaming request paths now attach a system cache point (when a system prompt is set) and retry once without any cache points on error
- `cmd/chat.go` — attaches a system cache point the same way; wraps the `runChatTurnWithTools` call with the same retry-once-without-cache policy
- `README.md`, `docs/usage.md` — documented that caching is automatic (no flag) with graceful fallback

## Verification
- `make test`: all green, no regressions
- `make lint`: clean
- `go test -tags=integration -v .`: all 7 pass
- Coverage: `cmd` 18.7% → 22.0%; total statement coverage 62.4% → 64.7%
- **Not verified in this environment**: an actual cache-hit/cache-miss round-trip against real Bedrock (no AWS credentials available). The cache-point insertion, stripping, and question/document splitting logic are all unit-tested directly; the untested seam is whether a real model actually honors/rejects the `CachePointBlock` the way the retry logic assumes — worth confirming with real credentials before relying on this in production.

## Story 3.1 Acceptance Criteria Status
- [x] Cache point inserted after system prompt, when set
- [x] Cache point inserted between document and question, when a document is piped
- [x] Any error on a cache-enabled request triggers exactly one retry without cache points (both `chat` and `prompt`, all 3 request-building call sites)
- [x] No behavior change when there's nothing to cache (guarded by `hasSystemCachePoint`/`hasContentCachePoint` before ever attempting a retry)
