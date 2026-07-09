# NFR Requirements and Design (Combined): Unit 7, New Built-in Tools (#86)

Combined presentation, same pattern as Unit 6. Security remains dominant - these are the tools Unit 6's gate exists to protect against.

## Security

### Requirement
`write_file` must stay confined to the working directory exactly like `read_file`. `run_shell` has no path confinement to speak of (arbitrary commands), so the confirmation gate (Unit 6) is the entire control - this unit must not weaken that by, say, executing before the gate is consulted.

### Design
- `write_file` reuses the exact same confinement algorithm as `read_file`/`--document`/`--image` (`ValidateLocalPathForWrite`, sharing `confineToWorkingDir` with `ValidateLocalPath`) - no new traversal-safety logic invented for this unit.
- Both `WriteFileTool.Execute` and `RunShellTool.Execute` are only ever reachable through `Registry.Dispatch`, which (per Unit 6) always consults the gate first since both declare `RequiresConfirmation() == true`. Nothing in this unit calls `Execute` directly.
- `run_shell`'s lack of a command allowlist (Requirements Analysis Assumption 2) is a deliberate, already-approved tradeoff - re-flagged here rather than silently accepted, since it's the single highest-risk capability added in this entire initiative. The gate (once/session/always/deny, per-base-command pattern matching) is the sole mitigation.
- `git_diff` has no destructive capability at all (it can only read repository state) - Security review for it is limited to "does it accept arbitrary flags that could do something unexpected" - it does not; only a single positional `arg` is passed, never interpreted as multiple shell tokens (passed as one `exec.Command` argument, not through a shell that could split/reinterpret it - see Reliability below for why this also matters there).

### Compliance
✅ Compliant - `write_file` reuses proven confinement logic; `run_shell`'s risk is explicitly the gate's job, not this unit's; `git_diff` has no meaningful attack surface.

## Reliability

### Requirement
`run_shell` must not hang `chat` indefinitely, and a failing command (bad exit code) must not be indistinguishable from a tool-implementation bug.

### Design
- BR8: a hard 30-second timeout via `context.WithTimeout`, always applied - no way for a single `run_shell` call to hang the session forever.
- BR10: a non-zero exit code flows back to the model as normal tool output (with the exit code visible), not as a Go error - keeps the important reliability distinction between "the tool itself is broken" (a real Go error: failed to start `sh`, timeout) and "the command the model chose to run didn't succeed" (expected, normal operation, the model should just see it and adapt).
- `git_diff`'s `arg` is passed as a single `exec.Command` argument (not concatenated into a shell string) - this isn't just security, it's reliability too: a ref/path containing spaces or shell metacharacters behaves as literal text, not as multiple arguments or shell syntax, so `git diff` gets exactly what the model asked for.

### Compliance
✅ Compliant - bounded execution time, and a clear, non-error-based signal for the very common "command failed" case.

## Usability

### Requirement
Same as Unit 6's carried-forward NFR5: the confirmation summary must be informative without flooding the terminal.

### Design
- `write_file`'s summary truncates content past 4KB with an explicit note (BR4), same threshold and rationale as Unit 6's NFR design.
- `run_shell`'s summary is simply the exact command string - already inherently compact, no truncation needed at the confirmation stage (the *output*, not the confirmation prompt, is where truncation applies - BR9, and that's shown after the fact, not at the gate).

### Compliance
✅ Compliant.

## Non-Applicable Categories
- **Scalability/Performance/Availability**: N/A, same rationale as every prior unit - single local process. `run_shell`'s 30s cap is a reliability bound, not a performance target.
