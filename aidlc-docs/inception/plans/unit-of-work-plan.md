# Unit of Work Plan

**Context**: chat-cli is a monolith (single deployable binary), not microservices. Per `units-generation.md`'s definition, "the single unit represents the entire application with logical modules" — here, the 5 epics from `stories.md` are the natural logical-module boundaries, since Application Design already scoped genuinely new components (the `tools` package) around exactly one of them.

## Decomposition Approach

**Grouping strategy**: One unit per epic/GitHub issue (feature-based, consistent with the breakdown already used in `requirements.md` and `stories.md`). 5 units total:

| Unit | Stories | Issue | New Components Touched (from Application Design) |
|---|---|---|---|
| Unit 1 — System Prompt | 1.1 | #81 | None (flag/config addition only) |
| Unit 2 — Tool Use | 2.1, 2.2 | #82 | `tools.Tool`, `tools.Registry`, `tools.ReadFileTool`, `utils.ValidateLocalPath` |
| Unit 3 — Prompt Caching | 3.1 | #83 | Cache-point helper (`utils`) |
| Unit 4 — Document Input | 4.1 | #84 | `utils.ValidateLocalPath` (shared with Unit 2), `utils.ReadDocument` |
| Unit 5 — Extended Thinking | 5.1 | #85 | None (flag/field addition only) |

## Assumptions Stated (in place of a Q&A round)

Same calibration used throughout this session:

1. **Story grouping**: 1:1 with the epics already established — no further splitting or merging. Each unit is independently small enough for one Code Generation pass.
2. **Dependencies are soft/sequencing hints, not hard blockers**: Since this is solo, sequential AI-DLC execution (not parallel teams), "depends on" below means recommended implementation order, not a technical blocking dependency requiring a merged PR first.
   - Unit 2 (Tool Use) recommended after Unit 1 (System Prompt) — a system prompt is the natural place to instruct the model about available tools, though not strictly required to compile/test Unit 2 in isolation.
   - Unit 3 (Prompt Caching) recommended after Unit 1 — needs a system prompt to exist to exercise the "cache after system prompt" path (FR3.1); the "cache after document" path (FR3.2) can be tested independently with a stubbed document, so Unit 3 doesn't strictly require Unit 4 first.
   - `utils.ValidateLocalPath` is introduced in whichever of Unit 2 / Unit 4 lands first; the second one reuses it rather than re-defining it.
   - Units 1 and 5 have no dependencies on any other unit.
3. **Team alignment / ownership**: N/A — solo maintainer, no team-boundary concerns.
4. **Technical/deployment considerations**: N/A — single binary, no per-unit deployment differences; all 5 units ship in the same binary/release.
5. **Business domain boundaries**: N/A — single CLI application domain, no bounded-context splitting needed.
6. **Code organization (brownfield, not greenfield)**: No new top-level directory structure beyond the one new `tools/` package already identified in Application Design; everything else extends existing files in place.

## Recommended Sequence for Construction (Per-Unit Loop)

1. **Unit 1 — System Prompt** (#81) — foundational, lowest risk, no new components
2. **Unit 2 — Tool Use** (#82) — highest complexity, benefits from Unit 1 existing
3. **Unit 3 — Prompt Caching** (#83) — benefits from Unit 1 existing
4. **Unit 4 — Document Input** (#84) — independent; reuses `ValidateLocalPath` from Unit 2 if it landed first, otherwise introduces it
5. **Unit 5 — Extended Thinking** (#85) — independent, smallest surface area, natural last unit

This matches the sequence already recommended in `requirements.md` and `execution-plan.md` — Units Generation confirms it at the unit-of-work level rather than changing it.

## Mandatory Artifacts (Part 2, pending this plan's approval)
- [x] `aidlc-docs/inception/application-design/unit-of-work.md`
- [x] `aidlc-docs/inception/application-design/unit-of-work-dependency.md`
- [x] `aidlc-docs/inception/application-design/unit-of-work-story-map.md`
