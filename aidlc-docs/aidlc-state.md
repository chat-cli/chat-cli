# AI-DLC State Tracking

## Project Information
- **Project Type**: Brownfield
- **Start Date**: 2026-07-08T00:00:00Z
- **Current Stage**: INCEPTION - User Stories (Generation complete, awaiting review/approval)

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
- [x] User Stories - Plan approved, stories/personas generated 2026-07-08, awaiting review/approval
- [ ] Workflow Planning

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
