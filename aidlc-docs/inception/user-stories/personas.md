# Personas

chat-cli is a single-user-type terminal tool — there is one persona. A separate "Requirements Analysis" persona split was not warranted (see `aidlc-docs/inception/plans/user-stories-assessment.md`).

## Persona: The chat-cli User

- **Who they are**: A developer or technically comfortable professional who already has AWS credentials configured locally and uses the terminal as a primary workspace. May be using chat-cli standalone, or alongside other CLI tools (editors, git, other AI coding assistants).
- **Goals**:
  - Get fast, scriptable access to Bedrock foundation models without leaving the terminal or standing up any infrastructure.
  - Keep a running, resumable conversation with useful context (system instructions, attached documents) without repeating themselves every session.
  - Increasingly, use the model to *do* things (not just answer questions) — inspect files, reason over documents — without switching to a heavier agentic tool.
- **Pain points this initiative addresses**:
  - No way to set persistent instructions/persona for the model (no system prompt) — every session starts "cold."
  - Can't have the model take any action beyond generating text — no tool use.
  - Repeated large system prompts or documents cost more and respond slower than necessary — no caching.
  - Can only attach images, not the PDFs/CSVs/docs they actually work with day to day.
  - Can't see or use the newer reasoning/extended-thinking capabilities some models now offer.
- **Technical proficiency**: High — comfortable with CLI flags, config files, and reading error messages; not necessarily a Go developer.
- **Relationship to GitHub issues**: This persona's needs map directly to issues #81-#85.

## Persona Extension: Initiative 3 (Built-in Agent Tools, #86)

Same persona, same file — no new persona type, extending the existing one with goals/pain points specific to this initiative:

- **Additional goals**:
  - Let the model take real action in the local project (edit a file, run a command, inspect a diff) without switching to a separate, heavier coding-agent tool.
  - Trust that destructive actions (writes, shell commands) never happen without an explicit decision, but without being nagged for the same decision repeatedly within — or across — sessions.
- **Additional pain points this initiative addresses**:
  - Had to remember and pass `--tools` every session to get any tool behavior at all, even the safe, read-only `read_file` tool from #82.
  - No way for the model to edit files, run commands, or see a diff — `chat` could only read, never act.
  - No middle ground between "the model can never do anything destructive" and "every single destructive call needs a fresh yes/no" — no way to say "yes, and stop asking for `git` commands this session."
