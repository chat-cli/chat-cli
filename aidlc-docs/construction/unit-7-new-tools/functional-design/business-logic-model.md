# Functional Design: Business Logic Model - Unit 7, New Built-in Tools (#86)

## Context
Verified against the actual current codebase, not assumed.

## Important Finding: `utils.ValidateLocalPath` Requires the File to Already Exist

`utils.ValidateLocalPath` (`utils/utils.go:74`) does two things: confines resolution to the working directory, **and** errors if the file doesn't exist (`os.Stat`/`os.IsNotExist`). This is correct for its existing callers (`read_file`, `--document`, `--image` - all read-only), but `write_file` must support *creating* a new file (FR2.2), which would always fail existence-checking.

**Decision**: rather than changing `ValidateLocalPath`'s signature (which would ripple across Units 2/4's already-shipped call sites for no benefit to them), extract the shared confinement-only logic into a private helper both functions call, and add a new sibling function:

```go
func ValidateLocalPath(filename string) (string, error)         // unchanged contract - existence required
func ValidateLocalPathForWrite(filename string) (string, error) // confinement only, no existence check
```

This is additive, not a refactor of existing behavior - `ValidateLocalPath`'s callers (Units 2/4) are unaffected, zero regression risk to already-shipped code. Contrast with Unit 6's `FindGitBoundary` extraction, which *did* change an existing private function's location - this is lower-risk since nothing existing is being moved or altered.

## `WriteFileTool`
- `RequiresConfirmation()` → `true`
- `ConfirmationSummary`: unmarshals `{"path":"...","content":"..."}`, calls `utils.ValidateLocalPathForWrite`. On success, summary is `"Write to <path>:\n<content, or a truncated preview past 4KB per the NFR usability decision>"`. Pattern key (BR2, Unit 6's design): the containing directory of the resolved path, relative to the git repo root (`utils.FindGitBoundary(cwd)`) if inside one, else relative to cwd itself.
- `Execute`: re-validates via `ValidateLocalPathForWrite` (never trusts the summary step's validation alone - each method is independently complete, per Unit 6's stated principle), writes the file (`os.WriteFile`, mode `0600` for new files; an existing file's mode is preserved by `os.WriteFile`'s overwrite behavior... **correction**: `os.WriteFile` always applies the given perm only when *creating* a new file - an existing file keeps its current mode regardless of the perm argument, which is exactly the desired behavior here with no extra code needed).

## `RunShellTool`
- `RequiresConfirmation()` → `true`
- `ConfirmationSummary`: unmarshals `{"command":"..."}`, summary is `"Run: <command>"`, pattern key is the command's first whitespace-separated token (`strings.Fields(command)[0]`, or `""` if the command is empty/whitespace-only - never matches a prior approval, always prompts).
- `Execute`: `exec.CommandContext` with a 30s timeout (`context.WithTimeout`), `Dir` set to `os.Getwd()`'s result, `sh -c "<command>"`, combined stdout+stderr via `CombinedOutput()`, truncated at 32KB (mirrors #88's precedent constant). A **non-zero exit code is not a Go error** - it's embedded in the returned text (`"[exit code: N]\n<output>"`) so the model sees the command ran and can react to its failure, the same way a human running the command would see both the output and the exit status. A **timeout** (`ctx.Err() == context.DeadlineExceeded`) or a failure to start the shell at all *are* returned as Go errors (distinct failure classes from "the command the model asked for ran and failed").

## `GitDiffTool`
- `RequiresConfirmation()` → `false` (read-only)
- `Execute`: unmarshals `{"arg":"..."}` (optional, may be absent/empty), runs `git diff [arg]` via `exec.CommandContext` (same 30s timeout for consistency, though git diff is expected to be fast) in `os.Getwd()`. No `FindGitBoundary` involvement needed - `git diff` itself already reports "not a git repository" (or an invalid-ref error) via a non-zero exit and stderr text if run outside a repo or with a bad argument; that stderr text becomes the tool's returned error, giving the model the same clear signal git itself would give a human.
