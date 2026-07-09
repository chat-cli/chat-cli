# Requirements Clarification Questions - Built-in Agent Tools (#86)

This feature has real safety implications (`write_file` is destructive, `run_shell` is arbitrary command execution), so more of this needs your input up front than Initiative 2 did.

## Question 1
Issue #86 names three tools: `write_file`, `run_shell`, `git_diff`. Should this pass build all three together, or stage them?

A) All three together - they're a natural set and share the same confirmation-flow design work

B) `write_file` and `git_diff` only this pass - defer `run_shell` (the highest-risk one) to a follow-up issue once the confirmation UX is proven out

C) `git_diff` only this pass (read-only, no confirmation-flow design needed at all) - defer both destructive tools

D) Other (please describe after [Answer]: tag below)

[Answer]:

## Question 2
Today, `--tools` unconditionally exposes the one existing built-in tool (`read_file`, read-only, cwd-confined). Should `write_file`/`run_shell` be available under that same `--tools` flag, or gated separately?

A) Same `--tools` flag - once you opt into tool use at all, all built-in tools (including destructive ones) are available, gated instead by a per-call confirmation prompt

B) A separate opt-in is required for destructive tools (e.g. `--dangerous-tools` or similar) on top of `--tools` - so someone can enable read-only tool use without exposing write/exec at all

C) Other (please describe after [Answer]: tag below)

[Answer]: 

## Question 3
What should the confirmation prompt look like for a destructive tool call (`write_file`, `run_shell`)?

A) Show exactly what will happen (the file path + new content diff for `write_file`; the exact command string for `run_shell`), then a `y/n` prompt that blocks the conversation until answered

B) Same as A, but also offer a session-level "always allow" choice (e.g. `y`/`n`/`a`) so the user isn't re-prompted for every subsequent destructive call in the same session

C) No interactive prompt in this pass - require an explicit `--auto-approve-tools` (or similar) flag at startup instead, and refuse all destructive calls without it

D) Other (please describe after [Answer]: tag below)

[Answer]: 

## Question 4
Should there be a non-interactive escape hatch (a flag to skip confirmation prompts entirely, for scripting/automation use), given `chat` is normally interactive?

A) Yes - an explicit flag (e.g. `--yolo` or `--auto-approve`) that skips all confirmation prompts, off by default

B) No - confirmation is always required in this pass; a bypass flag can be a follow-up if there's demand

C) Other (please describe after [Answer]: tag below)

[Answer]: 

## Question 5
Scope for `write_file` - should it reuse the exact same cwd-confinement rule as `read_file` (`utils.ValidateLocalPath`), and can it create new files or only overwrite existing ones?

A) Same cwd confinement as `read_file`; can both create new files and overwrite existing ones (within cwd)

B) Same cwd confinement, but overwrite-only - creating brand-new files is out of scope for this pass

C) Other (please describe after [Answer]: tag below)

[Answer]: 

## Question 6
Scope for `run_shell` - how should the command actually run?

A) Run via the shell (e.g. `sh -c "<command>"`) with the working directory fixed to `chat-cli`'s own cwd (same confinement spirit as the file tools), a fixed timeout, and combined stdout+stderr (truncated past some size) returned to the model - no allowlist/denylist of specific commands

B) Same as A, but with an explicit command allowlist (e.g. only `git`, `ls`, `cat`, a small fixed set) rather than allowing arbitrary commands

C) Other (please describe after [Answer]: tag below)

[Answer]: 

## Question 7
Scope for `git_diff` - what exactly should it run, and does it need arguments?

A) Runs `git diff` with no arguments in `chat-cli`'s cwd, returns the raw output - if cwd isn't a git repo, returns a clear (non-fatal) error to the model

B) Same as A, but accepts an optional argument (e.g. a path or ref) the model can pass, mirroring `git diff <path>`/`git diff <ref>`

C) Other (please describe after [Answer]: tag below)

[Answer]: 
