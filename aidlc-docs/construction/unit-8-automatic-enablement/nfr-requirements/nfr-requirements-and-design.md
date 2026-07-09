# NFR Requirements and Design (Combined): Unit 8, Automatic Tool-Use Enablement (#86)

## Reliability

### Requirement
Removing the `--tools` opt-in must not make `chat` break for any model that doesn't support tool use - this is the entire point of FR1.2's retry-without-tools fallback, and it's the most important property in this unit.

### Design
- The fallback reuses `converseStreamWithFallbacks`'s already-tested cascading-retry shape - same function, same error-handling conventions, same "strip and retry once" policy already proven correct for cache points and sampling params.
- `isToolUseUnsupportedError`'s exact string-matching heuristic is unverified against real Bedrock error text (flagged in the functional design) - if it under-matches (misses a real rejection message), the practical failure mode is "the retry doesn't trigger and the original error surfaces to the user," not a crash or hang. This is a *graceful* degradation of the *fallback itself*, not a new failure mode - acceptable risk, consistent with how `isDeprecatedSamplingParamsError` shipped with the same category of uncertainty.

### Compliance
✅ Compliant, with a flagged unverified assumption (heuristic error matching) carried to the real-credential verification list, not silently accepted as certain.

## Backward Compatibility (revised, per Initiative 3's Requirements Analysis)
This unit is the one that actually executes the intentional default-behavior change agreed in Requirements Analysis (Assumption 7): `chat` now always attempts tool use, where it previously required `--tools`. This is not a regression to guard against - it's the deliberate deliverable of this unit, already explicitly approved.

## Non-Applicable Categories
- **Security**: N/A for this unit specifically - the security-relevant work (confinement, the confirmation gate) is Units 6/7's job; this unit is pure wiring/retry-logic.
- **Scalability/Performance/Availability/Usability**: N/A, same rationale as every prior unit.
