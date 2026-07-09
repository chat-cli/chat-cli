# Functional Design (Combined with NFR): Unit 8, Automatic Tool-Use Enablement (#86)

Lower novelty than Units 6-7 - this unit extends an already-established, already-tested pattern (`cmd/inferenceconfig.go`'s cascading fallback retries) rather than introducing anything new. Verified against the actual current source, not assumed.

## The Existing Pattern (verified)
`converseStreamWithFallbacks` (`cmd/inferenceconfig.go:95`) already does exactly this shape of thing twice: try the request, and on failure, check if a specific optional field was present (`hasSystemCachePoint`/`hasContentCachePoint`, or a deprecated-sampling-params error match), strip it, retry once, log a `log.Printf` notice either way. This is the *only* function on chat's send path (chat always streams; `converseWithFallbacks`, the non-streaming sibling, is `prompt`-only and is **not** touched by this unit, since tool use has never applied to `prompt`).

## Design: Add a Third Fallback Stage
Insert a tool-use-rejection check between the existing cache-point and deprecated-sampling-params stages:
```go
if input.ToolConfig != nil && isToolUseUnsupportedError(err) {
    log.Printf("tool use not supported for this model, retrying without tools: %v", err)
    input.ToolConfig = nil
    output, err = svc.ConverseStream(ctx, input)
    if err == nil { return output, nil }
}
```
- The `input.ToolConfig != nil` guard mirrors `hasSystemCachePoint(...)`'s role - never attempt a pointless retry when there's nothing to strip.
- **"Disabled for the rest of the session" (FR1.2) falls out for free**: `converseStreamInput` is built once per `chat` session and mutated across turns (the existing `for` loop appends to `.Messages` on the same struct rather than rebuilding it). Setting `input.ToolConfig = nil` here persists across every subsequent turn automatically - no new session-state tracking needed, the same mechanism the cache/sampling-params fallbacks already rely on.
- **The FR1.3 user-visible notice is the existing `log.Printf` call itself** - it already goes to stderr and is how every other fallback in this function already surfaces itself. Reusing this convention satisfies "a clear, one-line notice" without inventing new UI styling for one code path.

## `isToolUseUnsupportedError` - Flagged as an Unverified Assumption
**This is the highest-risk assumption in this unit**, the same category of uncertainty Initiative 1's Unit 5 had for `reasoning_config`'s shape: there is no way to statically verify Bedrock's exact error text for "this model/request doesn't support tool use" without live credentials. Designed as a heuristic, matching `isDeprecatedSamplingParamsError`'s existing style:
```go
func isToolUseUnsupportedError(err error) bool {
    if err == nil { return false }
    msg := strings.ToLower(err.Error())
    if !strings.Contains(msg, "tool") { return false }
    return strings.Contains(msg, "not supported") || strings.Contains(msg, "does not support") || strings.Contains(msg, "unsupported")
}
```
Flagged prominently for the real-credential verification list (same list Unit 5's `reasoning_config` shape is already on).

## `--tools` Removal
- `cmd/root.go`: the `rootCmd.PersistentFlags().Bool("tools", false, ...)` registration is deleted.
- `cmd/chat.go`: the `toolsEnabled` flag-read and the `if toolsEnabled { ... }` gate around tool registration are both removed - registration becomes unconditional (all 4 tools always registered).
