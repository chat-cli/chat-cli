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
