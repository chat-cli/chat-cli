# Functional Design: Business Rules - AGENTS.md Convention (#88)

## Precedence and Matching Rules
- **BR1**: Filename matching is exact-case (`AGENTS.md`, not `agents.md`/`Agents.md`), per Assumption 6 in requirements.md.
- **BR2**: A candidate name must resolve to a *regular file*. A directory or other non-regular entry (symlink to a directory, device file, etc.) with a matching name is not a match - move to the next candidate. A symlink to a regular file IS a valid match (normal file-read semantics, no special-casing - this is a local convenience file a user placed themselves, not untrusted input from a tool call).
- **BR3**: Within a checked directory, the first filename in precedence order that matches wins - remaining candidates at that directory level are not checked.
- **BR4**: cwd is checked before the repo-root boundary directory. If cwd itself IS the repo root (has `.git` directly), only one directory is checked (no duplicate check).
- **BR5**: The default precedence list is `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md` (requirements.md Assumption 1). `.github/copilot-instructions.md` is a relative path (one path segment with a directory component) - matching for this candidate means "does `<checked-dir>/.github/copilot-instructions.md` exist as a regular file," same file-existence rule as the flat filenames.

## Precedence vs. Explicit System Prompt (FR2)
- **BR6**: If `--system` is passed with any value (including a value that happens to be an empty string via `--system ""`), OR a `system-prompt` config value is set, discovery is never attempted. This is decided once, before Phase A/B run at all - no wasted filesystem work when it's going to be discarded anyway.

## Content Handling
- **BR7**: File content is read as raw bytes, decoded as UTF-8, then whitespace-trimmed (leading and trailing) before any further processing.
- **BR8**: If content is empty after trimming (BR7) - whether the file was truly empty or contained only whitespace - it is treated as **no match** and the search rule in BR3/BR4 (move to next candidate, then next directory) continues to apply, rather than "found but empty." Rationale: a vacuous file shouldn't silently produce a no-op FR6 notice, and it costs nothing extra to keep checking the remaining candidates.
- **BR9**: Content longer than 32KB (after trimming) is truncated to exactly 32KB, and a one-line warning is printed to **stderr** (not stdout - FR3.2) identifying the truncated file's path and original size, before the (truncated) content is used as the system prompt.
- **BR10**: A file that exists but can't be read (permission error, race condition where it's deleted between stat and read, etc.) is treated as **no match** per NFR4 - the search continues exactly as if the file didn't exist. This is never a fatal error for `chat` to start.

## Configuration Rules (FR4)
- **BR11**: `context-files` config value, when set, is parsed as a comma-separated list. Each entry is whitespace-trimmed; empty entries (from `,,` or leading/trailing commas) are dropped.
- **BR12**: If `context-files` is set to an empty string (`chat-cli config set context-files ""`) or a value that trims down to zero entries after BR11, the effective candidate list is empty - this is precisely FR5.2's disable-via-config mechanism, not an error.
- **BR13**: If `context-files` is unset entirely, the default 3-name list (BR5) is used.

## Disabling (FR5)
- **BR14**: `--no-context-file` (a session-scoped flag, not persisted) always wins over any config state - if passed, discovery is skipped unconditionally, even if `context-files` is configured with a non-empty list.

## Notice (FR6)
- **BR15**: The notice is only printed when a match was actually used (i.e., content was non-empty after BR8's check) - never printed for the "no match" or "disabled" cases, since there's nothing to announce.
- **BR16**: The notice shows the path exactly as resolved (e.g. `AGENTS.md` if found in cwd, or a relative-to-cwd path like `../AGENTS.md` if found at the repo-root boundary instead) so the user can tell which of the two checked locations supplied it.
