# Business Rules — Unit 2 (Tool Use / Function Calling)

## Rule 1: Multiple tool calls in one turn are batched into one follow-up message
If the model requests N tools in a single turn, `Registry.Dispatch` is called once per tool, but all N `ToolResultBlock`s are combined into a **single** follow-up user-role message before re-sending — not N separate messages. This matches Bedrock's expected conversation shape (one message per "turn", with multiple content blocks inside it).

## Rule 2: Unknown tool name never crashes the CLI (FR2.3)
If the model requests a tool name not present in the registry, `Registry.Dispatch` returns a `ToolResultBlock{Status: "error", Content: [text: "unknown tool: <name>"]}`. This is sent back to the model like any other result — the model sees the error and can explain it to the user or try something else. The CLI process does not exit and no Go error is returned up the call stack for this case.

## Rule 3: Tool execution failure never crashes the CLI (FR2.3)
If a registered tool's `Execute` returns a Go `error` (e.g. `read_file` given a path outside the working directory), `Registry.Dispatch` wraps it as `ToolResultBlock{Status: "error", Content: [text: err.Error()]}`. Same non-fatal handling as Rule 2.

## Rule 4: Malformed tool-input JSON is treated as a dispatch-time error, not a parse panic
If the accumulated `ToolUseBlockDelta.Input` fragments don't form valid JSON when finalized (rare, but possible with a misbehaving model), this is caught and converted to an error `ToolResultBlock` (`"invalid tool input: <parse error>"`) rather than allowed to panic or silently corrupt state.

## Rule 5: Tool round-trips per user turn are capped (new rule, not explicit in original FRs — added here as a reliability safeguard)
A single user turn can trigger at most **10** consecutive tool round-trips (`StopReasonToolUse` responses) before the CLI stops looping and surfaces a clear message to the user ("stopped after 10 tool calls in a single turn to avoid a runaway loop — you can ask a follow-up to continue"). This protects against a misbehaving or overly agentic model looping indefinitely without user input, consistent with NFR4 (consistent, non-crashing error handling) even though no story explicitly asked for it — flagging this as an assumption for review rather than silently adding unbounded-loop risk.

## Rule 6: Only the final text response and the original prompt are persisted per turn
See `business-logic-model.md`'s "Persistence Decision" — restated here as a rule because it constrains `ChatRepository.Create` call sites: exactly 2 `Create` calls per user turn (one "User", one "Assistant"), regardless of how many tool round-trips happened in between. This is unchanged from Unit 1/today's behavior.

## Rule 7: `read_file` tool is read-only and working-directory-confined (FR2.5, NFR3)
The built-in `read_file` tool can only read files; it has no write/delete/execute capability. Path resolution reuses `utils.ValidateLocalPath` (introduced in this unit, per `application-design/components.md`) — any path that resolves outside the current working directory is rejected before the file is touched, returning an error result (Rule 3) rather than reading it.
