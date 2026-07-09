# Component Dependency

## Dependency Matrix

| Component | Depends On | Depended On By |
|---|---|---|
| `cmd/chat.go` (`chatCmd.Run`) | `config.FileManager`, `tools.Registry`, `tools.ReadFileTool`, `utils` (cache-point helper), `repository.ChatRepository`, AWS SDK Bedrock types | (entry point — nothing depends on it) |
| `cmd/prompt.go` (`promptCmd.Run`) | `config.FileManager`, `utils` (cache-point helper, `ValidateLocalPath`, `ReadDocument`, existing `ReadImage`), AWS SDK Bedrock types | (entry point — nothing depends on it) |
| `tools.Registry` | `tools.Tool` interface, AWS SDK Bedrock types (`ToolConfig`, `ToolUseBlock`, `ToolResultBlock`) | `cmd/chat.go` |
| `tools.ReadFileTool` | `tools.Tool` interface, `utils.ValidateLocalPath` | `tools.Registry` (registered into it) |
| `utils.ValidateLocalPath` | standard library (`os`, `path/filepath`) | `tools.ReadFileTool`, `utils.ReadDocument`, `utils.ReadImage` (refactored to reuse it) |
| `utils.ReadDocument` (new) | `utils.ValidateLocalPath` | `cmd/prompt.go` |
| `utils` cache-point helper (new) | AWS SDK Bedrock types | `cmd/chat.go`, `cmd/prompt.go` |
| `config.FileManager` | Viper (existing) | `cmd/chat.go`, `cmd/prompt.go` (existing + new `system-prompt` key) |
| `repository.ChatRepository` | `db.Database` (existing) | `cmd/chat.go` (existing, unchanged interface) |

No changes to: `db`, `db/sqlite`, `factory` — none of the 5 features touch persistence connection/migration logic.

## Dependency Diagram

```mermaid
flowchart TD
    chatCmd["cmd/chat.go"] --> configFM["config.FileManager"]
    chatCmd --> registry["tools.Registry"]
    chatCmd --> cacheUtil["utils cache-point helper"]
    chatCmd --> chatRepo["repository.ChatRepository"]

    promptCmd["cmd/prompt.go"] --> configFM
    promptCmd --> cacheUtil
    promptCmd --> validatePath["utils.ValidateLocalPath"]
    promptCmd --> readDoc["utils.ReadDocument"]
    promptCmd --> readImg["utils.ReadImage (existing)"]

    registry --> toolIface["tools.Tool interface"]
    readFileTool["tools.ReadFileTool"] --> toolIface
    readFileTool --> validatePath
    registry -.->|"registers"| readFileTool

    readDoc --> validatePath
    readImg --> validatePath

    chatRepo --> dbIface["db.Database (existing, unchanged)"]
```

### Text Alternative
```
cmd/chat.go -> config.FileManager, tools.Registry, utils cache-point helper, repository.ChatRepository
cmd/prompt.go -> config.FileManager, utils cache-point helper, utils.ValidateLocalPath,
                 utils.ReadDocument, utils.ReadImage (existing)
tools.Registry -> tools.Tool interface; registers tools.ReadFileTool
tools.ReadFileTool -> tools.Tool interface, utils.ValidateLocalPath
utils.ReadDocument -> utils.ValidateLocalPath
utils.ReadImage (refactored) -> utils.ValidateLocalPath
repository.ChatRepository -> db.Database (unchanged)
```

## Communication Patterns
- All calls are synchronous, in-process Go function calls — no new IPC, no new network boundaries, no new goroutine coordination beyond what `ProcessStreamingOutput` already does for streaming.
- `tools.Registry.Dispatch` returns a Bedrock content block value (not an error propagated up the call stack) by design — tool failures are conversation-level events the model should see and can respond to (FR2.3), not CLI-level fatal errors.

---

# Initiative 3 Dependencies (#86)

## Dependency Matrix

| Component | Depends On | Depended On By |
|---|---|---|
| `cmd/chat.go` (extended) | `tools.Registry` (now 4 tools), `cmd.InteractivePermissionGate`, `tools.ApprovalStore`, `utils.FindGitBoundary` | (entry point) |
| `tools.Registry.Dispatch` (extended) | `tools.Tool` (extended interface), `tools.PermissionGate` | `cmd/chat.go` |
| `tools.WriteFileTool` | `tools.Tool` interface, `utils.ValidateLocalPath` | `tools.Registry` |
| `tools.RunShellTool` | `tools.Tool` interface, standard library (`os/exec`) | `tools.Registry` |
| `tools.GitDiffTool` | `tools.Tool` interface, standard library (`os/exec`) | `tools.Registry` |
| `tools.PermissionGate` (interface) | none | `tools.Registry.Dispatch`, `cmd.InteractivePermissionGate` (implements it) |
| `tools.ApprovalStore` | `utils.FindGitBoundary` | `cmd.InteractivePermissionGate` |
| `utils.FindGitBoundary` (extracted) | standard library (`os`, `path/filepath`) | `tools.ApprovalStore`, `cmd/projectcontext.go` (refactored to reuse it instead of its private copy) |
| `cmd.InteractivePermissionGate` | `tools.PermissionGate` interface, `tools.ApprovalStore` | `cmd/chat.go` |

No changes to: `db`, `db/sqlite`, `factory`, `repository`, `config` — this initiative is entirely within `tools`/`cmd`/one `utils` extraction.

## Dependency Diagram

```mermaid
flowchart TD
    chatCmd2["cmd/chat.go"] --> registry2["tools.Registry (4 tools)"]
    chatCmd2 --> gate["cmd.InteractivePermissionGate"]

    registry2 --> toolIface2["tools.Tool (extended)"]
    registry2 --> gateIface["tools.PermissionGate"]

    writeTool["tools.WriteFileTool"] --> toolIface2
    writeTool --> validatePath2["utils.ValidateLocalPath (existing)"]
    shellTool["tools.RunShellTool"] --> toolIface2
    diffTool["tools.GitDiffTool"] --> toolIface2

    registry2 -.->|"registers"| writeTool
    registry2 -.->|"registers"| shellTool
    registry2 -.->|"registers"| diffTool

    gate --> gateIface
    gate --> approvalStore["tools.ApprovalStore"]
    approvalStore --> gitBoundary["utils.FindGitBoundary (extracted)"]

    projectContext["cmd/projectcontext.go (#88, refactored)"] --> gitBoundary
```

### Text Alternative
```
cmd/chat.go -> tools.Registry (4 tools), cmd.InteractivePermissionGate
tools.Registry.Dispatch -> tools.Tool (extended), tools.PermissionGate
tools.WriteFileTool/RunShellTool/GitDiffTool -> tools.Tool interface
tools.WriteFileTool -> utils.ValidateLocalPath (existing, reused)
cmd.InteractivePermissionGate -> tools.PermissionGate, tools.ApprovalStore
tools.ApprovalStore -> utils.FindGitBoundary (extracted from #88's cmd/projectcontext.go)
cmd/projectcontext.go -> utils.FindGitBoundary (refactored to reuse, was private findGitBoundary)
```

## Communication Patterns (Initiative 3)
- `Registry.Dispatch`'s permission check is synchronous and blocking, same execution model as everything else in the tty-loop — `InteractivePermissionGate.Check` reads from stdin directly, no new goroutines.
- `ApprovalStore`'s persisted tier does file I/O only on a `RecordAlways` call (write) and once at construction (read) — not on every `IsApproved` lookup, which stays in-memory after initial load.
