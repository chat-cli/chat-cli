# Code Generation Plan: Unit 7, New Built-in Tools (#86)

## Unit Context
- **Stories**: 7.1-7.3 (FR2.1-FR2.3, FR3.1-FR3.2, FR4.1-FR4.3)
- **Design source**: `aidlc-docs/construction/unit-7-new-tools/functional-design/`, `.../nfr-requirements/nfr-requirements-and-design.md`
- **Dependencies**: Unit 6 (HARD - `RequiresConfirmation`/`ConfirmationSummary`/`PermissionGate` must exist, which they do, per Unit 6's approved and merged code)
- **Testability note**: `RunShellTool`'s 30s timeout would make a real timeout test slow. Since tests live in the same `tools` package, `RunShellTool`'s timeout is an unexported struct field constructible directly in tests (`&RunShellTool{timeout: 50 * time.Millisecond}`), not a hardcoded constant baked into `Execute`.

## Steps

- [ ] **Step 1 - Failing tests: `utils.ValidateLocalPathForWrite`**
  Add cases to `utils/utils_test.go`: a path to a file that doesn't exist yet succeeds (unlike `ValidateLocalPath`); a path traversal attempt (`../../etc/passwd`) is still rejected; re-run the full existing `TestValidateLocalPath` suite unmodified as a regression check that the refactor didn't change `ValidateLocalPath`'s behavior. Confirm compile failure.

- [ ] **Step 2 - Implement `confineToWorkingDir` + `ValidateLocalPathForWrite`**
  Extract the confinement logic from `ValidateLocalPath` into a private `confineToWorkingDir(filename string) (string, error)`; refactor `ValidateLocalPath` to call it plus its existing existence check; add `ValidateLocalPathForWrite` calling only the confinement step. Run Step 1 tests plus full `utils` package, confirm all green.

- [ ] **Step 3 - Failing tests: `WriteFileTool`**
  Create `tools/writefile_test.go`: `RequiresConfirmation() == true`; `ConfirmationSummary` returns the expected summary text and a directory-based pattern key (BR5) for both an in-repo and a no-repo cwd; content over 4KB is truncated in the summary with the BR4 note; `Execute` creates a new file; `Execute` overwrites an existing file; `Execute` rejects a path outside the working directory. Confirm compile failure.

- [ ] **Step 4 - Implement `WriteFileTool`**
  Create `tools/writefile.go` per `domain-entities.md`. Run Step 3 tests, confirm green.

- [ ] **Step 5 - Failing tests: `RunShellTool`**
  Create `tools/runshell_test.go`: `RequiresConfirmation() == true`; `ConfirmationSummary` returns `"Run: <command>"` and the first-token pattern key (BR11), including the empty-command edge case; `Execute` returns command output on success; `Execute` embeds `[exit code: N]` for a non-zero exit **without** returning a Go error (BR10); `Execute` returns a Go error on timeout (using the injectable short timeout field, not the real 30s default); output over 32KB is truncated with the BR9 marker. Confirm compile failure.

- [ ] **Step 6 - Implement `RunShellTool`**
  Create `tools/runshell.go` per `domain-entities.md`, with `timeout time.Duration` as an unexported field (default 30s via `NewRunShellTool`). Run Step 5 tests, confirm green.

- [ ] **Step 7 - Failing tests: `GitDiffTool`**
  Create `tools/gitdiff_test.go`: `RequiresConfirmation() == false`; `Execute` with no `arg` runs a plain `git diff` in a temp git repo fixture; `Execute` with an `arg` passes it through; `Execute` outside a git repository returns an error containing git's own message (BR14). Confirm compile failure.

- [ ] **Step 8 - Implement `GitDiffTool`**
  Create `tools/gitdiff.go` per `domain-entities.md`. Run Step 7 tests, confirm green.

- [ ] **Step 9 - Register the 3 new tools in `cmd/chat.go`**
  Add `registry.Register(tools.NewWriteFileTool())`, `registry.Register(tools.NewRunShellTool())`, `registry.Register(tools.NewGitDiffTool())` alongside the existing `read_file` registration, still inside the `if toolsEnabled` block (Unit 8 removes that gate, not this unit). No new unit test needed beyond re-running the full `cmd` suite (registration is a one-line wiring change, same category as Initiative 1's established precedent for untested `chat.go` wiring).

- [ ] **Step 10 - Full verification**
  `make test`, `make lint`, `make test-coverage` (confirm no regression), `make cli && go test -tags=integration -v .`. Manual smoke test against the compiled binary with `--tools` set: trigger `write_file` and confirm the prompt shows path+content and the once/session/always/deny choices work as expected; trigger `run_shell` similarly; trigger `git_diff` and confirm it runs with **no** prompt (read-only). This is the first point in the initiative where the confirmation gate actually fires - Unit 6 alone couldn't demonstrate this end-to-end.
