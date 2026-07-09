# Application Design (Consolidated)

**Scope**: Issues #81-#85. Design is deliberately light for 4 of the 5 features (system prompt, prompt caching, document input, extended thinking are small extensions to existing `cmd/chat.go`/`cmd/prompt.go` orchestration) and focused where it matters: the new tool-use subsystem, plus two small shared utility extractions that prevent duplicated logic across features. See `application-design-plan.md` for the assumptions this scoping is based on.

This document consolidates `components.md`, `component-methods.md`, `services.md`, and `component-dependency.md` — see those files for full detail.

## New Components (summary)
1. **`tools.Tool`** (interface) — uniform contract for a callable tool.
2. **`tools.Registry`** — builds `ToolConfig`, dispatches tool calls, wraps unknown/failed tools as error results instead of CLI crashes.
3. **`tools.ReadFileTool`** — the one built-in tool shipped this pass (Story 2.2).
4. **`utils.ValidateLocalPath`** — path-traversal-safe validation, extracted from `ReadImage` and reused by `ReadFileTool` and the new document-attachment flow (avoids adding to the existing duplicated-validation-logic debt tracked in #92).
5. **Cache-point helper** (in `utils`) — shared "attempt cache point, fall back once on rejection" logic reused by both commands.

## Extended (Not Redesigned) Components
- `config.FileManager` — one new config key (`system-prompt`), same precedence mechanism as `model-id`.
- `repository.ChatRepository` — used as-is, no interface changes.
- `cmd/chat.go`, `cmd/prompt.go` — existing orchestration extended in place (see `services.md`); no new command files, no new Cobra commands.

## What's Explicitly Deferred
- Detailed business rules for tool dispatch, cache-fallback retry, and document format validation → per-unit **Functional Design** in Construction.
- Exact error message text, flag help text wording → **Code Generation**.
- Whether NFR Requirements/NFR Design execute per unit (likely yes for Tool Use and Document Input, given the security-sensitive file access both introduce) → decided per-unit during Construction, per `execution-plan.md`.

## Consistency Check
- No component introduces a circular dependency (see `component-dependency.md`'s matrix — all new components depend only on existing `utils`/stdlib, never the reverse).
- No component duplicates the path-safety logic already in `utils.ReadImage` — it's extracted and shared instead (directly addresses the pattern flagged as a risk in `code-quality-assessment.md`'s "Duplicated model-validation logic" finding, applied here proactively to file-path validation).
- No new component touches `db`/`db/sqlite`/`factory` — confirms the "no data model changes" call in `execution-plan.md`.

---

# Initiative 3 Application Design (Consolidated, #86)

**Scope**: Extends the tool-use subsystem from Initiative 1 (`tools.Tool`/`Registry`) with 3 new tools and a genuinely new permission-engine subsystem (`PermissionGate`, `ApprovalStore`, `InteractivePermissionGate`), plus wiring changes to `cmd/chat.go` (automatic enablement, `--tools` removal, retry-without-tools). Unlike Initiative 2, this warranted full Application Design treatment given the real new architecture involved — see `builtin-tools-execution-plan.md` for the risk-based rationale.

This section summarizes; see `components.md`, `component-methods.md`, `services.md`, `component-dependency.md` (each has an "Initiative 3" section appended) for full detail.

## New Components (summary)
1. **Extended `tools.Tool` interface** — adds `RequiresConfirmation()`/`ConfirmationSummary()`.
2. **`tools.WriteFileTool`, `tools.RunShellTool`, `tools.GitDiffTool`** — the 3 new built-in tools.
3. **`tools.PermissionGate`** (interface) + **`tools.ApprovalStore`** — the abstract permission-decision contract and its session/persisted-tier bookkeeping.
4. **`cmd.InteractivePermissionGate`** — the concrete terminal-prompting implementation.
5. **`utils.FindGitBoundary`** (extracted) — shared repo-root detection, reused by both `tools.ApprovalStore` (new) and `cmd/projectcontext.go` (#88, refactored from a private copy).

## Extended (Not Redesigned) Components
- `tools.Registry.Dispatch` — signature gains a `PermissionGate` parameter, consulted before any destructive tool's `Execute`.
- `cmd/chat.go` — `--tools` flag removed; registry/gate construction and the retry-without-tools fallback added to the existing orchestration.

## What's Explicitly Deferred
- Exact confirmation-prompt UI text, the persisted-store's file format/location, and `run_shell`'s exact timeout/output-cap values → per-unit **Functional Design** in Construction.
- Whether the interactive prompt itself is unit-testable via an injected reader, or needs a documented manual-verification gap → decided in the Permission Engine unit's Functional Design.

## Consistency Check
- `tools` still does not import `cmd` — the one component that would have needed it (`ApprovalStore`'s repo-boundary detection) is resolved via extracting the logic to `utils` instead, preserving the existing one-directional dependency shape (`cmd` → `tools`, never the reverse).
- The extraction of `utils.FindGitBoundary` is the only change to already-shipped Initiative 2 code in this whole initiative — flagged explicitly, not incidental.
- No new component touches `db`/`db/sqlite`/`factory`/`repository`/`config` — this initiative is fully contained within `tools`/`cmd`/one `utils` extraction.
