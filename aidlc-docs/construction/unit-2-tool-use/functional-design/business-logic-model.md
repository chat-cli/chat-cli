# Business Logic Model — Unit 2 (Tool Use / Function Calling)

## SDK Types Verified (correcting Application Design's naming)

Inspecting the installed `bedrockruntime@v1.23.0` SDK directly (not guessed) surfaced two naming corrections versus `application-design/component-methods.md`, which used placeholder names before this level of detail was needed:

- The request field is `ConverseStreamInput.ToolConfig *types.ToolConfiguration` (not `ToolConfig` as a type name) — `ToolConfiguration{Tools []types.Tool, ToolChoice}`.
- `types.Tool` is itself a union interface (satisfied by `types.ToolMemberToolSpec{Value: types.ToolSpecification{Name, Description, InputSchema}}`) — the SDK's `Tool` is the "entry in the tools list" concept, distinct from this codebase's own `tools.Tool` interface (different package, no collision, but worth noting to avoid confusion when reading code).
- `types.ToolInputSchema` is a union too, satisfied by `types.ToolInputSchemaMemberJson{Value: document.Interface}` — the JSON schema is supplied as a `document.Interface`, not a raw `json.RawMessage`.

## Core Algorithm: Streaming Tool-Use Loop

Bedrock's tool-use protocol over `ConverseStream` is a **content-block-indexed streaming protocol**, not a single "here's a tool call" event — this is materially more complex than the existing `utils.ProcessStreamingOutput`, which only ever handles one text block. The full sequence:

1. `ContentBlockStartMemberToolUse{ToolUseBlockStart{Name, ToolUseId}}` arrives for a given block index — a tool call has begun.
2. Zero or more `ContentBlockDeltaMemberToolUse{ToolUseBlockDelta{Input *string}}` arrive for that same index — each is a fragment of the tool's input, streamed as partial JSON text that must be concatenated (not parsed until complete).
3. A `ContentBlockStop` (implicit, by index) signals that block is finalized — only then is the accumulated JSON string valid to parse.
4. Text and tool-use blocks can be interleaved by index in the same turn (a model can say something and call a tool in one response) and **multiple tool-use blocks can appear in one turn** (the model can request several tools at once).
5. `ConverseStreamOutputMemberMessageStop{MessageStopEvent{StopReason}}` ends the turn. `StopReasonToolUse ("tool_use")` specifically means: "I'm done talking for now, execute the tool(s) I asked for and tell me the results."

### Algorithm
```
function runTurn(userInput):
    append user message to conversation
    roundTrips := 0
    loop:
        send ConverseStream(conversation, ToolConfig)
        accumulate content blocks by index (text -> print immediately as today;
            tool-use -> buffer Name/ToolUseId/Input fragments, do not print)
        on MessageStop:
            if StopReason != "tool_use":
                finalize accumulated text as the assistant's response for this turn
                append assistant message (text only) to conversation
                return finalized text   # normal end, same as today
            else:
                roundTrips++
                if roundTrips > maxToolRoundTrips:
                    treat as an error turn (see business-rules.md) and return
                parse each finalized tool-use block's buffered JSON
                append assistant message (containing the ToolUseBlock(s)) to conversation
                for each tool-use block:
                    result := registry.Dispatch(toolUseBlock)   # never panics, see business-rules.md
                append one user message containing all ToolResultBlock(s) to conversation
                continue loop   # re-send immediately, no user input needed
```

## Persistence Decision (Functional Design refinement of FR2.4)

The `chats` SQLite table only stores `Persona` ("User"/"Assistant") + plain-text `Message` — there is no column for structured tool-call content (confirmed in Application Design; no schema change is in scope). Storing a serialized tool call as fake plain-text history would be lossy and, worse, would be replayed on a future `--chat-id` resume as an ordinary text message — which is semantically wrong (a resumed `ToolResultBlock` reference with no matching live `ToolUseId` is not valid input to Bedrock, and simply guarantees confusing conversation flow).

**Decision**: Only the user's original prompt and the *final* natural-language assistant response are persisted to SQLite for a turn — exactly as today. Intermediate tool-use/tool-result exchanges exist only in the in-memory `converseStreamInput.Messages` for the remainder of that live session (so the model retains full context within the session) and are discarded when the process exits. This satisfies Story 2.1's acceptance criterion ("the turn is present in history the same way a normal turn is") — a `chat list`/resume shows the question and the final answer, identical UX to any other turn — without a lossy or unsafe persistence hack.
