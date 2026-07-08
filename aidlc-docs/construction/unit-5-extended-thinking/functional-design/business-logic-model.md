# Functional Design — Unit 5 (Extended Thinking / Reasoning Mode)

## SDK Facts Verified by Direct Inspection (v1.55.0)
- `ConverseInput`/`ConverseStreamInput.AdditionalModelRequestFields` is `document.Interface` — an **untyped, free-form** field. Unlike every other unit so far, there is no dedicated `ReasoningConfig` struct to inspect — the exact JSON shape for "enable extended thinking" is provider-specific and passed through opaquely.
- Response-side types **are** strongly typed and confirmed: `types.ContentBlockMemberReasoningContent{Value: ReasoningContentBlock}`; `ReasoningContentBlockMemberReasoningText{Value: ReasoningTextBlock{Text *string, Signature *string}}`; streaming deltas via `ReasoningContentBlockDeltaMemberText{Value string}`, `ReasoningContentBlockDeltaMemberSignature{Value string}`, and `ReasoningContentBlockDeltaMemberRedactedContent{Value []byte}`.
- **The SDK's own doc comment is explicit**: "If you pass a reasoning block back to the API in a multi-turn conversation, include the text and its signature unmodified." This means `chat`'s conversation history must preserve the reasoning block (text + signature) across turns, not just print it and discard it — the same category of requirement Unit 2's tool-use blocks had.

## ⚠️ Known Limitation: Request-Side Shape Is Unverified
Because `AdditionalModelRequestFields` is untyped, its correct shape for enabling reasoning can't be confirmed by static SDK inspection the way every other unit's request/response shapes were. Based on training knowledge (not verified against a live source in this environment), the design assumes Anthropic-family models on Bedrock accept:
```json
{"reasoning_config": {"type": "enabled", "budget_tokens": <n>}}
```
This is the single highest-risk assumption in the whole initiative — if wrong, `--thinking` will reliably fail with a Bedrock validation error rather than silently misbehaving (consistent with FR5.4's "surface the API's error clearly" requirement, which already anticipated needing to handle this gracefully). Flagging this explicitly rather than asserting it with unwarranted confidence.

## Design Decision 1: `--thinking-budget` Flag, Not Just `--thinking`
Extended thinking needs a token budget (`budget_tokens`), and that budget must leave room under `--max-tokens` for the actual answer — Bedrock is expected to reject a request where the budget doesn't fit. **Decision**: add `--thinking-budget <int>` (default `1024`) alongside `--thinking`, and document the `--max-tokens` interaction clearly rather than silently auto-adjusting `--max-tokens` (which could surprise users in a different way). If the combination doesn't fit, FR5.4 already covers surfacing that error clearly.

## Design Decision 2: Extend `accumulateStream`, Don't Duplicate It
`chat.go`'s `accumulateStream` (Unit 2) already tracks content blocks by index and handles text + tool-use. Reasoning content is a third block kind streamed the same way (start/delta/stop by index) — extending the existing `blockAccumulator`/`accumulateStream` to also handle `ReasoningContentBlockDeltaMemberText`/`...MemberSignature` is far less risky than writing a second, parallel accumulation function. The finalized reasoning block (text + signature) is included in the assistant `Message` appended to conversation history, satisfying the SDK's multi-turn preservation requirement automatically (since that message already flows through the existing history-append logic).

## Design Decision 3: `prompt` Renders Reasoning But Doesn't Need to Preserve It
`prompt` is one-shot — there's no next turn to preserve a signature for. It only needs to detect `ContentBlockMemberReasoningContent` in the response and print it distinctly before the final answer text. No `accumulateStream`-style indexed tracking is needed there since `prompt`'s response handling is comparatively simple already (see `Converse`'s direct `response.Value.Content[0]` and `ProcessStreamingOutput`'s single-block-focused loop).

## Business Rules
- **Rule 1 (FR5.1)**: `--thinking` (bool) + `--thinking-budget` (int, default 1024) on both `chat` and `prompt` set `AdditionalModelRequestFields` per the shape above. Without `--thinking`, no field is set — behavior unchanged (NFR1).
- **Rule 2 (FR5.2)**: Reasoning text is printed visually distinct from the final answer — dimmed, using the same ANSI convention (`\033[90m...\033[0m`) already used for the user-input echo in `chat.go`, prefixed with a label (e.g. `[thinking] `).
- **Rule 3 (FR5.2, chat only)**: The finalized reasoning block (text + signature) is preserved in the assistant message appended to conversation history, per the SDK's multi-turn requirement (Design Decision 2).
- **Rule 4 (FR5.4)**: Any error from a `--thinking`-enabled request surfaces via the existing `log.Fatalf` pattern, unchanged — no retry (unlike caching), since silently disabling reasoning would change what the user asked for.
- **Rule 5 (Redacted content)**: `ReasoningContentBlockDeltaMemberRedactedContent`/`ReasoningContentBlockMemberRedactedContent` (provider-side safety redaction) is preserved in history (Design Decision 2 covers this automatically via the same accumulator) but not printed as visible text, since it's encrypted bytes, not human-readable reasoning.
