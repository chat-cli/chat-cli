# AI-DLC State Tracking

## Project Information
- **Project Type**: Brownfield
- **Start Date**: 2026-07-08T00:00:00Z
- **Current Stage**: CONSTRUCTION - Unit 2 (Tool Use / Function Calling, #82) - Functional Design complete, awaiting approval

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
- [x] Functional Design - Complete 2026-07-08, awaiting approval
  - Artifacts: aidlc-docs/construction/unit-2-tool-use/functional-design/
- [ ] NFR Requirements - Next (EXECUTE: security-sensitive file access via read_file tool)
- [ ] NFR Design - Pending NFR Requirements
- [x] Infrastructure Design - SKIP (no infrastructure in this project, decided globally)
- [ ] Code Generation - Not started
- [ ] Build and Test - Pending all 5 units

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
