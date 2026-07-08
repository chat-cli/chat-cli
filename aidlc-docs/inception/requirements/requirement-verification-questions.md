# Requirements Clarification Questions

Your original request was to "document and understand what this project is all about in preparation to continue new development ideas." Reverse Engineering is now complete and approved. Before I can produce a requirements document, I need to know what you actually want to build or change next — please answer the questions below.

## Question 1
What type of work are you looking to do next?

A) Add a new user-facing feature or capability to chat-cli

B) Fix technical debt / bugs identified during the reverse-engineering review (region flag bug, duplicated model-validation logic, unimplemented `Repository[T]` methods, stray backup file, unmaintained `go.uuid` dependency, etc.)

C) Non-functional improvement (raise test coverage, add CI, refactor without behavior change)

D) Not sure yet — I'd like suggestions based on what the reverse-engineering review found

E) Other (please describe after [Answer]: tag below)

[Answer]: A (answered conversationally - see below)

**Resolution (2026-07-08, conversational)**: User asked for a brainstorm session, which surfaced 4 idea groups (issues #81-96 filed in GitHub for all of them). User chose to start with group 1, "Catch up to current Bedrock/Claude capabilities" — 5 new user-facing features: system prompt support (#81), tool use/function calling (#82), prompt caching (#83), native document input (#84), extended thinking (#85). Groups 2-4 (agentic tools, UX modernization, technical debt) are logged as issues for later.

## Question 2
Do you already have a specific idea in mind?

A) Yes — I'll describe it in the Other field below

B) No — please propose a shortlist of candidate ideas based on the codebase review, and I'll pick from those

X) Other (please describe after [Answer]: tag below)

[Answer]: B → then A (suggestions were proposed conversationally; user picked group 1, see #81-#85)

## Question 3
How much should we take on in this round?

A) One focused feature or fix

B) A small bundle of related quick wins (e.g., 2-3 of the technical debt items together)

C) A larger initiative spanning multiple components

X) Other (please describe after [Answer]: tag below)

[Answer]: C - five related capability upgrades (#81-#85) spanning cmd/chat.go, cmd/prompt.go, cmd/image.go, and utils

## Question 4
Are there any constraints I should design around (backward compatibility with existing `chat-cli` flags/config, specific AWS regions, timeline, must avoid new dependencies, etc.)? If none, say "none."

A) None — no special constraints

X) Other (please describe after [Answer]: tag below)

[Answer]: A - none explicitly stated; implicit constraints from CLAUDE.md apply: TDD required, don't break backward compatibility with existing flags/config, maintain/improve test coverage
