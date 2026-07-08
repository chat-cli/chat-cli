# Components

Scope: only genuinely new/reusable components introduced by issues #81-#85. System prompt (#81) and extended thinking (#85) are thin flag/field additions directly inside existing `cmd/chat.go`/`cmd/prompt.go` orchestration and are not separate components — see `services.md`.

## Component: `Tool` (interface)
- **Purpose**: A uniform contract any tool-callable capability implements, so the tool-use loop (Story 2.1) doesn't need to know about concrete tools.
- **Responsibilities**: Advertise a name, description, and input schema to build Bedrock's `ToolConfig`; execute given the model-supplied input; report success/error distinctly.
- **Package**: `tools` (new)

## Component: `Registry`
- **Purpose**: Own the set of registered tools and mediate between Bedrock's tool-use protocol and concrete `Tool` implementations (FR2.1-FR2.3).
- **Responsibilities**: Build a `types.ToolConfig` from registered tools; dispatch an incoming `ContentBlockMemberToolUse` to the matching tool; wrap unknown-tool-name and execution-failure cases into an error `ToolResultBlock` instead of failing the CLI (FR2.3).
- **Package**: `tools` (new)
- **Consumed by**: `cmd/chat.go`'s conversation loop

## Component: `ReadFileTool`
- **Purpose**: The one built-in tool shipped with this initiative (Story 2.2/FR2.5) — lets the model read a local file mid-conversation.
- **Responsibilities**: Implement the `Tool` interface; validate the requested path via the shared path-safety helper before reading; return file contents or a clear error.
- **Package**: `tools` (new), implements the `Tool` interface

## Component: `utils.ValidateLocalPath` (extracted/generalized helper)
- **Purpose**: Single source of truth for "is this local file path safe to read," reused by `ReadFileTool` (#82) and the new document-attachment flow (#84), instead of duplicating the traversal-safety logic `utils.ReadImage` already has inline.
- **Responsibilities**: Confine resolution to the working directory (same rule `ReadImage` already enforces); return the validated absolute path or a descriptive error.
- **Package**: `utils` (existing package, new exported function; `ReadImage` is refactored to call it)

## Component: Cache-point helper (working name: `utils.CachePoint` support)
- **Purpose**: Shared logic for appending a Bedrock cache checkpoint after system-prompt/document content and retrying once without it if the API rejects it (FR3.1-FR3.3), reused by both `chat` and `prompt` instead of duplicating retry logic in each command.
- **Responsibilities**: Build the cache-point content block; wrap a Converse/ConverseStream call with a "retry once without cache point on rejection" policy; log a non-fatal warning on fallback.
- **Package**: `utils` (existing package, new function(s))

## Existing Components Extended (not redesigned)
- **`config.FileManager`** — gains one new supported config key (`system-prompt`), using the exact same `GetConfigValue` precedence mechanism already in place for `model-id`/`custom-arn`. No interface change.
- **`repository.ChatRepository`** — used as-is; tool-call turns persist through the existing `Create` method with existing `Persona` values (FR2.4). No interface change.
- **`cmd/prompt.go`'s attachment handling** — gains a document-attachment code path parallel to the existing image-attachment path (FR4.1), built on `utils.ValidateLocalPath` and a new `ContentBlockMemberDocument` builder. Not a new component — an extension of existing command logic (see `services.md`).
