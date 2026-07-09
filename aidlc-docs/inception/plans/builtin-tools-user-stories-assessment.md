# User Stories Assessment: Built-in Agent Tools (#86)

## Request Analysis
- **Original Request**: Add `write_file`/`run_shell`/`git_diff` tools, plus (per the user's clarifying answers) replace the `--tools` opt-in flag with automatic enablement and a new destructive-action confirmation system with three-tier (once/session/always) pattern-based sticky approval.
- **User Impact**: Direct - this changes `chat`'s default behavior (tools now always attempted) and introduces an entirely new interaction pattern (confirmation prompts with a persistent choice) that users will encounter on essentially every session that uses a destructive tool.
- **Complexity Level**: Complex - new subsystem (permission engine), multiple new interaction flows, a revision to previously-established default behavior.
- **Stakeholders**: Single persona (the chat-cli User, per Initiative 1/2's established assessment - this remains a single-user-type terminal tool).

## Assessment Criteria Met
- [x] **High Priority**: New User Features (3 new tools) - any new functionality users directly interact with
- [x] **High Priority**: User Experience Changes - the automatic-tool-use + confirmation-gate change modifies the existing `chat` workflow non-trivially (Unit 2's `--tools` flag is being removed)
- [x] **Medium Priority**: Security Enhancements - the confirmation/approval system is functionally a permissions feature
- [x] **Complexity Factor**: Ambiguity - the three-tier approval UX (once/session/always) has several distinct scenarios (first destructive call, a repeat call matching a session approval, a repeat call matching a persisted approval, a denial, an auto-disable-on-model-rejection) that benefit from being made explicit and testable
- [x] **Complexity Factor**: Risk - destructive actions (file writes, arbitrary shell execution) carry real consequences if the approval/confirmation logic has a gap

## Decision
**Execute User Stories**: Yes

**Reasoning**: This is a meaningfully different case from Initiative 2 (a single, simple, automatic discovery mechanism with no new subsystem). Here, the number of distinct interaction scenarios (5+, listed above) each need a clear "given/when/then" so that: (a) implementation has an unambiguous spec for each approval-tier branch, (b) Build and Test has concrete scenarios to verify against, and (c) the security-sensitive paths (deny, auto-disable-on-rejection) are explicitly called out rather than left implicit in FR prose alone.

## Expected Outcomes
- Explicit acceptance criteria for every branch of the confirmation/approval decision tree, directly reusable as test-case scaffolding during Code Generation (same role stories played in Initiative 1).
- A clear separation between "read-only tool, no gate" stories and "destructive tool, full gate" stories, reducing the risk of the gate being accidentally applied inconsistently.
- Explicit coverage of the two "unhappy path" flows that most need to be right: a denied action, and the automatic tools-disabled-on-rejection degradation.
