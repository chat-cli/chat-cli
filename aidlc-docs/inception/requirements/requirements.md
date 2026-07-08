# Requirements: Catch Up to Current Bedrock/Claude Capabilities

## Intent Analysis Summary

- **User Request**: Following a brainstorm session about how chat-cli should evolve given how much has changed in LLMs and agentic coding tools since it was last worked on, the user asked to log GitHub issues for all brainstormed ideas and then begin implementation of idea-group 1: "Catch up to current Bedrock/Claude capabilities."
- **Request Type**: New Feature (five related capability additions to existing commands)
- **Scope Estimate**: Multiple Components — touches `cmd/chat.go`, `cmd/prompt.go`, `config/config.go`, `utils/`, and likely introduces at least one new file (e.g. a tool registry)
- **Complexity Estimate**: Complex — tool use in particular introduces new control flow (multi-turn tool-call/tool-result round trips inside the streaming loop) that doesn't exist anywhere in the codebase today
- **Related GitHub Issues**: [#81](https://github.com/chat-cli/chat-cli/issues/81) system prompt, [#82](https://github.com/chat-cli/chat-cli/issues/82) tool use, [#83](https://github.com/chat-cli/chat-cli/issues/83) prompt caching, [#84](https://github.com/chat-cli/chat-cli/issues/84) document input, [#85](https://github.com/chat-cli/chat-cli/issues/85) extended thinking
- **Depth**: Standard — five distinct capabilities each need functional + non-functional coverage, but this is a single team working on an existing well-understood codebase, not a multi-stakeholder/high-risk system

## Assumptions Made (flag in "Request Changes" if any are wrong)

Rather than run another clarifying-question round on top of the one already completed this session, the following reasonable assumptions are made explicit here so they can be corrected during review instead of blocking progress:

1. **System prompt (#81)** applies to `chat` and `prompt` (not `image` — image prompts aren't conversational). For `chat`, it is fixed for the session at startup; a mid-session `/system` command is out of scope here (tracked separately as #90).
2. **Tool use (#82)** ships as an extensible tool-registry + execution loop in `chat` only (not `prompt`), plus one minimal, read-only, safe built-in tool (e.g. `read_file`, reusing the path-safety pattern from `utils.ReadImage`) so the feature is demoable and testable end-to-end without waiting on the fuller toolset in #86. Destructive tools (write/exec) stay out of scope — that's #86's job, which also needs a confirmation/safety design.
3. **Prompt caching (#83)** is applied automatically (no new flag) around the system prompt and/or piped document content when present, since Bedrock's `GetFoundationModel` doesn't expose a "supports caching" capability flag the way it does for modalities — the implementation should attempt a cache point and fail gracefully (log a warning, continue without caching) if the API rejects it for a given model.
4. **Document input (#84)** adds a new flag (`--document`/`-d`) alongside the existing `--image` flag on `prompt` (not replacing it, and not extended to `chat` in this pass) — image and document are distinct Bedrock content block types with different format/size rules.
5. **Extended thinking (#85)** adds a `--thinking` flag to both `chat` and `prompt`; since Bedrock doesn't expose pre-flight capability info for this either, unsupported models are handled the same way as #83 — attempt it, surface the API's error clearly if rejected, don't pre-validate.
6. All five features are additive and off-by-default: no existing flag, config key, or default behavior changes. Backward compatibility is a hard constraint per `CLAUDE.md`.

## Functional Requirements

### FR1 — System Prompt Support (#81)
- FR1.1: `chat` and `prompt` accept a `--system <text>` flag.
- FR1.2: A persisted `system-prompt` config key is supported via `chat-cli config set system-prompt "..."` / `unset` / `list`, following the existing `model-id`/`custom-arn` precedence pattern (flag → config → none).
- FR1.3: When set (from either source), the text is sent via Bedrock's `SystemContentBlocks` on `Converse`/`ConverseStream` requests.
- FR1.4: When no system prompt is set (today's default), request behavior is unchanged — no empty `SystemContentBlocks` sent.

### FR2 — Tool Use / Function Calling (#82)
- FR2.1: `chat` builds a `ToolConfig` from a small, internal tool registry (Go interface: name, description, JSON input schema, execute function).
- FR2.2: When a streamed response yields a `ContentBlockMemberToolUse` block, the CLI stops rendering, executes the matching registered tool with the model-supplied input, and appends a `ToolResultBlock` to the conversation before continuing the stream — this may repeat for multiple tool round-trips in one turn.
- FR2.3: An unrecognized tool name or a tool execution error is reported back to the model as an error `ToolResultBlock` (not a fatal CLI error), so the model can recover or explain the failure to the user.
- FR2.4: Tool-call turns and their results are persisted to the `chats` table like any other turn, so `chat list`/resume still work (existing `Persona` values "User"/"Assistant" are sufficient; no schema change required for this pass).
- FR2.5: One built-in `read_file` tool ships with this issue (read-only, path-traversal-safe, working-directory-confined — same rule as `utils.ReadImage`) so the loop is independently useful and testable.

### FR3 — Prompt Caching (#83)
- FR3.1: When a system prompt (FR1) is present, a cache checkpoint (`cachePoint` content block) is inserted after it.
- FR3.2: When piped document content (`utils.LoadDocument`) is present in a `prompt`/`chat` invocation, a cache checkpoint is inserted after it.
- FR3.3: If the Bedrock API rejects a request because of an unsupported cache point (model doesn't support caching), the CLI retries once without the cache point and logs a non-fatal warning, rather than failing the user's request.
- FR3.4: If cache token metrics are present in the response, they are available for a future token/cost display feature (existing issue #41) — not rendered by default in this pass.

### FR4 — Native Document Input (#84)
- FR4.1: `prompt` accepts a new `--document`/`-d <path>` flag for PDF, CSV, DOC/DOCX, XLS/XLSX, HTML, TXT, and MD files (Bedrock's supported `DocumentBlock` formats), sent as `ContentBlockMemberDocument` rather than raw stdin text.
- FR4.2: Reuses/generalizes the path-safety validation from `utils.ReadImage` (working-directory-confined, extension allow-list, existence check).
- FR4.3: A model that doesn't support document input (checked the same way `--image` is checked today via `InputModalities`, if Bedrock exposes a `DOCUMENT` modality; otherwise fall back to attempt + graceful API-error handling per Assumption 5's pattern) produces a clear, non-cryptic error before the request is sent if possible.
- FR4.4: `--image` behavior is completely unchanged; `--document` is independent and can be combined with `--image` in the same invocation if the model supports both.

### FR5 — Extended Thinking / Reasoning Mode (#85)
- FR5.1: `chat` and `prompt` accept a `--thinking` flag that sets `AdditionalModelRequestFields` to enable reasoning mode for models that support it.
- FR5.2: Reasoning content, when present in the response, is printed visually distinct from the final answer (e.g. dimmed/prefixed, consistent with the existing gray user-echo styling convention already used in `chat.go`).
- FR5.3: Without `--thinking`, behavior is completely unchanged (no new fields sent).
- FR5.4: If a model rejects the `AdditionalModelRequestFields` payload, the CLI surfaces the API's error message clearly rather than a generic failure.

## Non-Functional Requirements

- **NFR1 — Backward Compatibility**: Every new flag defaults to off/empty. No existing flag, config key, default value, or documented behavior changes. Existing integration tests (`integration_test.go`) must continue to pass unmodified.
- **NFR2 — TDD & Coverage**: Per `CLAUDE.md`, tests are written before implementation for each feature. `cmd` package coverage (currently 7.4%) must not regress and should improve given these all land in `cmd/chat.go` and `cmd/prompt.go`. `make test` and `make test-coverage` must pass before any unit is considered complete.
- **NFR3 — Security**: Any new file-reading capability (document input FR4, the `read_file` tool FR2.5) must reuse the existing path-traversal-safe pattern from `utils.ReadImage` — confined to the working directory, no arbitrary filesystem access.
- **NFR4 — Consistent Error Handling**: Fatal, pre-flight-detectable errors (bad flag combinations, unreadable local files) use the existing `log.Fatalf` pattern; recoverable/best-effort failures (cache-point rejection, unrecognized tool) degrade gracefully with a logged warning, matching existing conventions (e.g. deferred `Close()` error handling).
- **NFR5 — No New Required Dependencies**: All five features are expressible with the AWS SDK v2 Bedrock types already in `go.mod` (`SystemContentBlocks`, `ToolConfig`, `ContentBlockMemberToolUse`/`ToolResultBlock`, `cachePoint`, `ContentBlockMemberDocument`, `AdditionalModelRequestFields`). No new third-party dependency should be needed for this pass.
- **NFR6 — Documentation**: `README.md` and `docs/usage.md` are updated with each new flag as it ships (per `CLAUDE.md`'s documentation rules — no new root-level `.md` files).
- **NFR7 — Streaming Compatibility**: All features must work with `chat`'s always-streaming loop (`ConverseStream`) and with `prompt`'s both streaming (default) and `--no-stream` paths.

## Summary

Five additive, off-by-default capabilities bring chat-cli's Bedrock integration up to date with what the Converse API now supports: system prompts (#81), tool use (#82, with one safe built-in tool), prompt caching (#83), native document attachments (#84), and extended thinking (#85). Tool use is the highest-complexity item (new multi-turn control flow) and has the most natural follow-on value (unlocks #86 built-in tools and #87 MCP support later). System prompt support is the lowest-risk, most foundational item — useful on its own and a soft dependency for tool use (instructing the model about available tools) and caching (something to cache). Document input and extended thinking are independent of the other three and of each other. This ordering/dependency information is provided for the upcoming Workflow Planning stage, which will decide actual unit sequencing.
