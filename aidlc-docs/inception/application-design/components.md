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

---

# Initiative 3 Components (#86)

Scope: extends the `Tool`/`Registry` components above (Initiative 1 Unit 2) with 3 new tools and a genuinely new permission-engine subsystem. Verified against the current `tools/tool.go`/`tools/registry.go` source, not assumed.

## Component: Extended `Tool` Interface
- **Purpose**: Let a tool declare whether it's destructive and, if so, how to summarize a specific call for the confirmation prompt and derive its coarse sticky-approval pattern key.
- **Responsibilities**: Two new methods alongside the 4 existing ones (`Name`, `Description`, `InputSchema`, `Execute`). `read_file`/`git_diff` implement them as no-ops (non-destructive); `write_file`/`run_shell` implement them for real (FR5.1, FR6.1-FR6.2).
- **Package**: `tools` (existing package, interface extended)

## Component: `WriteFileTool`, `RunShellTool`, `GitDiffTool`
- **Purpose**: The 3 new built-in tools (FR2-FR4).
- **Responsibilities**: `WriteFileTool` creates/overwrites a file within the working directory (reuses `utils.ValidateLocalPath`), destructive. `RunShellTool` runs a shell command with a timeout and output cap, destructive. `GitDiffTool` runs `git diff [arg]`, read-only, non-destructive.
- **Package**: `tools` (new files), each implements the extended `Tool` interface, following `ReadFileTool`'s existing shape

## Component: `PermissionGate` (interface)
- **Purpose**: The abstract contract for "decide whether a destructive call may proceed" (FR5, FR6) - decouples the decision from how it gets made.
- **Responsibilities**: Defines a `Check` operation and `Decision` result type (allow/deny). `Registry.Dispatch` is extended to accept a `PermissionGate` and consult it before executing a tool that `RequiresConfirmation()`.
- **Package**: `tools` (new)

## Component: `ApprovalStore`
- **Purpose**: Tracks granted sticky approvals at the session tier (in-memory) and the "always" tier (persisted per git repository) (FR6.3, FR7).
- **Responsibilities**: Pure lookup/record logic - no I/O beyond its own persisted-tier backing file, no prompting/UI (that's `InteractivePermissionGate`). Uses the shared git-boundary helper below to scope "always" approvals per repository.
- **Package**: `tools` (new)

## Component: Shared Git-Boundary Helper (extracted to `utils`)
- **Purpose**: `ApprovalStore` needs the same repo-root-detection behavior #88 already built (`findGitBoundary`, currently private inside `cmd/projectcontext.go`), but `tools` cannot import `cmd` (import cycle - `cmd` already imports `tools`).
- **Decision**: Extract the boundary-walk logic into a new exported `utils.FindGitBoundary(dir string) string`; refactor `cmd/projectcontext.go` to call it instead of keeping a private copy. Justified refactor of already-shipped Initiative 2 code - avoids duplicating logic that's already had one real bug found and fixed in it (Initiative 2's epilogue commit).
- **Package**: `utils` (existing package, new exported function; `cmd/projectcontext.go` refactored to call it)

## Component: `InteractivePermissionGate`
- **Purpose**: The concrete, terminal-facing `PermissionGate` implementation - shows the confirmation prompt, reads the once/session/always/deny choice, records it via `ApprovalStore`.
- **Responsibilities**: Owns all user-facing I/O for this initiative - consistent with the existing split where `tools/` holds abstract contracts and `cmd/` holds concrete terminal orchestration (mirrors `cmd/toolloop.go`'s relationship to `tools.Registry`).
- **Package**: `cmd` (new)

## Existing Components Extended (Initiative 3)
- **`cmd/chat.go`** — tool-enablement wiring changes: always builds the registry (now 4 tools) and an `InteractivePermissionGate`; the `--tools` flag is removed; a retry-without-tools-on-rejection fallback is added, mirroring `cmd/promptcache.go`'s existing cache-point retry pattern (FR1).
- **`tools.Registry.Dispatch`** — signature extended to accept a `PermissionGate`, consulted before executing any tool where `RequiresConfirmation()` is true.
