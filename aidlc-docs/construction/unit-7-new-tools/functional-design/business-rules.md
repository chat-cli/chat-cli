# Functional Design: Business Rules - Unit 7, New Built-in Tools (#86)

## `write_file`
- **BR1**: Path confinement uses `utils.ValidateLocalPathForWrite` (new, confinement-only) - identical traversal protection to `read_file`/`--document`/`--image`, minus the existence requirement.
- **BR2**: Both creating a new file and overwriting an existing one are in scope (FR2.2) - no separate "create" vs "overwrite" tool or flag.
- **BR3**: `ConfirmationSummary` and `Execute` each independently call `ValidateLocalPathForWrite` - neither trusts the other's validation, consistent with Unit 6's stated design principle (business-logic-model.md: "keeping the two methods independent means a summary-generation bug can never accidentally skip validation Execute would have caught").
- **BR4**: The confirmation summary shows full content up to 4KB; past that, a truncated preview plus an explicit `"(N bytes total, shown truncated)"` note (NFR5/Unit 6 NFR's usability decision, carried forward).
- **BR5**: Pattern key is the resolved path's containing directory, relative to the git repo root if inside one (`utils.FindGitBoundary`), else relative to cwd (Unit 6 BR2).

## `run_shell`
- **BR6**: Runs via `sh -c "<command>"`, no allowlist/denylist (Requirements Analysis Assumption 2) - the confirmation gate (Unit 6) is the control.
- **BR7**: Working directory is fixed to `chat-cli`'s own cwd (`os.Getwd()`), not configurable by the model.
- **BR8**: A 30-second timeout applies to every invocation. Exceeding it returns a Go error (`"command timed out after 30s"`), distinct from a normal non-zero exit.
- **BR9**: Combined stdout+stderr is truncated at 32KB with a `"(output truncated)"` marker appended, mirroring #88's `maxContextFileSize` precedent for consistency across the codebase.
- **BR10**: A non-zero exit code is **not** a Go error - it's reported in the tool's success-path output text as `"[exit code: N]"` followed by the (possibly truncated) combined output, so the model can see both the failure and why.
- **BR11**: Pattern key is the command's first whitespace-separated token; an empty/whitespace-only command yields an empty pattern key (never matches, always prompts - same fail-safe shape as Unit 6 BR1's `run_shell` note).

## `git_diff`
- **BR12**: No confirmation gate - read-only (FR4.2).
- **BR13**: An optional `arg` (path or ref) is passed straight through to `git diff [arg]`; omitted/empty `arg` runs a plain `git diff`.
- **BR14**: `git diff`'s own error output (non-git-repo, invalid ref, etc.) becomes the tool's returned error text verbatim - no special-casing of "not a repository" versus other git errors, since git's own message is already clear (FR4.3).
- **BR15**: Same 30-second timeout as `run_shell`, for consistency, though not expected to matter in practice for a diff operation.

## Shared
- **BR16**: All 3 tools' JSON input is unmarshaled independently in each method that needs it (`ConfirmationSummary`, `Execute`) - no shared parsed-input cache between the two calls, consistent with BR3's independence principle.
