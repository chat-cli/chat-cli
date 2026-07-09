# Built-in Agent Tools (#86) - Clarification Questions

## What I heard from your answer

1. **No more `--tools` flag friction** — tool use should "just work." I'm reading this as: `chat` enables tool use automatically, no opt-in flag needed. Since Bedrock has no "does this model support tool use" capability check, I'd handle the risk the same way prompt caching (#83) already does — attempt it, and if a model/request rejects the `ToolConfiguration` field, silently retry once without it. This is a real change to Initiative 1's Unit 2 design (where `--tools` was deliberately opt-in specifically to avoid this risk) — confirming before I treat it as settled, since it revises a previous explicit decision rather than just adding to it.

2. **Every tool call requires confirmation**, with a session-level "sticky" option so you're not re-prompted for the same thing repeatedly.

3. **Pattern-based scoping for the "sticky" choice** — like Claude Code's own permission model (e.g. approving `git diff:*` covers any `git diff` invocation, not just one exact command). At the prompt, you'd choose whether to approve just this one call, or a broader pattern going forward.

This is a good, coherent design and it mirrors how Claude Code itself handles exactly this problem (fitting, since this session runs inside Claude Code). Before I write it into requirements.md, four things need pinning down:

## Question 1
Today's one existing tool, `read_file`, runs silently with no confirmation (Initiative 1 design). Should the new confirmation gate apply to **every** tool call uniformly (including `read_file` and the new read-only `git_diff`), or only to the destructive ones (`write_file`, `run_shell`)?

A) Every tool, uniformly - simpler mental model, no special-casing "safe" vs "risky" in the gate logic itself

B) Only `write_file`/`run_shell` - `read_file`/`git_diff` stay silent since they're read-only and this preserves `read_file`'s existing behavior exactly

C) Other (please describe after [Answer]: tag below)

[Answer]: B (destructive only - write_file/run_shell; read_file/git_diff stay silent)

## Question 2
What should the pattern granularity actually be for each tool, when you choose "allow going forward" rather than "just this once"?

A) `run_shell`: pattern matches on the **base command** (e.g. approving `git diff main` offers to approve all `git *` calls). `write_file`: pattern matches on **directory** (e.g. approving a write to `src/foo.go` offers to approve all writes under `src/*`).

B) `run_shell`: pattern matches on the **full command prefix** you can edit at the prompt (e.g. approve `git diff*` specifically, not all of `git`). `write_file`: pattern matches on a **glob you can edit** at the prompt (e.g. `src/**/*.go`).

C) Other (please describe after [Answer]: tag below)

[Answer]: A (coarse: base command for run_shell, directory for write_file)

## Question 3
Should approved patterns persist across `chat` sessions (saved to config, so you're not re-approving `git *` every time you start a new session), or are they always session-only (reset every time `chat` starts, matching how a fresh terminal session normally works)?

A) Session-only - resets every time `chat` starts (simpler, safer default, no new persisted state to manage)

B) Persisted - saved to the config file so approvals carry over between sessions

C) Other (please describe after [Answer]: tag below)

[Answer]: Neither strictly A nor B as originally framed - user wants a 3-way choice offered at prompt time: once / this session / always (persisted). Recording as a new decision, not a pre-listed letter.

## Question 4
Since `--tools` goes away as something you have to remember to pass, should there still be a flag to **disable** tool use entirely for someone who doesn't want it (mirroring `--no-context-file` from #88), or is tool use simply always-on with no opt-out?

A) Add `--no-tools` as the opt-out, mirroring `--no-context-file`'s precedent

B) Always-on, no opt-out flag in this pass

C) Other (please describe after [Answer]: tag below)

[Answer]: Neither A nor B as originally framed - no --no-tools flag; instead, a model/request that rejects the tool-use field is detected and tools are automatically disabled for the rest of that request (same retry-without-tools pattern as prompt caching's retry-without-cache), with a visible notice to the user.
