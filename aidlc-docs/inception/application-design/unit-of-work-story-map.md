# Unit of Work Story Map

Every story from `aidlc-docs/inception/user-stories/stories.md` is assigned to exactly one unit — no orphaned or duplicated stories.

| Story | Unit | FR/NFR IDs | Issue |
|---|---|---|---|
| 1.1 — Set a system prompt for a session | Unit 1 — System Prompt | FR1.1-FR1.4 | #81 |
| 2.1 — Model calls a registered tool mid-conversation | Unit 2 — Tool Use | FR2.1-FR2.4 | #82 |
| 2.2 — Use the built-in `read_file` tool | Unit 2 — Tool Use | FR2.5, NFR3 | #82 |
| 3.1 — Automatic prompt caching with graceful fallback | Unit 3 — Prompt Caching | FR3.1-FR3.4 | #83 |
| 4.1 — Attach a non-image document to a prompt | Unit 4 — Document Input | FR4.1-FR4.4 | #84 |
| 5.1 — Enable extended thinking and see the reasoning distinctly | Unit 5 — Extended Thinking | FR5.1-FR5.4 | #85 |

## Coverage Check
- **Total stories**: 6
- **Stories assigned**: 6
- **Unassigned stories**: 0
- **Units with no stories**: 0

Cross-cutting NFRs (NFR1 backward compatibility, NFR2 TDD/coverage, NFR4 error handling, NFR5 no new deps, NFR6 docs, NFR7 streaming compatibility) apply to all 6 stories/5 units and are re-verified per-unit during Build and Test, not owned by any single unit.
