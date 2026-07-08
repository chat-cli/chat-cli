# Functional Design: Domain Entities - AGENTS.md Convention (#88)

This feature has no persistent data model (no DB schema change, no new persisted entity) - "entities" here means the new pure-function surface and its inputs/outputs, following the same style as Initiative 1's `cmd/systemprompt.go`/`cmd/promptcache.go`.

## New File: `cmd/projectcontext.go`

Mirrors the existing single-purpose file-per-feature pattern (`systemprompt.go`, `promptcache.go`, `documentinput.go`, `reasoning.go`).

### `defaultContextFilenames`
A package-level `[]string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"}` - the BR5 default precedence list, used when the `context-files` config key is unset.

### `resolveContextFilenames(configValue string) []string`
Implements BR11-BR13: parses the comma-separated config value (trim, drop empties); returns `defaultContextFilenames` if the config value is empty/unset, otherwise the parsed list (which may itself be empty, per BR12 - the disable case).

### `findProjectContextFile(cwd string, candidates []string) (path string, ok bool)`
Implements the Phase A boundary walk + Phase B check from business-logic-model.md and BR1-BR4. Pure with respect to its inputs (`cwd`, `candidates`) but does touch the filesystem (`os.Stat`) - this is the one function that needs a temp-directory fixture in tests rather than being a fully pure computation, same testing shape as `utils.ValidateLocalPath` already has.

### `loadProjectContext(path string) (content string, truncated bool, originalSize int, err error)`
Implements BR7-BR10: reads the file, trims whitespace, truncates at 32KB if needed. Returns enough information for the caller to print the FR6 notice and the BR9 truncation warning without re-deriving anything. A read error (BR10) is returned as `err` for the caller to treat as "no match" - this function does not itself decide search-continuation policy, that's the caller's (`resolveAndLoadProjectContext`, below) job.

### `resolveAndLoadProjectContext(cwd string, candidates []string) (content string, sourcePath string, truncated bool, found bool)`
Composition root tying the above three together, plus BR8's "empty-after-trim counts as no match, keep searching" loop. This is the one function `cmd/chat.go` actually calls.

## `cmd/chat.go` Changes
- New flag: `--no-context-file` (bool, default `false`) - registered as a persistent flag alongside the existing `--tools`/`--thinking` flags in `cmd/root.go`'s `init()` (same file that already registers all other chat/prompt-shared flags).
- New call site immediately after the existing `systemPrompt := fm.GetConfigValue(...)` line (chat.go:114), gated per business-logic-model.md's "Entry Point and Call Site" conditions.
- New notice print (FR6/BR15/BR16), styled `\033[90m...\033[0m` matching the existing gray convention.

## `cmd/config.go` Changes
- `supportedConfigKeys["context-files"] = true`
- `configSetCmd`/`configUnsetCmd`'s `Long` help text and error-message key lists updated to mention `context-files` (matching the existing pattern where these are hand-maintained alongside the map - not refactored into a single source of truth here, out of scope for this feature).
- `configListCmd`'s hardcoded `configKeys` slice gains `"context-files"`.

## No Changes Needed
- `cmd/systemprompt.go`, `cmd/promptcache.go` - both consume "the resolved system prompt string" with no awareness of its source; reused completely as-is (FR3.1, FR3.3).
- `cmd/prompt.go` - out of scope, chat-only (decision 1).
- `db`/`repository` packages - no persistence changes.
