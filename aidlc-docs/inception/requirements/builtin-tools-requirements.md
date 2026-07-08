# Requirements: Built-in Agent Tools (#86)

## Intent Analysis Summary

- **User Request**: Following Initiative 2 (#88), user picked #86 next: "Once tool use is supported, add a small first-party toolset — read file, write file, run a shell command, git diff — so chat-cli chat can act as a lightweight, Bedrock-native coding assistant directly in the terminal... Needs careful scoping around confirmation prompts before destructive actions." Through two rounds of clarifying questions, the user pushed the scope further than the issue's original text: tool use should become automatic (no `--tools` flag to remember), gated instead by a per-call confirmation system for destructive actions, with pattern-based "sticky" approval modeled on how permission prompts work in coding-agent tools generally.
- **Request Type**: New Feature, with a revision to an existing feature (Initiative 1 Unit 2's `--tools` opt-in flag is being replaced by automatic enablement + graceful degradation)
- **Scope Estimate**: Multiple Components — extends the existing `tools/` package (2 new tools), introduces a genuinely new subsystem (the confirmation/permission-pattern engine) that nothing in the codebase resembles today, and changes `cmd/chat.go`'s tool-enablement wiring from Unit 2
- **Complexity Estimate**: Complex — the permission-pattern engine (three-tier once/session/always approval, coarse pattern matching, per-repo persisted storage) is new architecture, not a simple flag/config addition like Initiatives 1-2
- **Related GitHub Issue**: [#86](https://github.com/chat-cli/chat-cli/issues/86)
- **Depth**: Comprehensive — real security-relevant design (arbitrary shell execution, filesystem writes), multiple new interaction flows, and a new persisted-state concept. Given this, User Stories and Application Design are likely warranted (final call in Workflow Planning) rather than skipped as they were in Initiative 2.

## Clarifying Questions and Answers

Two rounds - see `aidlc-docs/inception/requirements/builtin-tools-questions.md` and `builtin-tools-clarification-questions.md` for the full record. Summary of decisions:

1. **Tool scope**: build `write_file`, `run_shell`, and `git_diff` together in this pass (alongside the existing `read_file` from Unit 2).
2. **Tool-use enablement**: automatic, no `--tools` flag. `chat` always builds and sends a `ToolConfiguration`. If a model/request rejects it, `chat` automatically retries once without tools (mirroring #83's cache-point retry pattern) and tells the user tools were turned off for that session - no `--no-tools` opt-out flag needed, since the automatic-disable-on-rejection already covers "a model that can't/shouldn't use tools."
3. **Confirmation gate scope**: only for destructive tools (`write_file`, `run_shell`). The existing `read_file` and the new `git_diff` remain silent/no-confirmation, since they're read-only - consistent with `read_file`'s existing Unit 2 behavior.
4. **Confirmation UX**: a blocking prompt before every destructive call, showing exactly what will happen (the command string for `run_shell`; the path + content for `write_file`), with a three-way choice: **approve once**, **approve for the rest of this session**, or **approve always** (persisted).
5. **Pattern granularity** (for "session"/"always" approval): coarse. `run_shell` patterns match on the **base command** (e.g. approving one `git diff` call offers to approve all `git *` calls for the chosen scope). `write_file` patterns match on **directory** (e.g. approving a write to `src/foo.go` offers to approve all writes under `src/*`).
6. **Persistence for "always"**: persisted approvals are scoped **per-repository** (see Assumption 4 below for the exact mechanism), not global across every project - approving `git *` in one repo shouldn't silently apply in an unrelated one.

## Assumptions Made (flag in "Request Changes" if any are wrong)

1. **`write_file`** reuses `utils.ValidateLocalPath` (the same cwd-confinement already used by `read_file` and `--document`) and can both create new files and overwrite existing ones within that confinement.
2. **`run_shell`** executes via the shell (`sh -c "<command>"`), working directory fixed to `chat-cli`'s own cwd, with a fixed timeout (proposed: 30s) and combined stdout+stderr truncated past a size cap (proposed: 32KB, matching #88's precedent) returned to the model. No command allowlist/denylist - the confirmation gate is the control, not a static list.
3. **`git_diff`** accepts an optional argument (path or ref) mirroring `git diff <arg>`; with no argument, runs a plain `git diff`. If cwd isn't inside a git repository, it returns a clear error *to the model* (a `ToolResultBlock` error, not a fatal CLI error) - consistent with Unit 2's fail-closed dispatch pattern.
4. **Persisted-approval storage**: a new store, scoped by git repository root (reusing #88's `.git`-boundary-detection concept - the same directory `resolveContextFilenames`'s boundary walk would find), separate from the existing flat `config set`/`unset` key-value system (which isn't shaped for structured, per-repo, per-pattern data). Exact file format/location is a Functional Design decision, not fixed here.
5. **Session-scoped approvals** (the "this session" tier) live in memory only, tied to the running `chat` process, and are gone once it exits - no file I/O involved for that tier.
6. **The existing `--tools` flag is removed** (not deprecated-but-kept) since automatic enablement supersedes it entirely, and Initiative 1 established a precedent of not carrying forward flags that no longer do anything meaningful.
7. **Backward compatibility scope**: this initiative *intentionally* changes default `chat` behavior (tools are now always attempted) - this is a deliberate, user-directed revision of Initiative 1's NFR1, not a violation of it. Everything else (no confirmation for read-only tools, graceful degradation for non-tool models) is designed to keep the change low-friction.
8. **`prompt` is unaffected** - this entire initiative, like Initiative 2, is `chat`-only. Tool use has never been wired into `prompt`.

## Functional Requirements

### FR1 - Automatic Tool-Use Enablement
- FR1.1: `chat` always builds a non-empty `ToolConfiguration` (via the existing `tools.Registry` from Unit 2, now populated with 4 tools) and attaches it to every request - no flag required.
- FR1.2: If a request is rejected because the model/request doesn't support tool use, `chat` retries once without the `ToolConfiguration`, mirroring #83's existing cache-point retry pattern exactly (detect failure, strip the field, retry, surface normally if that also fails).
- FR1.3: When the automatic retry-without-tools happens, the user sees a clear, one-line notice that tools were disabled for this session (not a silent degradation).
- FR1.4: The `--tools` flag is removed from `cmd/root.go`.

### FR2 - `write_file` Tool
- FR2.1: New tool, `write_file`, taking a path and content, confined to the working directory via `utils.ValidateLocalPath`.
- FR2.2: Can create a new file or overwrite an existing one.
- FR2.3: Requires confirmation (FR5) before executing - never writes without an approved gate.

### FR3 - `run_shell` Tool
- FR3.1: New tool, `run_shell`, taking a command string, executed via the shell in `chat-cli`'s cwd, with a fixed timeout and truncated combined stdout+stderr returned to the model.
- FR3.2: Requires confirmation (FR5) before executing.

### FR4 - `git_diff` Tool
- FR4.1: New tool, `git_diff`, taking an optional path/ref argument, running `git diff [arg]` in cwd.
- FR4.2: Read-only - does not require confirmation (FR5 doesn't apply).
- FR4.3: A cwd outside any git repository produces a clear `ToolResultBlock` error back to the model, not a fatal CLI error.

### FR5 - Confirmation Gate (Destructive Tools Only)
- FR5.1: Before `write_file` or `run_shell` executes, the CLI shows the user exactly what will happen (file path + content for `write_file`; the exact command string for `run_shell`) and blocks for a decision.
- FR5.2: The user chooses one of: approve **once**, approve for the rest of **this session**, approve **always** (persisted), or **deny**.
- FR5.3: A denied call is reported back to the model as a `ToolResultBlock` error (e.g. "user declined this action"), not a fatal error - the conversation continues.
- FR5.4: `read_file` and `git_diff` never trigger this gate (FR2.3/FR3.2 vs. FR4.2).

### FR6 - Pattern-Based Sticky Approval
- FR6.1: For `run_shell`, a "session"/"always" approval is scoped to the **base command** (the first whitespace-separated token, e.g. `git`, `npm`) - approving one call offers to approve that base command generally, not just the exact string.
- FR6.2: For `write_file`, a "session"/"always" approval is scoped to the **containing directory** - approving one write offers to approve further writes anywhere under that directory.
- FR6.3: Before executing a destructive call, the CLI checks existing session and persisted approvals (FR6.1/FR6.2 patterns) and skips the confirmation gate (FR5) entirely if a match is found.

### FR7 - Persisted ("Always") Approval Storage
- FR7.1: "Always" approvals are stored per git repository (keyed by the repo root, same boundary-detection concept as #88), so approvals in one project don't apply in another.
- FR7.2: "Session" approvals are in-memory only and never touch disk.

## Non-Functional Requirements

- **NFR1 - Security**: The confirmation gate (FR5) is the primary control for destructive actions, given there is no static command allowlist for `run_shell` (Assumption 2). `write_file` remains cwd-confined via the existing `utils.ValidateLocalPath` choke point. The persisted-approval store (FR7) must not be world-readable/writable (reuse the existing config file's permission conventions).
- **NFR2 - Reliability**: FR1.2's automatic retry-without-tools ensures `chat` never hard-fails for a model that rejects tool use - this is the single most important reliability property of this initiative, since tool use is no longer opt-in.
- **NFR3 - Backward Compatibility (revised)**: This initiative intentionally changes default behavior (Assumption 7) - flagged prominently rather than silently treated as a violation of Initiative 1's original NFR1.
- **NFR4 - TDD & Coverage**: Per `CLAUDE.md`, tests before implementation. `cmd`/`tools` package coverage must not regress.
- **NFR5 - Usability**: The confirmation prompt (FR5.1) must show enough information for an informed decision (full command string or file path+content), not a truncated/vague summary.

## Out of Scope (this pass)
- A static command allowlist/denylist for `run_shell` - the confirmation gate is the control instead.
- Editable/custom patterns at prompt time (Question 2's option B) - coarse, fixed granularity only (base command / directory).
- Extending any of this to `prompt` (chat-only, consistent with #88).
- MCP-sourced tools (#87) - this initiative only covers the 3 new built-in tools plus the existing `read_file`.
