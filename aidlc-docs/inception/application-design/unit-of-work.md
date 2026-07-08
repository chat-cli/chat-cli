# Unit of Work

chat-cli is a monolith — these 5 units are logical modules within the single `chat-cli` binary, not independently deployable services. All 5 ship together in the same release.

## Unit 1 — System Prompt Support
- **Scope**: FR1.1-FR1.4, Story 1.1, issue #81
- **Files touched**: `cmd/chat.go`, `cmd/prompt.go`, `config/config.go`
- **New components**: None — direct additions to existing request-building code and `FileManager`'s supported-keys list
- **Definition of done**: `--system` flag and `system-prompt` config key both work with correct precedence; `SystemContentBlocks` sent only when set; existing behavior unchanged when unset

## Unit 2 — Tool Use / Function Calling
- **Scope**: FR2.1-FR2.5, Stories 2.1 + 2.2, issue #82
- **Files touched**: `cmd/chat.go`; new `tools/` package
- **New components**: `tools.Tool` (interface), `tools.Registry`, `tools.ReadFileTool`, `utils.ValidateLocalPath`
- **Definition of done**: `chat` builds a `ToolConfig` when tools are registered; tool-use round trips work end-to-end via the built-in `read_file` tool; unknown-tool and execution-failure cases return error results to the model instead of crashing; tool-call turns persist through the existing `ChatRepository`

## Unit 3 — Prompt Caching
- **Scope**: FR3.1-FR3.4, Story 3.1, issue #83
- **Files touched**: `cmd/chat.go`, `cmd/prompt.go`; new cache-point helper in `utils`
- **New components**: Cache-point helper (`utils`)
- **Definition of done**: Cache checkpoints inserted after system prompt / piped document content; graceful one-time fallback retry on rejection; no user-visible failure when caching isn't supported by the selected model

## Unit 4 — Native Document Input
- **Scope**: FR4.1-FR4.4, Story 4.1, issue #84
- **Files touched**: `cmd/prompt.go`; new `utils.ReadDocument` (reuses `utils.ValidateLocalPath` from Unit 2, or introduces it if Unit 4 lands first)
- **New components**: `utils.ReadDocument`
- **Definition of done**: `--document` flag attaches supported file types as `ContentBlockMemberDocument`; unsupported extensions/out-of-bounds paths rejected with a clear error before calling Bedrock; `--image` behavior completely unchanged; both flags usable together

## Unit 5 — Extended Thinking / Reasoning Mode
- **Scope**: FR5.1-FR5.4, Story 5.1, issue #85
- **Files touched**: `cmd/chat.go`, `cmd/prompt.go`
- **New components**: None — direct additions to existing request-building code and streaming-output rendering
- **Definition of done**: `--thinking` flag sets `AdditionalModelRequestFields`; reasoning content rendered distinctly from the final answer; unsupported-model rejection surfaced clearly; behavior unchanged when flag is unset

## Code Organization Note (Brownfield)
No new top-level directory structure beyond the one new `tools/` package introduced in Unit 2 (see `application-design/components.md`). All other changes extend existing files in `cmd/`, `config/`, and `utils/` in place, per the "Application Code: Workspace root" rule in `aidlc-state.md`.

---

# Initiative 3 Units (#86)

3 units this time (vs. Initiative 2's single implicit unit) — see `builtin-tools-execution-plan.md` for why Units Generation executes here: multiple packages, new persisted state, and a natural dependency seam (the permission engine must exist before the destructive tools can use it).

## Unit 6 — Confirmation and Sticky Approval Engine
- **Scope**: FR5.1-FR5.4, FR6.1-FR6.3, FR7.1-FR7.2, Stories 8.1-8.3, issue #86
- **Files touched**: `tools/tool.go` (interface extended), new `tools/permission.go` (`PermissionGate`, `Decision`), new `tools/approvalstore.go` (`ApprovalStore`), new `cmd/permissionprompt.go` (`InteractivePermissionGate`), `tools/registry.go` (`Dispatch` signature extended), `utils/utils.go` (new exported `FindGitBoundary`), `cmd/projectcontext.go` (refactored to call it instead of its private copy)
- **New components**: `tools.PermissionGate`, `tools.ApprovalStore`, `cmd.InteractivePermissionGate`, `utils.FindGitBoundary`
- **Definition of done**: A destructive tool call blocks for a once/session/always/deny decision; session approvals are forgotten when `chat` exits; always approvals persist and are scoped per git repository (verified: an approval in one repo doesn't apply in another); a denied call returns a clear `ToolResultBlock` error, not a crash
- **No dependency** on Units 7/8 - this is the foundational unit

## Unit 7 — New Built-in Tools
- **Scope**: FR2.1-FR2.3, FR3.1-FR3.2, FR4.1-FR4.3, Stories 7.1-7.3, issue #86
- **Files touched**: new `tools/writefile.go`, `tools/runshell.go`, `tools/gitdiff.go`
- **New components**: `tools.WriteFileTool`, `tools.RunShellTool`, `tools.GitDiffTool`
- **Definition of done**: all 3 tools implement the extended `Tool` interface correctly (`write_file`/`run_shell` declare `RequiresConfirmation() == true` with correct summaries/pattern keys; `git_diff` declares `false`); `write_file` stays cwd-confined; `run_shell` respects its timeout/output cap; `git_diff` handles the non-repo case cleanly
- **Depends on**: Unit 6 (the extended `Tool` interface and `PermissionGate` must exist before these tools can implement/use them)

## Unit 8 — Automatic Tool-Use Enablement
- **Scope**: FR1.1-FR1.4, Stories 6.1-6.2, issue #86
- **Files touched**: `cmd/chat.go` (registry/gate construction, `--tools` removed, retry-without-tools wrapper), `cmd/root.go` (`--tools` flag registration removed)
- **New components**: None - wiring changes to existing orchestration, plus a retry wrapper structurally identical to `cmd/promptcache.go`'s existing pattern
- **Definition of done**: no `--tools` flag exists; every `chat` request always attaches a `ToolConfiguration`; a model/request that rejects it triggers exactly one retry without it, with a visible notice; `chat --help` no longer lists `--tools`
- **Depends on**: Units 6+7 conceptually complete first (so the registry being "always built" has the full 4-tool set and working gate to attach), though technically buildable independently — sequenced last for a coherent incremental build

## Recommended Build Order
Unit 6 → Unit 7 → Unit 8 (foundation → consumers → the final flip-the-switch wiring change)
