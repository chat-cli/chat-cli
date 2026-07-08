# Unit 1 — System Prompt Support — Code Generation Plan

**Unit**: Unit 1 (System Prompt Support), issue [#81](https://github.com/chat-cli/chat-cli/issues/81)
**Stories implemented**: Story 1.1 (`aidlc-docs/inception/user-stories/stories.md`)
**FR/NFR coverage**: FR1.1-FR1.4, NFR1 (backward compatibility), NFR2 (TDD/coverage), NFR6 (docs)
**Dependencies on other units**: None (Unit 1 has no dependencies)
**Workspace root**: `/home/user/chat-cli` (from `aidlc-state.md`) — application code goes here, never in `aidlc-docs/`

## Context From Application Design / Reverse Engineering
- No new components for this unit (`application-design/components.md`) — direct additions to existing `cmd/root.go`, `cmd/prompt.go`, `cmd/chat.go`, `cmd/config.go`.
- `config.FileManager.GetConfigValue` (`config/config.go`) is already generic over string keys and already has full precedence test coverage (`TestGetConfigValue` in `config/config_test.go`) — no changes or new tests needed there; `"system-prompt"` is just a new key name used the same way `"model-id"` is used today.
- Verified against the installed AWS SDK v2 `bedrockruntime` module (`v1.23.0`): `ConverseInput.System` and `ConverseStreamInput.System` are both `[]types.SystemContentBlock`; `types.SystemContentBlockMemberText{Value: string}` implements that interface. This confirms FR1.3 is directly achievable with existing SDK types already in `go.mod` (NFR5 — no new dependency).
- `cmd/cmd_test.go` today only asserts command/flag *metadata* (no AWS calls are mocked) — `Run()` bodies aren't unit-tested directly. This unit follows that existing pattern: flag/config wiring is tested directly; the request-building logic is extracted into a small pure helper function so it's unit-testable without needing to mock the AWS SDK.

## TDD Order (per CLAUDE.md: tests before implementation)

### Step 1 — Failing test: config supports `system-prompt` key
- [x] Add test to `cmd/cmd_test.go`: `TestConfigCommandSupportsSystemPrompt` — asserts `"system-prompt"` is accepted by set/unset/list (by checking against the same supported-keys pattern `TestConfigCommand` already uses, extended with a table covering the three subcommands). Run `go test ./cmd/... -run TestConfigCommandSupportsSystemPrompt` and confirm it fails (key not yet supported).

### Step 2 — Implementation: add `system-prompt` to config command
- [x] Modify `cmd/config.go`: add `"system-prompt": true`. **Refinement during generation**: extracted the 3 duplicated local `supportedKeys` map literals into one package-level `supportedConfigKeys` var (reduces duplication, matches the spirit of #92, and makes the key list directly unit-testable). Added `"system-prompt"` to it and to `configListCmd`'s `configKeys` slice; updated both `Long` help strings.
- [x] Re-run Step 1's test — confirm it now passes.

### Step 3 — Failing test: system-content-block builder helper
- [x] Add test file `cmd/systemprompt_test.go` (new, small, focused): `TestBuildSystemContentBlocks` table-driven test covering:
  - empty string in → `nil` (no blocks) — proves NFR1 (unchanged behavior when unset)
  - non-empty string in → single `[]types.SystemContentBlock{&types.SystemContentBlockMemberText{Value: "..."}}`
- [x] Run and confirm it fails (function doesn't exist yet).

### Step 4 — Implementation: extract testable helper
- [x] Add `cmd/systemprompt.go` (new file): `buildSystemContentBlocks(systemPrompt string) []types.SystemContentBlock` — pure function, no AWS calls, no I/O. Returns `nil` for an empty string, else a single `SystemContentBlockMemberText`.
- [x] Re-run Step 3's test — confirm it passes.

### Step 5 — Failing test: `--system` flag exists on root and prompt commands
- [x] Extend `TestRootCommand` (or add `TestSystemPromptFlag`) in `cmd/cmd_test.go` asserting `rootCmd.PersistentFlags().Lookup("system")` exists with an empty-string default.
- [x] Add equivalent assertion for `promptCmd.PersistentFlags().Lookup("system")`.
- [x] Run and confirm failure (flags don't exist yet).

### Step 6 — Implementation: register `--system` flag
- [x] Modify `cmd/root.go`'s `init()`: add `rootCmd.PersistentFlags().String("system", "", "set a system prompt")`, alongside the existing `--model-id`/`--custom-arn`/etc. registrations (same pattern, so it works at both `chat-cli` and `chat-cli chat`).
- [x] Modify `cmd/prompt.go`'s `init()`: add `promptCmd.PersistentFlags().String("system", "", "set a system prompt")`, alongside its own existing `--model-id`/`--custom-arn` duplication.
- [x] Re-run Step 5's tests — confirm they pass.
- [x] Re-run `TestFlagInheritance`-style check is not required for `system` since (like `chat-id`/`temperature`) it doesn't need to appear in that specific test's list unless the user wants full parity — out of scope here, existing test only enumerates `region`/`model-id`/`custom-arn`.

### Step 7 — Implementation: wire system prompt into `chat` (`cmd/chat.go`)
- [x] Read `--system` flag via the existing `flagCmd.PersistentFlags().GetString(...)` pattern already used for `model-id`/`custom-arn`/`chat-id`/etc.
- [x] Resolve via `fm.GetConfigValue("system-prompt", systemFlag, "")` — same precedence call already used for `model-id`.
- [x] Call `buildSystemContentBlocks(...)`; if non-nil, set `converseStreamInput.System = ...` before the tty-loop starts (system prompt is fixed for the session, per Assumption 1 in `requirements.md`).
- [x] No test added in this step — covered by Step 4's unit test of the pure helper, plus Step 9's manual verification (per NFR1/NFR2, `Run()` itself follows the codebase's existing untested-at-the-Run-level pattern).

### Step 8 — Implementation: wire system prompt into `prompt` (`cmd/prompt.go`)
- [x] Same flag read + `GetConfigValue` resolution as Step 7, using `cmd.PersistentFlags()` (prompt's own flag set, matching its existing `model-id`/`custom-arn` read pattern).
- [x] Call `buildSystemContentBlocks(...)`; if non-nil, set `.System = ...` on **both** `converseInput` (the `--no-stream` path) and `converseStreamInput` (the default streaming path) — both code paths exist in `promptCmd.Run` today and both need the system prompt attached (FR1.3).

### Step 9 — Full test suite and lint
- [x] Run `make test` — confirm all existing tests plus the new ones pass, no regressions.
- [x] Run `make lint` — confirm clean.
- [x] Run `make test-coverage` — confirm `cmd` package coverage does not regress (NFR2); expect a small improvement given 2 new test functions.

### Step 10 — Documentation
- [x] Update `README.md`: add a "System Prompt" section (near the existing "Configuration"/"Prompt"/"Chat" sections) documenting `--system`, `chat-cli config set system-prompt "..."`, and precedence (flag → config → none), mirroring the existing model-id/custom-arn documentation style.
- [x] Update `docs/usage.md` with the same information for the Sphinx docs site (per `CLAUDE.md`'s documentation rules — edit existing docs, no new root-level `.md` files).

### Step 11 — Unit Documentation Summary
- [x] Create `aidlc-docs/construction/unit-1-system-prompt/code/summary.md` (markdown summary only, per Code Location Rules) listing modified/created files and confirming story 1.1's acceptance criteria are met.

## Story Traceability
- Story 1.1 (all 4 acceptance criteria) → Steps 1-8 implement and test them directly; Step 9 verifies the full suite; Step 10 satisfies NFR6.
