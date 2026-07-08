# Code Generation Plan: AGENTS.md Convention (#88)

## Unit Context
- **Stories/Requirements**: FR1-FR6, NFR1-NFR5 from `aidlc-docs/inception/requirements/agents-md-convention-requirements.md`
- **Design source**: `aidlc-docs/construction/agents-md-convention/functional-design/` + `nfr-requirements/nfr-requirements-and-design.md`
- **Dependencies**: Reuses `cmd/systemprompt.go` (`buildSystemContentBlocks`) and `cmd/promptcache.go` (`withSystemCachePoint`) as-is, no changes to either. Scope is `chat` only (`cmd/prompt.go` untouched).
- **Workspace root**: `/home/user/chat-cli` (per `aidlc-docs/aidlc-state.md`)
- **Project type**: Brownfield, Go module, existing `cmd/` package structure - all new code goes in `cmd/` alongside the Initiative 1 precedent files.

## Steps

- [x] **Step 1 - Failing tests: `resolveContextFilenames`**
  Create `cmd/projectcontext_test.go` with cases for BR11-BR13: unset config → default 3-name list; custom CSV → parsed/trimmed list; empty entries dropped (`"AGENTS.md,,CLAUDE.md"`); explicit `""` → empty list (disable case). Run `go test ./cmd/... -run TestResolveContextFilenames` and confirm it fails to compile (function doesn't exist yet).

- [x] **Step 2 - Implement `resolveContextFilenames`**
  Create `cmd/projectcontext.go` with `defaultContextFilenames` and `resolveContextFilenames(configValue string) []string`. Run the Step 1 tests, confirm green.

- [x] **Step 3 - Failing tests: `findProjectContextFile`**
  Add cases using `t.TempDir()` fixtures for BR1-BR4 + the Phase A/B algorithm: match in cwd; no match in cwd but match at a `.git`-boundary parent; no match anywhere; a directory named `AGENTS.md` is not treated as a match (BR2); cwd match takes precedence even when a boundary match also exists (BR4); no `.git` anywhere → only cwd checked. Confirm compile failure.

- [x] **Step 4 - Implement `findProjectContextFile`**
  Implement Phase A (upward `.git`-existence walk, capped at 64 levels per the NFR design note) and Phase B (check candidates at cwd, then the boundary dir if different) in `cmd/projectcontext.go`. Run Step 3 tests, confirm green.

- [x] **Step 5 - Failing tests: `loadProjectContext`**
  Add cases for BR7-BR10: normal read + trim; content over 32KB → truncated to exactly 32KB with `truncated=true`/correct `originalSize`; content under 32KB → `truncated=false`; unreadable file (e.g. no read permission, or a path to a directory) → returns an error. Confirm compile failure.

- [x] **Step 6 - Implement `loadProjectContext`**
  Implement read + trim + 32KB truncation in `cmd/projectcontext.go`. Run Step 5 tests, confirm green.

- [x] **Step 7 - Failing tests: `resolveAndLoadProjectContext`**
  Add composition-level cases: full pipeline finds and loads a match; BR8's empty-after-trim file is skipped and search continues to the next candidate; no candidates match anywhere → `found=false`; multiple candidates present → precedence order (BR3) respected end-to-end. Confirm compile failure.

- [x] **Step 8 - Implement `resolveAndLoadProjectContext`**
  Wire `resolveContextFilenames` (only needed by the caller, not this function - takes `candidates []string` directly per domain-entities.md) → `findProjectContextFile` → `loadProjectContext`, with the BR8 skip-and-continue loop. Run Step 7 tests, confirm green.

- [x] **Step 9 - `context-files` config key**
  In `cmd/config.go`: add `"context-files": true` to `supportedConfigKeys`; update the `Long` help text on `configSetCmd`/`configUnsetCmd` and the error-message key lists; add `"context-files"` to `configListCmd`'s `configKeys` slice. Extend `cmd/config_test.go` if it asserts against the supported-keys set (check first), otherwise add a minimal case confirming `context-files` is accepted by `config set`.

- [x] **Step 10 - `--no-context-file` flag**
  Register a persistent bool flag `--no-context-file` (default `false`) in `cmd/root.go`'s `init()`, alongside the existing `--tools`/`--thinking` registrations. Update `TestCLIFlagsExist` in `integration_test.go` to assert the new flag is present (grep-style check, matching how the other Initiative-1 flags were added to that test).

- [x] **Step 11 - Wire into `cmd/chat.go`**
  Immediately after the existing `systemPrompt := fm.GetConfigValue("system-prompt", systemFlag, "").(string)` line: read the `--no-context-file` flag and the `context-files` config value; if `systemPrompt == ""` and the flag is false and `resolveContextFilenames(...)` yields a non-empty list, call `resolveAndLoadProjectContext(cwd, candidates)` (cwd via `os.Getwd()`). On a truncation, print the BR9 warning to stderr. On a found match, assign `systemPrompt` to the loaded content and print the FR6/BR15/BR16 notice to stdout using the existing `\033[90m...\033[0m` gray styling convention already used elsewhere in this file.

- [x] **Step 12 - Unit tests for the `chat.go` call site**
  Since `chat.go`'s `RunE` isn't unit-tested directly today (established precedent - CLI wiring is covered by integration tests, not unit tests), verify this step by re-running the full suite plus a manual smoke test in Step 14, rather than adding a new brittle unit test around Cobra command execution.

- [x] **Step 13 - Documentation**
  Update `README.md` and `docs/usage.md` with a new "Project Context" section: what `AGENTS.md`/`CLAUDE.md`/`.github/copilot-instructions.md` do, the precedence rule, `context-files` config, `--no-context-file`, and the chat-only scope.

- [x] **Step 14 - Full verification**
  `make test`, `make lint`, `make test-coverage` (confirm no regression), `make cli && go test -tags=integration -v .`. Manual smoke test: run `chat-cli` from a directory containing an `AGENTS.md`, confirm the notice prints and the system prompt is used; confirm `--system` suppresses discovery; confirm `--no-context-file` suppresses discovery.
