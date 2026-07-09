# Code Generation Plan: Unit 8, Automatic Tool-Use Enablement (#86) - Final Unit

## Unit Context
- **Stories**: 6.1-6.2 (FR1.1-FR1.4)
- **Design source**: `aidlc-docs/construction/unit-8-automatic-enablement/functional-design/business-logic-model.md`, `.../nfr-requirements/nfr-requirements-and-design.md`
- **Dependencies**: Units 6+7 (soft - registry now has the full 4-tool set and working gate to attach unconditionally)
- **Testing precedent** (verified, not assumed): `cmd/inferenceconfig_test.go` unit-tests `isDeprecatedSamplingParamsError` directly (pure function) but never `converseStreamWithFallbacks`/`converseWithFallbacks` themselves - they call the real, unmockable `*bedrockruntime.Client`. This unit follows the exact same pattern: `isToolUseUnsupportedError` gets direct unit tests; the new retry branch inside `converseStreamWithFallbacks` is verified via the full suite + manual check, not a new mocked-client test.

## Steps

- [x] **Step 1 - Failing tests: `isToolUseUnsupportedError`**
  Add `TestIsToolUseUnsupportedError` to `cmd/inferenceconfig_test.go`, mirroring `TestIsDeprecatedSamplingParamsError`'s table-driven style: nil error → false; unrelated error → false; a message containing "tool" + "not supported" → true; "tool" + "does not support" → true; "tool" + "unsupported" → true; a message containing "tool" but no matching qualifier → false; a message with a qualifier but no "tool" (e.g. a caching-rejection message) → false. Confirm compile failure.

- [x] **Step 2 - Implement `isToolUseUnsupportedError`**
  Add to `cmd/inferenceconfig.go` per `business-logic-model.md`. Run Step 1 tests, confirm green.

- [x] **Step 3 - Add the tool-use retry stage to `converseStreamWithFallbacks`**
  Insert the `input.ToolConfig != nil && isToolUseUnsupportedError(err)` branch between the existing cache-point and deprecated-sampling-params stages, per `business-logic-model.md`. No new unit test for this branch itself (see Unit Context's testing-precedent note) - verified via Step 5's full suite and manual check.

- [x] **Step 4 - Remove `--tools`**
  Delete the `rootCmd.PersistentFlags().Bool("tools", false, ...)` registration in `cmd/root.go`. Delete the `toolsEnabled` flag-read and the `if toolsEnabled { ... }` gate in `cmd/chat.go` - tool registration (all 4 tools, plus the gate construction already unconditional since Unit 6) becomes fully unconditional. Update `TestCLIFlagsExist` in `integration_test.go` if `--tools` is asserted there (check first - Initiative 2's flags were the ones added to that test, `--tools` predates it).

- [x] **Step 5 - Full verification**
  `make test`, `make lint`, `make test-coverage` (confirm no regression), `make cli && go test -tags=integration -v .`. Manual verification (same real-components approach as Unit 7, since live Bedrock isn't reachable here): confirm `chat-cli --help` no longer lists `--tools`; confirm a `chat` session's registry always contains all 4 tools regardless of any flag (inspect via a quick scratch driver, not the full TUI); this is the last unit of Initiative 3 - once green, proceed to the initiative-wide Build and Test stage covering all 3 units together.
