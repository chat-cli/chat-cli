# Requirements: Universal Project-Context File Convention (#88)

## Intent Analysis Summary

- **User Request**: Following completion of the Bedrock capability catch-up initiative (#81-#85), user asked to discuss Phase 2 (the "agentic direction" group of issues: #86-#88). After discussing all three, user chose to redefine #88 from a chat-cli-specific `CHATCLI.md` idea into a universal `AGENTS.md`-first convention that also respects other tools' existing convention files already present in a repo (`CLAUDE.md`, Cursor rules, Copilot instructions), so users don't have to duplicate project context that's already written for another agentic tool.
- **Request Type**: New Feature (single cohesive capability)
- **Scope Estimate**: Single Component — new file-discovery/loading logic feeding into the existing system-prompt pipeline (`cmd/systemprompt.go`, `cmd/chat.go`) from #81
- **Complexity Estimate**: Simple — no new control flow like tool use introduced; this is discovery + read + compose, on top of already-existing system-prompt plumbing
- **Related GitHub Issue**: [#88](https://github.com/chat-cli/chat-cli/issues/88)
- **Depth**: Standard — one feature, but with several concrete precedence/scope decisions that need to be explicit since they affect UX and aren't obvious defaults

## Clarifying Questions and Answers

See `aidlc-docs/inception/requirements/agents-md-convention-questions.md` for the full record. Summary of decisions:

1. **Scope**: `chat` only (not `prompt` — project-context is a longer-session concept, matching the interactive-agent use case).
2. **Activation**: Automatic by default — no flag needed, matching how Claude Code/Cursor/etc. silently load their convention files. A way to disable it is still required (see FR6) since automatic, silent system-prompt injection needs an escape hatch.
3. **Precedence vs. explicit system prompt**: An explicit `--system` flag or configured `system-prompt` (from #81) wins entirely — the project-context file is not loaded at all if either is set. This is a simpler mental model than merging two system-prompt sources and avoids surprising interactions with #81's existing precedence chain.
4. **Cursor scope**: Dropped from this pass entirely. Only `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` are in scope. Cursor's real convention (`.cursor/rules/*.mdc`, a directory of frontmatter-scoped files) is structurally different enough from a flat markdown file that it deserves its own follow-up issue rather than a bolt-on here.
5. **Directory walk-up**: AI's best judgment (user deferred) — walk up from cwd toward the filesystem root, stopping at (and including) the first directory containing a `.git` entry; if no `.git` is found anywhere in the ancestry, only check the original cwd. This mirrors how Claude Code/Cursor resolve these files (repo-root-relative) while avoiding scanning unrelated ancestor directories when not inside a repo.

## Assumptions Made (flag in "Request Changes" if any are wrong)

1. **Filename precedence list is fixed at**: `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, in that order — first match wins, no concatenation of multiple files. `README.md` is explicitly excluded (too noisy/human-oriented for this to be a good automatic default).
2. **The list is configurable** via a new `context-files` config key (comma-separated, same config precedence pattern as `model-id`/`system-prompt`), so a user can reorder, trim, or extend it without a code change. The three built-in filenames above are the default when the config key is unset.
3. **Disabling it**: since this is automatic-by-default (decision 2 above), a `--no-context-file` flag (or an explicit empty `context-files` config value) is the escape hatch. `--system` already implicitly disables it per decision 3.
4. **Size guard**: content past 32KB is truncated (with a one-line warning to stderr) before being folded into the system prompt, to avoid a large file silently dominating the context window or the request. 32KB is a reasonable, generous ceiling for hand-written instruction files — most real-world `AGENTS.md`/`CLAUDE.md` files are a few KB.
5. **Cache synergy with #83**: the loaded file content is treated exactly like the existing system prompt for caching purposes — it flows through `buildSystemContentBlocks`/`withSystemCachePoint` so it gets the same automatic cache-point treatment already in place, no special-casing needed.
6. **Case sensitivity**: filename matching is exact-case (`AGENTS.md`, not `agents.md`), matching how the equivalent tools' own conventions behave and how most filesystems Go targets are actually case-sensitive (Linux/most CI); case-insensitive matching is not attempted.

## Functional Requirements

### FR1 — Automatic Discovery
- FR1.1: On `chat` startup, if neither `--system` nor a configured `system-prompt` is set, the CLI searches for a project-context file using the configured (or default) filename precedence list.
- FR1.2: Search starts in the current working directory. If no match is found there, the CLI walks up parent directories, stopping at (and including) the first directory containing a `.git` entry. If no `.git` is found in the ancestry, only the original cwd is checked.
- FR1.3: The first matching filename (in precedence order) at the first directory level that has any match is used. No merging of multiple files, even across different directory levels.

### FR2 — Precedence Over Explicit System Prompt
- FR2.1: If `--system` is passed, or a `system-prompt` config value is set, the project-context file is not searched for or loaded at all — the explicit value is used exactly as #81 already behaves today.

### FR3 — Content Loading and Composition
- FR3.1: The matched file's content becomes the system prompt for the session, via the existing `buildSystemContentBlocks` path from #81.
- FR3.2: Content longer than 32KB is truncated to 32KB with a one-line warning printed to stderr (not stdout, so it doesn't pollute piped output).
- FR3.3: The loaded content flows through the existing prompt-caching logic (#83) unchanged — a cache point is attached the same way it is for an explicit system prompt.

### FR4 — Configurable File List
- FR4.1: A new `context-files` config key (`chat-cli config set/unset/list`) holds a comma-separated, ordered list of filenames overriding the default `AGENTS.md,CLAUDE.md,.github/copilot-instructions.md` precedence list.

### FR5 — Disabling
- FR5.1: A `--no-context-file` flag on `chat` disables discovery for that invocation regardless of config.
- FR5.2: Setting `context-files` to an empty string via config disables discovery by default (until overridden per-invocation... there is no positive override needed since `--system` already takes precedence per FR2).

### FR6 — Visibility
- FR6.1: When a project-context file is loaded, `chat` prints a one-line notice identifying which file was used (e.g. `Using project context: AGENTS.md`), so the user isn't left guessing why the model's behavior changed. This uses the same dim/gray styling convention already used for other session metadata in `chat.go`.

## Non-Functional Requirements

- **NFR1 — Backward Compatibility**: Existing `chat` behavior (no `--system`, no config, no project-context file present) is completely unchanged. Existing integration tests must continue to pass unmodified.
- **NFR2 — TDD & Coverage**: Tests written before implementation, per `CLAUDE.md`. `cmd` package coverage must not regress.
- **NFR3 — Security**: File discovery is read-only and confined to the cwd-to-git-root walk-up path (FR1.2) — no arbitrary filesystem access, no symlink-following concerns beyond what a normal file read already has. This reuses the spirit of `utils.ValidateLocalPath` from #82/#84, adapted for a directory search rather than a single user-supplied path.
- **NFR4 — Reliability**: A missing, unreadable, or empty candidate file is treated as "no match" and search continues to the next filename/directory level — never a fatal error.
- **NFR5 — Performance**: The walk-up search is bounded by the filesystem's actual directory depth (typically single-digit levels in practice) and runs once at `chat` startup, not per-turn.

## Out of Scope (this pass)
- `prompt` command support (decision 1) — could be a fast follow if wanted later.
- Cursor's native `.cursor/rules/*.mdc` convention (decision 4) — tracked as a candidate follow-up issue if requested.
- `README.md` as an automatic fallback — remains available only via explicit `context-files` config, never a built-in default.
- Merging multiple matched files together — first-match-wins only.
