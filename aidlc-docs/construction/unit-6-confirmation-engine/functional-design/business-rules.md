# Functional Design: Business Rules - Unit 6, Confirmation and Sticky Approval Engine (#86)

## Pattern Key Derivation
- **BR1**: `run_shell`'s pattern key is the command string's first whitespace-separated token (e.g. `"git diff main"` → `"git"`). Leading/trailing whitespace is trimmed before splitting. An empty command string (shouldn't happen given the tool's own input validation, but defensively) yields an empty pattern key, which never matches any prior approval and always prompts.
- **BR2**: `write_file`'s pattern key is the containing directory of the resolved (validated) path, **relative to the repo root** determined by `utils.FindGitBoundary`. A file directly at the repo root has pattern key `"."`. If `chat` isn't running inside a git repository at all (`FindGitBoundary` returns `""`), the pattern key falls back to the directory relative to the current working directory instead - "always" approval in that case is scoped to `ApprovalStore`'s in-memory-only behavior (see BR9), since there's no meaningful repo root to key persistence off of.

## Confirmation Gate Sequencing
- **BR3**: `ConfirmationSummary` is called (and can fail-closed, per BR4) **before** the gate is ever consulted - a call whose input can't be summarized never reaches the user or the approval-lookup logic.
- **BR4**: A `ConfirmationSummary` parse error is treated identically to a denial: an error `ToolResultBlock` is returned to the model immediately, no prompt shown, no store lookup performed, `Execute` never called.
- **BR5**: `git_diff`/`read_file` (`RequiresConfirmation() == false`) never call `ConfirmationSummary` at all - `Dispatch` calls `Execute` directly, exactly as it does today for `read_file`.

## Approval Lookup and Recording
- **BR6**: Before showing a prompt, `gate.Check` always checks `store.IsApproved(toolName, patternKey)` first. A match (from either tier) skips the prompt entirely and proceeds as if the user had just chosen "once" for this specific call.
- **BR7**: `store.IsApproved` checks the **session tier first, then the always tier** - functionally equivalent either order (both are simple "does an approval exist" checks), but session-first avoids an unnecessary disk-backed lookup path in the common case of a within-session repeat.
- **BR8**: Choosing "once" never calls any `Record*` method - by definition, it doesn't persist, in either tier.
- **BR9**: Choosing "session" calls `RecordSession` (in-memory only, never touches disk) - covers both the "inside a repo" and "not inside a repo" cases identically, since the session tier was never repo-scoped to begin with (only the always tier needs `FindGitBoundary`).
- **BR10**: Choosing "always" calls `RecordAlways`, which writes to `<fm.ConfigPath>/tool-approvals.yaml`. If `FindGitBoundary` returned `""` (not inside a git repo), "always" is **not offered** as a choice at the prompt (only once/session/deny) - there's no repo root to scope it to, and scoping an "always" approval to an arbitrary non-repo cwd would be a much easier approval to accidentally over-apply.
- **BR11**: A denied call (`DecisionDeny`) never calls any `Record*` method - a denial is never sticky (Story 8.3), the user is prompted again next time regardless of tier.

## Prompt Input Parsing
- **BR12**: The prompt reads exactly one line from the injected reader. The first non-whitespace character, lowercased, is matched: `o` → once, `s` → session, `a` → always (only offered per BR10), `n` → deny. Any other input (empty line, EOF, unrecognized character) defaults to **deny** - a fail-closed default, never fail-open.
- **BR13**: The prompt is shown exactly once per `Check` call that isn't short-circuited by BR6 - there is no retry-on-invalid-input loop in this pass (an invalid entry is simply a denial for that call; the model can ask again, prompting the user again).

## Storage Integrity
- **BR14**: `tool-approvals.yaml` is created with `0600` permissions if it doesn't exist. An existing file with looser permissions is not forcibly re-chmod'd in this pass (out of scope - not a regression this unit introduces, since the file doesn't exist before this unit ships).
- **BR15**: A malformed or unreadable `tool-approvals.yaml` at `ApprovalStore` construction is treated as "no persisted approvals" (empty always-tier) rather than a fatal error - consistent with #88's precedent of degrading gracefully on file-read problems (NFR4-equivalent reliability rule).
