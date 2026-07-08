# User Stories

Persona: **The chat-cli User** (see `personas.md`). Breakdown: Feature-based, one epic per FR group in `aidlc-docs/inception/requirements/requirements.md`. Each story is Independent, Negotiable, Valuable, Estimable, Small, and Testable (INVEST) — noted per story.

---

## Epic 1 — System Prompt Support (FR1, [#81](https://github.com/chat-cli/chat-cli/issues/81))

### Story 1.1 — Set a system prompt for a session
**As** the chat-cli User, **I want** to provide a system prompt for a `chat` or `prompt` invocation, **so that** I can give the model a persona or standing instructions without repeating them in every message.

**Acceptance Criteria**:
- Given no `--system` flag and no `system-prompt` config value set, when I run `chat-cli prompt "hi"` or start `chat-cli`, then the request is sent exactly as it is today (no `SystemContentBlocks`). *(FR1.4)*
- Given I pass `--system "You are a terse assistant"`, when I send any message in that session, then the model's request includes that text as `SystemContentBlocks`. *(FR1.1, FR1.3)*
- Given I run `chat-cli config set system-prompt "You are a terse assistant"`, when I later run `chat-cli` or `chat-cli prompt` without `--system`, then the persisted value is used, following the existing flag → config → default precedence. *(FR1.2)*
- Given both a `--system` flag and a persisted config value are present, when I run a command, then the flag value wins (same precedence rule as `--model-id`). *(FR1.2)*

**INVEST**: Small, independent of the other 4 epics, testable via unit tests on the flag/config precedence path (mirrors existing `GetConfigValue` tests).

---

## Epic 2 — Tool Use / Function Calling (FR2, [#82](https://github.com/chat-cli/chat-cli/issues/82))

### Story 2.1 — Model calls a registered tool mid-conversation
**As** the chat-cli User, **I want** the model to be able to invoke a tool during `chat` and see the result folded back into the conversation, **so that** I get answers that require more than the model's own knowledge (e.g. reading a local file).

**Acceptance Criteria**:
- Given at least one tool is registered, when I start `chat-cli`, then the request includes a `ToolConfig` describing the registered tool(s). *(FR2.1)*
- Given the model responds with a `ContentBlockMemberToolUse` block, when chat-cli receives it, then it executes the matching registered tool with the model-supplied input and sends a `ToolResultBlock` back, continuing the conversation without me having to do anything. *(FR2.2)*
- Given the model requests a tool name that isn't registered, when chat-cli processes the response, then it returns an error `ToolResultBlock` to the model (not a fatal CLI crash), and the model can recover or explain the issue to me. *(FR2.3)*
- Given a registered tool's execution fails (e.g. an error), when that happens, then the error is returned to the model as an error `ToolResultBlock`, not surfaced as an uncaught crash. *(FR2.3)*
- Given a tool-call round-trip happened during a turn, when I later run `chat-cli chat list` or resume with `--chat-id`, then the turn is present in history the same way a normal turn is. *(FR2.4)*

**INVEST**: Independent of Epics 1/3/4/5 (though it benefits from Epic 1 for tool-usage instructions); the largest/most complex story in this set — still small enough to test with a mocked tool registry and a fake tool-use response.

### Story 2.2 — Use the built-in `read_file` tool
**As** the chat-cli User, **I want** a built-in tool that lets the model read a local file when it decides it needs to, **so that** I can ask questions about a file without manually piping it in first.

**Acceptance Criteria**:
- Given the `read_file` tool is registered by default, when the model requests it with a path, then chat-cli reads that file and returns its contents as the tool result — using the same working-directory-confined, path-traversal-safe validation as `utils.ReadImage`. *(FR2.5, NFR3)*
- Given the model requests a path outside the working directory (e.g. `../../etc/passwd`), when chat-cli validates it, then it refuses and returns an error result to the model rather than reading the file. *(FR2.5, NFR3)*

**INVEST**: Small, depends on Story 2.1's plumbing but is separately testable and separately shippable.

---

## Epic 3 — Prompt Caching (FR3, [#83](https://github.com/chat-cli/chat-cli/issues/83))

### Story 3.1 — Automatic prompt caching with graceful fallback
**As** the chat-cli User, **I want** repeated system prompts and piped-in documents to be cached automatically, **so that** my sessions are cheaper and faster without me having to configure anything.

**Acceptance Criteria**:
- Given a system prompt is set (Epic 1), when a request is sent, then a cache checkpoint is inserted immediately after it. *(FR3.1)*
- Given a document was piped in via stdin, when a request is sent, then a cache checkpoint is inserted immediately after the document content. *(FR3.2)*
- Given the selected model doesn't support prompt caching, when the API rejects the cache point, then chat-cli automatically retries once without it and logs a non-fatal warning — my request still succeeds. *(FR3.3)*
- Given a response includes cache token metrics, when it's received, then those metrics are available internally for a future display feature (issue #41) — no visible output changes required in this story. *(FR3.4)*

**INVEST**: Independent of the other epics; testable by asserting the request shape (cache point presence) and by simulating a rejection response to verify the fallback retry.

---

## Epic 4 — Native Document Input (FR4, [#84](https://github.com/chat-cli/chat-cli/issues/84))

### Story 4.1 — Attach a non-image document to a prompt
**As** the chat-cli User, **I want** to attach a PDF, CSV, or similar document directly (not just images), **so that** I can ask questions about real working documents, not just pictures.

**Acceptance Criteria**:
- Given I run `chat-cli prompt "summarize this" --document report.pdf`, when the model supports document input, then the file is sent as a `ContentBlockMemberDocument`, not stuffed into the prompt text as raw bytes. *(FR4.1)*
- Given the file path resolves outside the working directory, or the extension isn't in the supported list (PDF, CSV, DOC/DOCX, XLS/XLSX, HTML, TXT, MD), when I run the command, then chat-cli rejects it with a clear error before calling Bedrock. *(FR4.2)*
- Given the selected model doesn't support document input, when I use `--document`, then I get a clear, specific error rather than a cryptic API failure. *(FR4.3)*
- Given I use both `--image` and `--document` together on a model that supports both, when I run the command, then both attachments are sent and `--image` behavior is completely unchanged from today. *(FR4.4)*

**INVEST**: Independent of all other epics; testable via the same pattern as existing `utils.ReadImage` tests, generalized to more extensions.

---

## Epic 5 — Extended Thinking / Reasoning Mode (FR5, [#85](https://github.com/chat-cli/chat-cli/issues/85))

### Story 5.1 — Enable extended thinking and see the reasoning distinctly
**As** the chat-cli User, **I want** to turn on extended thinking for models that support it and see the reasoning separated from the final answer, **so that** I can understand how the model arrived at a complex answer without it being mixed into the response text.

**Acceptance Criteria**:
- Given I don't pass `--thinking`, when I run `chat` or `prompt`, then no `AdditionalModelRequestFields` are sent and behavior is unchanged. *(FR5.3)*
- Given I pass `--thinking` on a supported model, when a response includes reasoning content, then it's printed visually distinct from the final answer (e.g. dimmed/prefixed), consistent with the existing gray-echo styling already used for user input in `chat.go`. *(FR5.1, FR5.2)*
- Given I pass `--thinking` on a model that rejects the field, when the API call fails, then I see a clear error message identifying the cause, not a generic failure. *(FR5.4)*

**INVEST**: Independent of all other epics; testable by asserting request shape when the flag is set/unset and by asserting output formatting given a mocked reasoning content block.

---

## Traceability Summary

| Story | Epic | FR/NFR IDs | Issue |
|---|---|---|---|
| 1.1 | System Prompt | FR1.1-FR1.4 | #81 |
| 2.1 | Tool Use (plumbing) | FR2.1-FR2.4 | #82 |
| 2.2 | Tool Use (`read_file`) | FR2.5, NFR3 | #82 |
| 3.1 | Prompt Caching | FR3.1-FR3.4 | #83 |
| 4.1 | Document Input | FR4.1-FR4.4 | #84 |
| 5.1 | Extended Thinking | FR5.1-FR5.4 | #85 |

NFR1 (backward compatibility), NFR2 (TDD/coverage), NFR4 (error handling), NFR5 (no new deps), NFR6 (docs), NFR7 (streaming compatibility) apply across all 6 stories and are re-asserted per-story implicitly by "behavior unchanged when flag not set" acceptance criteria.
