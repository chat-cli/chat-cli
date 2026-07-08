# Unit 1 — System Prompt Support — Code Generation Summary

**Issue**: [#81](https://github.com/chat-cli/chat-cli/issues/81) | **Story**: 1.1 | **Plan**: `aidlc-docs/construction/plans/unit-1-system-prompt-code-generation-plan.md`

## Files Modified
- `cmd/config.go` — extracted the 3 duplicated `supportedKeys` map literals into one package-level `supportedConfigKeys` map; added `"system-prompt"`; updated `Long` help text and `config list`'s `configKeys` slice
- `cmd/root.go` — registered `--system` persistent flag (default `""`)
- `cmd/prompt.go` — registered `--system` persistent flag; resolved system prompt (flag → config → none) and attached to both the `--no-stream` (`ConverseInput`) and streaming (`ConverseStreamInput`) request paths
- `cmd/chat.go` — resolved system prompt the same way and attached it to `ConverseStreamInput.System` once, before the session loop starts
- `cmd/cmd_test.go` — added `TestConfigCommandSupportsSystemPrompt`; extended `TestRootCommand` and `TestPromptCommand` with `--system` flag assertions
- `README.md`, `docs/usage.md` — documented the new flag, config key, and precedence behavior

## Files Created
- `cmd/systemprompt.go` — `buildSystemContentBlocks(systemPrompt string) []types.SystemContentBlock`, a pure, unit-testable helper
- `cmd/systemprompt_test.go` — `TestBuildSystemContentBlocks` (table-driven, covers empty and non-empty input)

## Verification
- `make test`: all packages pass, no regressions (existing `cmd`, `config`, `repository`, `utils` suites all green)
- `make lint` (`go vet` + `go fmt`): clean
- `go test -tags=integration -v .`: all 7 integration tests pass, including `TestCLIFlagsExist`
- `go test -coverprofile`: `cmd` package coverage 7.4% → 8.0% (no regression, per NFR2); `buildSystemContentBlocks` at 100%
- Manual smoke test: `--system` appears in `--help` for both root and `prompt`; `config set/list/unset system-prompt` all behave correctly
- `golangci-lint` (the stricter ruleset from `.golangci.yml`) could not run — the installed binary version doesn't match this repo's config format. Pre-existing environment issue, unrelated to this change; reinforces the value of issue #96 (add CI workflow)

## Story 1.1 Acceptance Criteria Status
- [x] No flag/config set → behavior unchanged (verified via `buildSystemContentBlocks("")` returning `nil`, and existing tests/integration tests still passing unmodified)
- [x] `--system` flag → `SystemContentBlocks` sent with that text
- [x] `system-prompt` config value → used when no flag is passed
- [x] Flag takes precedence over config (reuses `FileManager.GetConfigValue`'s existing, already-tested precedence logic)
