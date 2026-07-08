# Functional Design: Business Logic Model - AGENTS.md Convention (#88)

## Context
No Units Generation stage ran for this initiative (single unit, per the execution plan) — this document designs the whole feature directly against `aidlc-docs/inception/requirements/agents-md-convention-requirements.md` (FR1-FR6).

Verified against the actual current codebase (not guessed):
- `cmd/systemprompt.go` — `buildSystemContentBlocks(systemPrompt string)`, already returns `nil` for an empty string, so anything feeding into it degrades safely.
- `cmd/chat.go:114` — `systemPrompt := fm.GetConfigValue("system-prompt", systemFlag, "").(string)` is the exact resolution point FR2 hooks into.
- `cmd/config.go` — `supportedConfigKeys` map and three separate hardcoded key lists (`Long` help text ×2, `configListCmd`'s `configKeys` slice) all need the new `context-files` key added, following the existing (imperfect but established) pattern rather than refactoring it.
- Gray/dim ANSI styling (`\033[90m...\033[0m`) is already used in `chat.go` for the user-echo line and `[thinking]` prefix — reused for the startup notice (FR6) for visual consistency.

## Core Algorithm: Project-Context Discovery

Resolves FR1.2's "walks up parent directories, stopping at the first directory containing `.git`" into a concrete, two-phase algorithm (rather than checking every intermediate directory level for candidate files, which would mean more filesystem reads for no real benefit):

**Phase A - Locate the search boundary** (cheap: only checks for a `.git` entry's existence, never reads file content):
1. Start at cwd.
2. If `.git` exists in the current directory (file or directory — worktrees use a `.git` file), that directory is the boundary. Stop.
3. Otherwise move to the parent directory and repeat.
4. If filesystem root is reached with no `.git` found, there is no boundary directory (only cwd itself will be checked in Phase B).

**Phase B - Check for a candidate file**:
1. Check cwd first: for each filename in the effective precedence list (in order), does a *regular file* (not a directory) with that exact name exist and is it readable? First match wins - stop and use it.
2. If no match in cwd, and a boundary directory was found in Phase A and it differs from cwd, check the boundary directory the same way (precedence order, first match wins).
3. If still no match, discovery yields no result. `chat` behaves exactly as it does today (no system prompt).

This means: at most two directories are ever checked for candidate files (cwd and the repo root), regardless of how deep cwd is nested inside the repo. Intermediate directories are only ever probed for `.git`'s existence, never read for content. This satisfies FR1.2 ("walks up ... stopping at ... `.git`") while keeping the search cheap and bounded, and matches the "only cwd is checked" fallback when no repo is detected at all.

## Entry Point and Call Site

- New pure function, `resolveProjectContext(cwd string, candidates []string) (path string, content string, found bool)`, in a new file (see domain-entities.md for the exact file/function shape).
- Called from `cmd/chat.go`, immediately after the existing line `systemPrompt := fm.GetConfigValue("system-prompt", systemFlag, "").(string)`, **only when**:
  - `systemPrompt == ""` (neither `--system` nor configured `system-prompt` supplied anything - FR2.1), **and**
  - the new `--no-context-file` flag is `false` (FR5.1), **and**
  - the effective `context-files` candidate list (FR4.1, default or config-overridden) is non-empty (FR5.2).
- On a match, `systemPrompt` is reassigned to the file's (possibly truncated) content, and the FR6 notice is printed. Everything downstream (`buildSystemContentBlocks`, `withSystemCachePoint`) is completely unchanged — it already treats whatever string it's given as "the system prompt," with no knowledge of where it came from.
- `cmd/prompt.go` is untouched (chat-only scope, decision 1).

## Data Flow

```
cwd, --no-context-file, context-files config
        |
        v
[effective candidate list resolved: config value if set, else default 3-name list]
        |
        v
[Phase A: walk up from cwd, stat-only, find .git boundary or none]
        |
        v
[Phase B: check cwd, then boundary dir, for first matching filename]
        |
        v
[no match] -> systemPrompt stays "" (today's behavior, unchanged)
[match]    -> read file -> trim -> truncate at 32KB with stderr warning if needed
              -> if resulting content is non-empty after trim: systemPrompt = content,
                 print FR6 notice; else: treat as no match (vacuous file), systemPrompt stays ""
```
