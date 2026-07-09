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

---

# Initiative 3 Story Map (#86)

| Story | Unit | FR/NFR IDs | Issue |
|---|---|---|---|
| 8.1 — First-time confirmation for a destructive call | Unit 6 — Confirmation Engine | FR5.1-FR5.2 | #86 |
| 8.2 — Sticky approval is remembered and reused | Unit 6 — Confirmation Engine | FR5.2, FR6.1-FR6.3, FR7.1-FR7.2 | #86 |
| 8.3 — Denying a destructive call | Unit 6 — Confirmation Engine | FR5.2-FR5.3 | #86 |
| 7.1 — Model edits a local file | Unit 7 — New Tools | FR2.1-FR2.3, NFR1 | #86 |
| 7.2 — Model runs a shell command | Unit 7 — New Tools | FR3.1-FR3.2, NFR2 | #86 |
| 7.3 — Model inspects the working tree diff | Unit 7 — New Tools | FR4.1-FR4.3 | #86 |
| 6.1 — Tool use works without any flag | Unit 8 — Automatic Enablement | FR1.1, FR1.4 | #86 |
| 6.2 — Graceful degradation for models that reject tool use | Unit 8 — Automatic Enablement | FR1.2-FR1.3 | #86 |

## Coverage Check
- **Total stories**: 8
- **Stories assigned**: 8
- **Unassigned stories**: 0
- **Units with no stories**: 0

Cross-cutting NFRs (NFR3 revised backward compatibility, NFR4 TDD/coverage, NFR5 prompt usability) apply to all 8 stories/3 units and are re-verified per-unit during Build and Test.
