# AI-DLC State Tracking

## Project Information
- **Project Type**: Brownfield
- **Start Date**: 2026-07-08T00:00:00Z
- **Current Stage**: CONSTRUCTION - Unit 3 (Prompt Caching, #83) - Code Generation complete, awaiting approval

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

## Unit 3 Status: CODE COMPLETE, AWAITING REVIEW
Next: Unit 4 (Native Document Input, #84) once Unit 3 is approved.

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
