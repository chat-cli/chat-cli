# Story Generation Plan: Built-in Agent Tools (#86)

## Methodology Decisions (stated as assumptions, consistent with Initiative 1's calibration - not reopening a second Q&A round given how much was already resolved in Requirements Analysis)

- **Persona**: Reuse the single existing "chat-cli User" persona from Initiative 1/2 (`aidlc-docs/inception/user-stories/personas.md`), extended with this initiative's specific goals/pain-points rather than creating a new persona file - this remains a single-user-type terminal tool (per the standing assessment already made twice).
- **Breakdown approach**: Feature-based, one epic per FR group from `builtin-tools-requirements.md` (FR1 automatic enablement, FR2-4 the three tools, FR5-7 the confirmation/approval system) - same approach as Initiative 1's `stories.md`.
- **Story format**: `As <persona>, I want <capability>, so that <benefit>` with Given/When/Then acceptance criteria tagged back to specific FR numbers, plus an INVEST note per story - identical format to the existing `stories.md`.
- **Granularity**: The approval-tier scenarios (once/session/always/deny/auto-disable) are broken into their own stories rather than folded as acceptance-criteria bullets under one giant "confirmation" story, given the User Stories Assessment's specific reasoning for executing this stage at all (making each branch of that decision tree independently testable).

## Steps
- [ ] Extend `aidlc-docs/inception/user-stories/personas.md` with this initiative's goals/pain-points (same file, additive - not a new persona)
- [ ] Generate `aidlc-docs/inception/user-stories/stories.md` additions: Epic 6 (Automatic Tool-Use Enablement, FR1), Epic 7 (write_file/run_shell/git_diff, FR2-4), Epic 8 (Confirmation and Sticky Approval, FR5-7)
- [ ] Verify every FR (FR1.1-FR7.2) is traceable to at least one story's acceptance criteria
- [ ] Verify each story is independently testable per INVEST
