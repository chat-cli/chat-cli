# AI-DLC State Tracking

## Project Information
- **Project Type**: Brownfield
- **Start Date**: 2026-07-08T00:00:00Z
- **Current Stage**: INITIATIVE COMPLETE - all units (#81-#85) approved, Build and Test approved. Operations phase remains a placeholder (no deployment/monitoring workflow planned). No PR created yet (not requested).

## Workspace State
- **Existing Code**: Yes
- **Programming Languages**: Go (26 .go files)
- **Build System**: Go modules (go.mod), Makefile, GoReleaser
- **Project Structure**: Monolith CLI (Cobra command layers: cmd/, config/, db/, repository/, factory/, utils/)
- **Reverse Engineering Needed**: Yes
- **Workspace Root**: /home/user/chat-cli

## Code Location Rules
- **Application Code**: Workspace root (NEVER in aidlc-docs/)
- **Documentation**: aidlc-docs/ only
- **Structure patterns**: See code-generation.md Critical Rules

## Extension Configuration
(none loaded yet - no extensions opted into)

## Reverse Engineering Status
- [x] Reverse Engineering - Completed on 2026-07-08T00:00:00Z
- **Artifacts Location**: aidlc-docs/inception/reverse-engineering/

## Stage Progress
- [x] Workspace Detection - Completed 2026-07-08
- [x] Reverse Engineering - Completed 2026-07-08 (approved by user)
- [x] Requirements Analysis - Completed 2026-07-08 (approved by user)
- [x] User Stories - Completed 2026-07-08 (approved by user)
- [x] Workflow Planning - Completed 2026-07-08 (approved by user)
- [x] Application Design - Completed 2026-07-08 (approved by user)
  - **Artifacts Location**: aidlc-docs/inception/application-design/
- [x] Units Generation - Completed 2026-07-08 (approved by user) - INCEPTION PHASE COMPLETE
  - **Artifacts Location**: aidlc-docs/inception/application-design/unit-of-work*.md

## Units of Work (finalized)
1. Unit 1 - System Prompt Support (#81) - no dependencies
2. Unit 2 - Tool Use / Function Calling (#82) - soft-depends on Unit 1
3. Unit 3 - Prompt Caching (#83) - soft-depends on Unit 1
4. Unit 4 - Native Document Input (#84) - soft-depends on Unit 2 (shared ValidateLocalPath helper)
5. Unit 5 - Extended Thinking / Reasoning Mode (#85) - no dependencies

### Construction Phase - Unit 1 (System Prompt Support, #81)
- [x] Functional Design - SKIP (simple flag/config plumbing, no new business logic)
- [x] NFR Requirements - SKIP (no new security/performance concerns beyond cross-cutting NFRs already in requirements.md)
- [x] NFR Design - SKIP (follows from NFR Requirements skip)
- [x] Infrastructure Design - SKIP (no infrastructure in this project, decided globally)
- [x] Code Generation - Completed 2026-07-08 (approved by user, commit 4901a88)
  - Plan: aidlc-docs/construction/plans/unit-1-system-prompt-code-generation-plan.md
  - Summary: aidlc-docs/construction/unit-1-system-prompt/code/summary.md
- [ ] Build and Test - Pending all 5 units (this unit individually verified: make test/lint/coverage + integration tests all pass)

## Unit 1 Status: COMPLETE AND APPROVED

### Construction Phase - Unit 2 (Tool Use / Function Calling, #82)
- [x] Functional Design - Completed 2026-07-08 (approved by user)
  - Artifacts: aidlc-docs/construction/unit-2-tool-use/functional-design/
- [x] NFR Requirements - Completed 2026-07-08 (approved by user, combined with NFR Design)
  - Artifacts: aidlc-docs/construction/unit-2-tool-use/nfr-requirements/
- [x] NFR Design - Completed 2026-07-08 (approved by user, combined presentation)
  - Artifacts: aidlc-docs/construction/unit-2-tool-use/nfr-design/
- [x] Infrastructure Design - SKIP (no infrastructure in this project, decided globally)
- [x] Code Generation - Completed 2026-07-08, awaiting user review/approval
  - Plan: aidlc-docs/construction/plans/unit-2-tool-use-code-generation-plan.md
  - Summary: aidlc-docs/construction/unit-2-tool-use/code/summary.md
  - **Decision made**: --tools opt-in flag (default false), confirmed with user before generation
- [ ] Build and Test - Pending all 5 units (this unit individually verified: make test/lint/coverage + integration tests all pass; cmd coverage 8.0%->18.7%, new tools package 90%)

## Unit 2 Status: COMPLETE AND APPROVED (commit ad327a0)

## Unplanned Prerequisite: AWS SDK Upgrade (COMPLETE)
Discovered while starting Unit 3 that the pinned bedrockruntime SDK (v1.23.0) predates
prompt-caching support (needs v1.28.0+) and also lacks reasoning-content types needed by
Unit 5. User confirmed upgrading to latest (v1.55.0) now. Done - see
aidlc-docs/construction/sdk-upgrade/summary.md. go.mod's `go` directive moved 1.23.4 -> 1.24
(toolchain go1.24.7) as a side effect; README.md/CLAUDE.md updated to match. Full
build+test+lint+integration verification passed with no regressions.

### Construction Phase - Unit 3 (Prompt Caching, #83)
- [x] Functional Design - Completed 2026-07-08 (approved by user)
  - Artifacts: aidlc-docs/construction/unit-3-prompt-caching/functional-design/
- [x] NFR Requirements - SKIP (no new security surface, reliability concern already fully specified as a business rule)
- [x] NFR Design - SKIP (follows from NFR Requirements skip)
- [x] Infrastructure Design - SKIP (no infrastructure in this project, decided globally)
- [x] Code Generation - Completed 2026-07-08, awaiting user review/approval
  - Plan: aidlc-docs/construction/plans/unit-3-prompt-caching-code-generation-plan.md
  - Summary: aidlc-docs/construction/unit-3-prompt-caching/code/summary.md
- [ ] Build and Test - Pending all 5 units (this unit individually verified: make test/lint/coverage + integration tests all pass; cmd coverage 18.7%->22.0%)

## Unit 3 Status: COMPLETE AND APPROVED (commit e315d18)

### Construction Phase - Unit 4 (Native Document Input, #84)
- [x] Functional Design - Completed 2026-07-08 (approved by user)
  - Artifacts: aidlc-docs/construction/unit-4-document-input/functional-design/
- [x] NFR Requirements - Completed 2026-07-08 (approved by user, combined with NFR Design)
  - Artifacts: aidlc-docs/construction/unit-4-document-input/nfr-requirements/
- [x] NFR Design - Completed 2026-07-08 (approved by user, combined presentation)
  - Artifacts: aidlc-docs/construction/unit-4-document-input/nfr-design/
- [x] Infrastructure Design - SKIP (no infrastructure in this project, decided globally)
- [x] Code Generation - Completed 2026-07-08, awaiting user review/approval
  - Plan: aidlc-docs/construction/plans/unit-4-document-input-code-generation-plan.md
  - Summary: aidlc-docs/construction/unit-4-document-input/code/summary.md
- [ ] Build and Test - Pending all 5 units (this unit individually verified: make test/lint/coverage + integration tests all pass; utils coverage 44.7%->49.3%)

## Unit 4 Status: COMPLETE AND APPROVED (commit 6edbcce)

### Construction Phase - Unit 5 (Extended Thinking, #85) - FINAL UNIT
- [x] Functional Design - Completed 2026-07-08 (approved by user)
  - Artifacts: aidlc-docs/construction/unit-5-extended-thinking/functional-design/
  - **IMPORTANT CAVEAT**: the request-side JSON shape for enabling reasoning (AdditionalModelRequestFields) is UNVERIFIED - it's an untyped free-form field, assumed shape based on training knowledge not a live source. Highest-risk assumption in the initiative; flagged prominently to user.
- [x] NFR Requirements - SKIP (no new security surface; request-shape risk is functional, not security)
- [x] NFR Design - SKIP (follows from NFR Requirements skip)
- [x] Infrastructure Design - SKIP (no infrastructure in this project, decided globally)
- [x] Code Generation - Completed 2026-07-08, awaiting user review/approval
  - Plan: aidlc-docs/construction/plans/unit-5-extended-thinking-code-generation-plan.md
  - Summary: aidlc-docs/construction/unit-5-extended-thinking/code/summary.md
  - **CAVEAT CARRIED FORWARD**: reasoning_config request shape is unverified, top item to test with real credentials
- [x] Build and Test - Completed 2026-07-08, awaiting final approval
  - Artifacts: aidlc-docs/construction/build-and-test/
  - Fresh full test suite: all green, no regressions
  - 3 cross-unit composition scenarios executed manually, all passed (no panics, clean expected failures at the AWS-credentials boundary)
  - Consolidated risk list: 5 items need real-credential verification, ranked by priority (Unit 5's reasoning_config shape highest)

## Unit 5 Status: COMPLETE AND APPROVED (commit a2b0b63)

## ALL 5 UNITS COMPLETE: #81, #82, #83, #84, #85
## BUILD AND TEST: COMPLETE AND APPROVED
## INITIATIVE STATUS: COMPLETE

All commits on branch claude/ai-dlc-documentation-rl4e5s, pushed to origin.
No PR opened (not requested by user). Recommended before merge: run the
5 real-credential verification items in build-and-test-summary.md.

### Operations Phase
- [ ] Operations - PLACEHOLDER (not in scope)

## Recommended Unit Sequence (from execution-plan.md, to be finalized in Units Generation)
1. System Prompt (#81) - foundational
2. Tool Use (#82) - highest complexity
3. Prompt Caching (#83) - depends on System Prompt
4. Document Input (#84) - independent
5. Extended Thinking (#85) - independent

## GitHub Issues Filed (brainstorm session, 2026-07-08)
**Group 1 - Catch up to current Bedrock/Claude capabilities (IN SCOPE for current Requirements Analysis)**
- #81 Add system prompt support to chat and prompt commands
- #82 Add tool use / function calling support (Bedrock Converse API)
- #83 Add prompt caching support (cachePoint)
- #84 Add native document input (PDF/CSV/DOCX) via Converse document content blocks
- #85 Add extended thinking / reasoning mode support

**Group 2 - Agentic coding tool direction (logged, not started)**
- #86 Add built-in agent tools (file read/write, shell exec, git diff)
- #87 Add MCP client support
- #88 Support a project-context file convention (e.g. CHATCLI.md)

**Group 3 - UX modernization (logged, not started)**
- #89 Render markdown and code blocks in chat/prompt output
- #90 Add slash commands to the interactive chat loop

**Group 4 - Technical debt / fix what's there (logged, not started)**
- #91 models list ignores --region flag (bug)
- #92 Deduplicate model-validation logic across chat/prompt/image commands
- #93 Repository[T] interface mismatch with ChatRepository
- #94 Replace unmaintained github.com/satori/go.uuid dependency
- #95 Repo housekeeping (.goreleaser.yaml.backup, README Go version drift)
- #96 Add CI workflow to enforce lint/test on PRs

**Related pre-existing open issues surfaced during triage**: #58 (file attachments, relates to #84), #46 (document in chat mode), #41 (token counts, overlaps future UX idea - not re-filed), #65 (models placeholder output, relates to #91), #21 (concept of modules, relates to #81)

---

# INITIATIVE 1 EPILOGUE (post-completion events, outside AI-DLC stages)

- PR #97 opened for branch `claude/ai-dlc-documentation-rl4e5s` -> `main`, closing #81-#85. Merged 2026-07-08T15:07:16Z.
- Separately, PRs #99/#100/#101 (release automation, CI workflow, Homebrew deploy key) merged to `main`, closing #98. Not part of this AI-DLC initiative's scope.
- GitHub issue cleanup performed in conversation (not an AI-DLC stage, just tracker hygiene): #96 closed as resolved by #99's `ci.yml`; #95 closed as resolved (backup file gone, README Go version fixed). #91, #92, #93, #94, #58, #46 verified still open/unresolved on `main` and left open.
- Branch `claude/ai-dlc-documentation-rl4e5s` reset to latest `origin/main` (`d1619d2`) to start Initiative 2 fresh, per merged-branch restart protocol - old commit history for Initiative 1 is fully captured in `main` via PR #97.

---

# INITIATIVE 2: Universal Project-Context File Convention (#88, redefined)

## Project Information
- **Start Date**: 2026-07-08 (same day, continued session)
- **Trigger**: User requested Phase 2 discussion; picked #88 (project-context file), redefined scope from a chat-cli-specific `CHATCLI.md` to a universal `AGENTS.md`-first convention with fallback to other tools' conventions (`CLAUDE.md`, Cursor rules, Copilot instructions), per brainstorm discussion in this conversation.
- **Issue #88 updated** on GitHub with the new design (title + body rewritten) before starting this initiative.

## Stage Progress
- [x] Workspace Detection - Brownfield confirmed, existing `aidlc-docs/aidlc-state.md` found (Initiative 1, complete)
  - **Reverse engineering artifacts exist but are STALE** relative to current `main` (Initiative 1 added `tools/` package, `cmd/systemprompt.go`, `cmd/promptcache.go`, `cmd/documentinput.go`, `cmd/reasoning.go`, `cmd/toolloop.go`, SDK upgrade to v1.55.0; separately `main` gained CI/release automation). **Decision**: do NOT re-run full Reverse Engineering for this narrow, additive feature - scope is well understood from this session's own recent work (system prompt + config precedence pattern directly reused). Full RE re-run would be disproportionate to a single-feature initiative. Proceeding directly to Requirements Analysis with current-state knowledge loaded ad hoc from `cmd/systemprompt.go`, `cmd/config.go`, `cmd/promptcache.go`.
- [x] Requirements Analysis - Completed 2026-07-08, awaiting user approval
  - Clarifying questions answered by user (chat-only scope, automatic activation, explicit-system-wins precedence, Cursor dropped from scope, walk-up-to-.git left to AI judgment)
  - Artifacts: aidlc-docs/inception/requirements/agents-md-convention-questions.md, aidlc-docs/inception/requirements/agents-md-convention-requirements.md
  - FR1-FR6 + NFR1-NFR5 documented; Out of Scope section captures prompt-command support, Cursor's real .mdc convention, and README.md-as-default as explicit non-goals for this pass
- [x] User Stories - SKIPPED (approved by user via "Approve and continue" without requesting the stage be added)
- [x] Workflow Planning - Completed 2026-07-08, awaiting user approval
  - Artifacts: aidlc-docs/inception/plans/agents-md-convention-execution-plan.md
  - Risk: Low. Application Design SKIP, Units Generation SKIP (single unit, no new subsystem shape). Functional Design + NFR combined EXECUTE. Code Generation + Build and Test ALWAYS EXECUTE.
- [x] Application Design - SKIP (see execution plan rationale)
- [x] Units Generation - SKIP (see execution plan rationale, this initiative proceeds as a single implicit unit straight into Construction) - INCEPTION PHASE COMPLETE

### Construction Phase - agents-md-convention (#88, single implicit unit)
- [x] Functional Design - Completed 2026-07-08, awaiting user approval
  - Artifacts: aidlc-docs/construction/agents-md-convention/functional-design/{business-logic-model,business-rules,domain-entities}.md
  - Resolved FR1.2's walk-up rule into a concrete two-phase algorithm (Phase A: cheap `.git`-boundary stat-walk; Phase B: check candidates at cwd, then boundary dir only) - verified against actual cmd/chat.go:114, cmd/systemprompt.go, cmd/config.go source, not guessed
- [x] NFR Requirements + Design - Completed 2026-07-08 (combined presentation, same pattern as Initiative 1 Units 2/4), awaiting user approval
  - Artifacts: aidlc-docs/construction/agents-md-convention/nfr-requirements/nfr-requirements-and-design.md
  - Security: bounded to 2 directories, no path-traversal surface (fixed filenames only, no user-supplied path). Reliability: filesystem failures degrade to "no match," never fatal.
- [x] Infrastructure Design - SKIP (no infrastructure in this project, decided globally)
- [x] Code Generation - Completed 2026-07-08, awaiting user review/approval
  - Plan: aidlc-docs/construction/plans/agents-md-convention-code-generation-plan.md (all 14 steps complete)
  - Summary: aidlc-docs/construction/agents-md-convention/code/summary.md
  - cmd coverage 23.6% -> 31.8%, total 66.3% -> 67.8%, no regressions. All 7 integration tests pass. Manual smoke test confirmed discovery, --system suppression, and --no-context-file suppression all work end-to-end.
- [x] Build and Test - Completed 2026-07-08, awaiting final approval
  - Artifacts: aidlc-docs/construction/build-and-test/agents-md-convention-summary.md
  - Build success, all unit/integration tests pass, no coverage regression, manual smoke test against the compiled binary confirmed discovery/--system-suppression/--no-context-file-suppression all work end-to-end. No new real-credential-verification surface (feature never touches Bedrock directly - reuses Initiative 1's existing system-prompt/cache-point pipeline as-is).

## Unit "agents-md-convention" Status: COMPLETE AND APPROVED (commit c1bb745)
## INITIATIVE 2 STATUS: COMPLETE - PR #102 merged (commit 955130f on main), #88 closed

### Initiative 2 Epilogue
User tested the merged code and pushed one follow-up fix commit (768c9f1, with Cursor)
before/at merge time: fixed a real bug where `resolveContextFilenames` couldn't
distinguish "context-files config key unset" from "explicitly set to empty string" -
both looked like the same empty input, so the documented disable-via-empty-config
mechanism (FR5.2/BR12) silently didn't work. Fixed via a new `FileManager.IsConfigSet`
plus a `configSet bool` parameter. Also improved notice/warning messages to show
cwd-relative paths instead of full absolute paths, and hardened symlink handling.
Branch `claude/ai-dlc-documentation-rl4e5s` reset to latest `origin/main` (955130f)
to start Initiative 3 fresh; remote branch had been auto-deleted after merge, recreated
via a plain push after `git remote prune origin` (confirmed safe: old branch head was
a verified ancestor of the merge commit before reset).

---

# INITIATIVE 3: Built-in Agent Tools (#86)

## Project Information
- **Start Date**: 2026-07-08 (same day, continued session)
- **Trigger**: User picked #86 next from the Group 2 (agentic direction) backlog after Initiative 2 (#88) merged.
- **Scope per issue #86**: "Once tool use is supported [done, #82], add a small first-party toolset — read file [already done, #82], write file, run a shell command, git diff — so chat-cli chat can act as a lightweight, Bedrock-native coding assistant directly in the terminal... Needs careful scoping around confirmation prompts before destructive actions (file writes, shell exec)."
- **Risk profile note**: Higher than Initiatives 1-2 - `run_shell` is arbitrary command execution, `write_file` is a destructive filesystem action. This will need real Security NFR treatment, not a skip, and likely a fuller Inception treatment (User Stories, possibly Application Design) given the confirmation-flow UX design surface.

## Stage Progress
- [x] Workspace Detection - Brownfield confirmed, reusing existing `aidlc-docs/aidlc-state.md` context (no re-run of Reverse Engineering - same rationale as Initiative 2, narrow addition to an already-understood `tools/` subsystem from Initiative 1 Unit 2)
- [ ] Requirements Analysis - IN PROGRESS, clarifying questions issued
