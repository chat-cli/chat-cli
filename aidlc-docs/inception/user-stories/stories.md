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

## Traceability Summary (Initiative 1, #81-#85)

| Story | Epic | FR/NFR IDs | Issue |
|---|---|---|---|
| 1.1 | System Prompt | FR1.1-FR1.4 | #81 |
| 2.1 | Tool Use (plumbing) | FR2.1-FR2.4 | #82 |
| 2.2 | Tool Use (`read_file`) | FR2.5, NFR3 | #82 |
| 3.1 | Prompt Caching | FR3.1-FR3.4 | #83 |
| 4.1 | Document Input | FR4.1-FR4.4 | #84 |
| 5.1 | Extended Thinking | FR5.1-FR5.4 | #85 |

NFR1 (backward compatibility), NFR2 (TDD/coverage), NFR4 (error handling), NFR5 (no new deps), NFR6 (docs), NFR7 (streaming compatibility) apply across all 6 stories and are re-asserted per-story implicitly by "behavior unchanged when flag not set" acceptance criteria.

---

# Initiative 3 — Built-in Agent Tools (#86)

FR/NFR IDs below reference `aidlc-docs/inception/requirements/builtin-tools-requirements.md` (a separate numbering space from Initiative 1's FR1-FR5 above - disambiguated here as "Epic 6/7/8" to avoid confusion with Epics 1-5).

## Epic 6 — Automatic Tool-Use Enablement (FR1, [#86](https://github.com/chat-cli/chat-cli/issues/86))

### Story 6.1 — Tool use works without any flag
**As** the chat-cli User, **I want** `chat` to make tools available to the model automatically, **so that** I don't have to remember and pass `--tools` every session.

**Acceptance Criteria**:
- Given I start `chat-cli` with no special flags, when a request is sent, then it includes a non-empty `ToolConfiguration` built from the registry (now `read_file`, `write_file`, `run_shell`, `git_diff`). *(FR1.1)*
- Given the `--tools` flag from Initiative 1 no longer exists, when I run `chat-cli --help`, then `--tools` is absent from the flag list. *(FR1.4)*

**INVEST**: Small, independent - pure wiring change, testable by asserting the request always carries a `ToolConfiguration`.

### Story 6.2 — Graceful degradation for models that reject tool use
**As** the chat-cli User, **I want** `chat` to keep working even on a model that doesn't support tool use, **so that** the automatic-enablement change (Story 6.1) never breaks a session outright.

**Acceptance Criteria**:
- Given a model/request rejects the `ToolConfiguration` field, when that failure occurs, then chat-cli automatically retries the same request once without it, mirroring the existing cache-point retry pattern from #83. *(FR1.2)*
- Given the retry-without-tools succeeds, when the response streams back to me, then I see a clear, one-line notice that tools were disabled for this session (not a silent change in behavior). *(FR1.3)*
- Given tools were disabled this way, when I continue the conversation, then no further requests in that session re-attempt the `ToolConfiguration` (no repeated failed round-trips). *(FR1.2, FR1.3)*

**INVEST**: Independent of Story 6.1's happy path but depends on it existing; testable with a mocked Converse call that rejects the tool field on the first attempt and succeeds on retry.

---

## Epic 7 — New Built-in Tools (FR2-FR4, [#86](https://github.com/chat-cli/chat-cli/issues/86))

### Story 7.1 — Model edits a local file
**As** the chat-cli User, **I want** the model to be able to create or overwrite a file in my project when I've approved it, **so that** it can act like a lightweight coding assistant instead of only describing what I should change myself.

**Acceptance Criteria**:
- Given the model requests `write_file` with a path and content, when the path resolves within the working directory (via `utils.ValidateLocalPath`), then, after approval (Epic 8), the file is created or overwritten with that content. *(FR2.1, FR2.2)*
- Given the model requests a path outside the working directory, when chat-cli validates it, then it refuses and returns an error result to the model, the same as `read_file`'s existing traversal protection. *(FR2.1, NFR1)*
- Given a `write_file` call has not yet been approved, when chat-cli processes it, then no write happens until the confirmation gate (Epic 8) resolves. *(FR2.3)*

**INVEST**: Small, depends on Epic 8's gate existing but is separately testable (path validation and file-write logic can be tested independently of the approval UI).

### Story 7.2 — Model runs a shell command
**As** the chat-cli User, **I want** the model to be able to run a shell command when I've approved it, **so that** it can do things like run tests, install a dependency, or check tool versions without me leaving the conversation.

**Acceptance Criteria**:
- Given the model requests `run_shell` with a command string, when it's approved (Epic 8), then chat-cli executes it via the shell in chat-cli's own working directory and returns combined, size-capped stdout+stderr to the model. *(FR3.1)*
- Given the command runs longer than the fixed timeout, when that timeout elapses, then chat-cli terminates it and returns a clear timeout result to the model rather than hanging the session. *(FR3.1, NFR2)*
- Given a `run_shell` call has not yet been approved, when chat-cli processes it, then no command executes until the confirmation gate (Epic 8) resolves. *(FR3.2)*

**INVEST**: Small, mirrors Story 7.1's shape (a destructive tool gated by Epic 8) but independently testable (command execution/timeout/truncation logic vs. approval flow).

### Story 7.3 — Model inspects the working tree diff
**As** the chat-cli User, **I want** the model to be able to see `git diff` output, **so that** it can reason about my current uncommitted changes without me manually pasting them in.

**Acceptance Criteria**:
- Given the model requests `git_diff` with no argument, when chat-cli runs it, then it returns the plain `git diff` output for the current working directory. *(FR4.1)*
- Given the model requests `git_diff` with a path or ref argument, when chat-cli runs it, then it returns `git diff <arg>` output instead. *(FR4.1)*
- Given the working directory isn't inside a git repository, when `git_diff` is requested, then chat-cli returns a clear error result to the model, not a fatal CLI crash. *(FR4.3)*
- Given `git_diff` is read-only, when the model requests it, then it executes immediately with no confirmation gate. *(FR4.2)*

**INVEST**: Small, fully independent of Epic 8 (no gate involved) - the simplest story in this initiative.

---

## Epic 8 — Confirmation and Sticky Approval (FR5-FR7, [#86](https://github.com/chat-cli/chat-cli/issues/86))

### Story 8.1 — First-time confirmation for a destructive call
**As** the chat-cli User, **I want** to see exactly what a destructive tool call will do before it happens, **so that** I stay in control of file writes and shell commands.

**Acceptance Criteria**:
- Given the model requests `write_file`, when no existing approval matches, then I'm shown the target path and the content that will be written, and asked to decide before anything happens. *(FR5.1)*
- Given the model requests `run_shell`, when no existing approval matches, then I'm shown the exact command string and asked to decide before anything runs. *(FR5.1)*
- Given I'm shown the prompt, when I respond, then my options are exactly: approve once, approve for this session, approve always, or deny. *(FR5.2)*

**INVEST**: Independent of Stories 8.2/8.3 (this is the "no prior approval exists" path); testable by asserting the prompt content and that execution blocks until a decision is made.

### Story 8.2 — Sticky approval is remembered and reused
**As** the chat-cli User, **I want** an "always" or "this session" approval to actually skip future prompts for matching calls, **so that** I'm not re-asked the same thing repeatedly.

**Acceptance Criteria**:
- Given I approved a `run_shell` call "for this session," when the model later requests another `run_shell` call with the same base command, then it executes without a new prompt, for the remainder of that session only. *(FR5.2, FR6.1, FR6.3, FR7.2)*
- Given I approved a `write_file` call "always," when the model later requests another `write_file` call under the same directory (in a later `chat` session, same repository), then it executes without a new prompt. *(FR5.2, FR6.2, FR6.3, FR7.1)*
- Given an "always" approval was granted in one repository, when I run `chat-cli` from a different, unrelated repository, then that approval does not apply there - the gate prompts again. *(FR7.1)*

**INVEST**: Depends on Story 8.1's gate existing, but independently testable against the pattern-matching and storage logic directly (no need to drive the actual prompt UI to test whether a stored pattern matches).

### Story 8.3 — Denying a destructive call
**As** the chat-cli User, **I want** to be able to say no to a specific destructive call, **so that** the model can't force an action I don't want.

**Acceptance Criteria**:
- Given I'm shown the confirmation prompt, when I choose deny, then the action does not execute, and the model receives an error `ToolResultBlock` indicating I declined, not a fatal error. *(FR5.3)*
- Given I denied one call, when the model tries the exact same or a similar call again later in the conversation, then I'm prompted again (a denial is not itself sticky). *(FR5.2, FR5.3)*

**INVEST**: Small, independent - the "unhappy path" companion to Story 8.1.

---

## Traceability Summary (Initiative 3, #86)

| Story | Epic | FR/NFR IDs | Issue |
|---|---|---|---|
| 6.1 | Automatic Enablement | FR1.1, FR1.4 | #86 |
| 6.2 | Graceful Degradation | FR1.2, FR1.3 | #86 |
| 7.1 | `write_file` | FR2.1-FR2.3, NFR1 | #86 |
| 7.2 | `run_shell` | FR3.1-FR3.2, NFR2 | #86 |
| 7.3 | `git_diff` | FR4.1-FR4.3 | #86 |
| 8.1 | First-time confirmation | FR5.1-FR5.2 | #86 |
| 8.2 | Sticky approval reuse | FR5.2, FR6.1-FR6.3, FR7.1-FR7.2 | #86 |
| 8.3 | Denial | FR5.2-FR5.3 | #86 |

NFR3 (revised backward compatibility - intentional default-behavior change), NFR4 (TDD/coverage), NFR5 (usability of the confirmation prompt) apply across all 8 stories in this initiative.
