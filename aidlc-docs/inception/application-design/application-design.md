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
