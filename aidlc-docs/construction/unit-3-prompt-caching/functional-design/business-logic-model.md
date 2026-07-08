# Functional Design — Unit 3 (Prompt Caching)

**Note**: Business logic model and business rules are presented in one combined document for this unit (not split into separate `business-logic-model.md`/`business-rules.md`/`domain-entities.md` files as Unit 2's did) — the actual new logic here is small (a cache-point content block plus a retry policy), so a single lightweight document is proportionate, consistent with how NFR Requirements+Design were combined for Unit 2.

## SDK Facts Verified by Direct Inspection (post-upgrade, v1.55.0)
- `types.CachePointBlock{Type: types.CachePointType, Ttl: types.CacheTTL}` — `CachePointType` has exactly one value today, `CachePointTypeDefault = "default"`.
- `types.SystemContentBlockMemberCachePoint{Value: CachePointBlock}` — goes in the `System []types.SystemContentBlock` slice, after the system-prompt text block (FR3.1).
- `types.ContentBlockMemberCachePoint{Value: CachePointBlock}` — goes in a message's `Content []types.ContentBlock` slice (FR3.2).
- **No dedicated exception type exists for "this model/request doesn't support cache points."** The full error type list (`types/errors.go`) has `ValidationException`, `AccessDeniedException`, `ModelErrorException`, etc. — nothing cache-specific. Confirmed by direct inspection, not assumed.

## Design Decision 1: Retry Policy Doesn't Try to Distinguish the Error Cause

Since there's no specific "unsupported cache point" exception to catch, attempting to string-match `ValidationException.Message` for cache-related wording would be brittle (breaks silently if AWS changes wording) and isn't worth the complexity for what's fundamentally a best-effort optimization.

**Decision**: On **any** error from a cache-point-enabled request, retry the same request **once** with cache points stripped out, and log a non-fatal warning ("prompt caching not supported for this request, retrying without it"). If the retry also errors, surface that second error normally (existing `log.Fatal` behavior, unchanged) — if the real problem wasn't caching, the retry fails for the same underlying reason and the user sees a normal, honest error rather than a caching-specific one that might be misleading.

## Design Decision 2: `prompt`'s Document + Question Must Become Separate Content Blocks

Today, `cmd/prompt.go` does `prompt := args[0]; document, _ := utils.LoadDocument(); prompt += document` — the piped document and the user's question are concatenated into **one** string before ever becoming a `ContentBlockMemberText`. For caching to be meaningful (cache the large, stable document; don't cache the small, always-different question), they must become **two separate content blocks** with a cache point in between: `[document text block, cache point, question text block]`.

**Decision**: Restructure `prompt.go`'s message-building so that when a piped document is present, the message content is `[]types.ContentBlock{documentTextBlock, cachePointBlock, questionTextBlock}` instead of one merged block. When no document is piped, behavior is unchanged (one text block, no cache point — nothing to cache) — this preserves NFR1 for the common case.

## Business Rules
- **Rule 1 (FR3.1)**: A cache point is appended to `System` immediately after the system-prompt text block, only when a system prompt is set (Unit 1). No system prompt → no cache point → `System` field unchanged from Unit 1's behavior.
- **Rule 2 (FR3.2)**: A cache point is inserted between the document block and the question block, only when a document was piped via stdin. No piped document → single text block, unchanged from today.
- **Rule 3 (FR3.3, Design Decision 1)**: Any error on a cache-point-enabled request triggers exactly one retry without cache points; a second failure surfaces normally.
- **Rule 4 (FR3.4)**: Cache token metrics, when present in `ConverseOutput`/`ConverseStreamOutput`'s usage metadata, are not rendered in this pass (deferred to a future token-display feature, issue #41) — this unit only needs to not break if they're absent.
- **Rule 5 (scope note)**: This unit applies to **both** `chat` and `prompt` for system-prompt caching (Rule 1), but **only `prompt`** for document caching (Rule 2) — `chat` has no document-input capability yet (tracked separately, issue #46), so there's nothing to cache there today.
