# Requirements Clarification Questions - Universal Project-Context File Convention (#88)

Most of the design was already settled in conversation (priority list order, no-merge/first-match-wins policy, configurable file list via a config key, size guard, prompt-cache synergy with #83). These questions cover the remaining open parameters before requirements.md is written.

## Question 1
Which command(s) should auto-load the project-context file into the system prompt?

A) Both `chat` and `prompt` (consistent with how `--system` already works in both, from #81)

B) `chat` only (project-context feels most relevant to a longer interactive session)

C) `prompt` only

D) Other (please describe after [Answer]: tag below)

[Answer]:

## Question 2
Should this be automatic/always-on (matching how Claude Code/Cursor/etc. silently load their convention files), or opt-in like tool use (`--tools`) was?

A) Automatic by default - matches the convention other agentic tools already follow, with a way to disable it (flag/config) for users who don't want it

B) Opt-in via a flag (e.g. `--agents-file`), off unless explicitly requested

C) Opt-in via config only (`chat-cli config set context-files ...`), no CLI flag needed for v1

D) Other (please describe after [Answer]: tag below)

[Answer]:

## Question 3
How should the project-context file's content combine with an explicit `--system` flag or configured `system-prompt` (from #81) when both are present?

A) Include both - project-context file content plus the explicit system prompt, as separate blocks (file first, then explicit prompt)

B) Explicit `--system`/config `system-prompt` wins entirely - project-context file is ignored if either is set

C) Project-context file is the base, explicit `--system` flag can only append to it (config `system-prompt` is ignored in favor of the file)

D) Other (please describe after [Answer]: tag below)

[Answer]:

## Question 4
What scope should Cursor's convention get in this first pass, given `.cursor/rules/*.mdc` is a directory of multiple frontmatter-scoped files (structurally different from a single flat markdown file), while `.cursorrules` is a legacy single-file format?

A) Support only the legacy single-file `.cursorrules` for now; defer `.cursor/rules/*.mdc` directory support to a follow-up issue

B) Support both now (adds meaningfully more parsing complexity - directory scan + frontmatter handling - to this unit)

C) Drop Cursor support entirely from this pass; only cover `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md`

D) Other (please describe after [Answer]: tag below)

[Answer]:

## Question 5
How far should the walk-up-to-repo-root search go if the file isn't found in the current directory?

A) Walk up until a `.git` directory is found (repo root); if no `.git` anywhere in the path, check cwd only - don't walk indefinitely toward filesystem root

B) cwd only, no walk-up (simplest, matches original #88 scope)

C) Walk up unconditionally until filesystem root if no `.git` is found

D) Other (please describe after [Answer]: tag below)

[Answer]:
