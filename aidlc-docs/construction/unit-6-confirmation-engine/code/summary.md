# Code Generation Summary: Unit 6, Confirmation and Sticky Approval Engine (#86)

## Files Created
- `tools/permission.go` - `Decision` type/constants, `PermissionGate` interface
- `tools/approvalstore.go` - `ApprovalStore` (session tier in-memory, always tier persisted YAML at `<ConfigPath>/tool-approvals.yaml`, `0600`, keyed by repo root), `CanRecordAlways`
- `tools/approvalstore_test.go` - 3 session-tier subtests, 5 always-tier subtests
- `cmd/permissionprompt.go` - `InteractivePermissionGate`, the concrete terminal-facing gate
- `cmd/permissionprompt_test.go` - choice parsing (11 subtests), session/always recording, not-offered-outside-a-repo, prior-approval-skips-the-prompt

## Files Modified
- `tools/tool.go` - `Tool` interface gains `RequiresConfirmation()`/`ConfirmationSummary()`
- `tools/readfile.go` - implements the 2 new methods (`RequiresConfirmation() == false`)
- `tools/registry.go` - `Dispatch` signature gains a `gate PermissionGate` parameter; gate-consulting logic inserted before `Execute`
- `tools/registry_test.go` - `fakeTool` updated for the extended interface; new `fakeDestructiveTool`/`fakeGate` test doubles; 4 new `Dispatch` gate-consulting test cases; 3 existing calls updated to pass `nil` (safe - none of their fakes require confirmation)
- `utils/utils.go` - new exported `FindGitBoundary` (extracted from `cmd/projectcontext.go`'s private `findGitBoundary`)
- `utils/utils_test.go` - new `TestFindGitBoundary`
- `cmd/projectcontext.go` - private `findGitBoundary` deleted, call site now uses `utils.FindGitBoundary`
- `cmd/toolloop.go` - `runChatTurnWithTools` gains a `gate tools.PermissionGate` parameter, threaded to `registry.Dispatch`
- `cmd/toolturn_test.go` - `fakeTurnTool` updated for the extended interface; 2 call sites pass `nil` (safe - non-destructive fake)
- `cmd/chat.go` - constructs a real `ApprovalStore`/`InteractivePermissionGate` alongside the existing registry-construction block; both `runChatTurnWithTools` call sites pass it through

## Design Deviations from Functional Design (minor, during implementation)
- **Persisted-entry format corrected during implementation**: the first draft accidentally used the internal NUL-joined key (`"toolName\x00patternKey"`) for on-disk storage. Caught and fixed before tests were written against it - the persisted format uses the documented human-readable `"toolName:patternKey"` (`business-logic-model.md`), parsed back with `strings.Cut` on load. The in-memory maps still use the NUL-joined key internally.
- **`ApprovalStore.CanRecordAlways() bool`** - a small addition not explicitly listed in `component-methods.md`, needed so `InteractivePermissionGate` can decide whether to offer/accept "always" as a choice (BR10) without reaching into the store's private `repoRoot` field.
- **Go's whole-package compilation forced Steps 11-13 together**: writing `cmd/permissionprompt_test.go` (Step 11) couldn't compile in isolation once `Registry.Dispatch`'s signature changed (Step 10), since `cmd/toolloop.go`'s existing call site broke immediately. Steps 12 (implement the gate) and 13 (thread it through `chat.go`/`toolloop.go`/`toolturn_test.go`) were completed together with Step 11 rather than strictly sequentially - the TDD red/green discipline was preserved per-function, just not per-step-boundary.

## Test Results
- `make test`: all packages pass, no regressions (2 pre-existing `SKIP`s unrelated to this unit)
- `make lint`: clean
- Coverage: `cmd` 31.8% → 33.3%, `tools` 90.0% → 84.1% (new untested edge branches in `approvalstore.go`'s malformed-file-merge path; still high), `utils` 51.8% → 53.7%, total 67.8% → 69.7% - no regression
- `go test -tags=integration -v .`: 7/7 pass
- No user-visible behavior change from this unit alone (by design) - `read_file` remains the only registered tool and is non-destructive, so the new gate is fully wired but never actually triggered yet. Unit 7 will be the first unit to exercise it end-to-end.
