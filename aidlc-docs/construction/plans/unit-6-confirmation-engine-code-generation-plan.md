# Code Generation Plan: Unit 6, Confirmation and Sticky Approval Engine (#86)

## Unit Context
- **Stories**: 8.1-8.3 (FR5.1-FR5.4, FR6.1-FR6.3, FR7.1-FR7.2)
- **Design source**: `aidlc-docs/construction/unit-6-confirmation-engine/functional-design/`, `.../nfr-requirements/nfr-requirements-and-design.md`
- **Dependencies**: None (foundational unit)
- **Verified call-site reality** (not assumed): `tools.Registry.Dispatch` is currently called from exactly one place, `cmd/toolloop.go:205` inside `runChatTurnWithTools`, which itself is called from `cmd/chat.go` (2 call sites) and `cmd/toolturn_test.go` (2 call sites). Changing `Dispatch`'s signature means `runChatTurnWithTools` needs a new `gate tools.PermissionGate` parameter too, threaded through all 4 call sites - this unit must leave the tree fully compiling and green, not just add new files, since `--tools`/registry wiring isn't removed until Unit 8.
- **Design choice for this unit's call-site update**: construct a real `InteractivePermissionGate` (backed by a real `ApprovalStore`) at the existing registry-construction point in `chat.go`, rather than a stub/nil gate. Currently inert in practice - the only registered tool (`read_file`) has `RequiresConfirmation() == false`, so the gate is never actually consulted until Unit 7 adds destructive tools - but this keeps every intermediate commit's behavior fully correct, not just compiling.

## Steps

- [x] **Step 1 - Failing test: extended `Tool` interface on `ReadFileTool`**
  Add `TestReadFileTool_RequiresConfirmation` to `tools/readfile_test.go` asserting `RequiresConfirmation() == false`. Run `go test ./tools/... -run TestReadFileTool_RequiresConfirmation`, confirm compile failure (method doesn't exist).

- [x] **Step 2 - Extend `Tool` interface + implement on `ReadFileTool`**
  Add `RequiresConfirmation() bool` and `ConfirmationSummary(input json.RawMessage) (summary, patternKey string, err error)` to `tools/tool.go`'s `Tool` interface. Implement on `ReadFileTool` (`RequiresConfirmation` returns `false`; `ConfirmationSummary` returns empty values, never called in practice per BR5). Update any fake `Tool` implementations in `tools/registry_test.go` to satisfy the extended interface (check current fakes first). Run Step 1's test, confirm green; run full `tools` package tests, confirm no regressions.

- [x] **Step 3 - Failing tests: `Decision`/`PermissionGate`/`ApprovalStore` session tier**
  Create `tools/approvalstore_test.go` with cases for BR6-BR9's session-tier behavior: `IsApproved` false before any record; true after `RecordSession`; independent per `toolName+patternKey` pair (approving `run_shell:git` doesn't approve `run_shell:npm` or `write_file:git`). Confirm compile failure.

- [x] **Step 4 - Implement `Decision`, `PermissionGate`, `ApprovalStore` session tier**
  Create `tools/permission.go` (`Decision` type + constants, `PermissionGate` interface). Create `tools/approvalstore.go` with `ApprovalStore` struct (session tier only for now - `map[string]bool` keyed per the `\x00`-joined key from `business-logic-model.md`), `NewApprovalStore`, `IsApproved`, `RecordSession`. Run Step 3 tests, confirm green.

- [x] **Step 5 - Failing tests: `ApprovalStore` persisted "always" tier**
  Add cases to `approvalstore_test.go` using `t.TempDir()` for a fake `ConfigPath`: `RecordAlways` writes to `<path>/tool-approvals.yaml` with `0600` perms (BR14); a fresh `NewApprovalStore` pointed at that path loads the prior "always" approval (`IsApproved` returns true without a `RecordSession` call); different repo roots don't see each other's "always" approvals (FR7.1, the security-critical case flagged in NFR); a missing or malformed YAML file at construction yields zero approvals, not an error (BR15); `RecordAlways` is a safe no-op when `repoRoot == ""` (BR10's "not in a repo" case). Confirm compile failure.

- [x] **Step 6 - Implement `ApprovalStore` persisted tier**
  Add YAML load/save (`gopkg.in/yaml.v3`, already a dependency per `cmd/config.go`), `RecordAlways`, the repo-root-keyed structure from `business-logic-model.md`. Run Step 5 tests, confirm green.

- [x] **Step 7 - Failing tests: `utils.FindGitBoundary`**
  Add `TestFindGitBoundary` to `utils/utils_test.go`, porting the exact scenarios `cmd/projectcontext_test.go`'s existing boundary-related subtests already cover (match at cwd's own `.git`, match at a nested-parent's `.git`, no `.git` anywhere returns `""`, capped walk) so the extracted function is proven equivalent before the extraction happens. Confirm compile failure (function doesn't exist yet in `utils`).

- [x] **Step 8 - Extract `utils.FindGitBoundary`, refactor `cmd/projectcontext.go`**
  Add `FindGitBoundary(dir string) string` to `utils/utils.go` (identical body to `cmd/projectcontext.go`'s current private `findGitBoundary`, including the 64-level cap). Delete the private copy in `cmd/projectcontext.go` and replace its one call site with `utils.FindGitBoundary(...)`. Run Step 7's new tests plus the full existing `cmd/projectcontext_test.go` suite unmodified, confirm all green (this proves the extraction preserved behavior exactly - #88's existing tests are the regression net).

- [x] **Step 9 - Failing tests: `Registry.Dispatch` gate-consulting behavior**
  Add cases to `tools/registry_test.go` using fake `Tool`/`PermissionGate` implementations: a non-destructive tool's `Execute` is called without ever touching the gate (BR5); a destructive tool with a `ConfirmationSummary` error returns an error `ToolResultBlock` without calling the gate or `Execute` (BR3-BR4); a gate returning `DecisionDeny` returns a declined-error `ToolResultBlock` without calling `Execute` (BR11); a gate returning any Allow variant calls `Execute` and returns its result. Confirm compile failure (signature doesn't match yet).

- [x] **Step 10 - Implement `Registry.Dispatch`'s extended signature**
  Change `Dispatch(ctx, call, gate PermissionGate) types.ToolResultBlock`, insert the gate-consulting logic from `business-logic-model.md`'s data-flow section. Run Step 9 tests, confirm green.

- [x] **Step 11 - Failing tests: `InteractivePermissionGate` prompt parsing**
  Create `cmd/permissionprompt_test.go` with cases using injected `strings.Reader`/`bytes.Buffer` (per Application Design's testability decision): each of `o`/`s`/`a`/`n` (case-insensitive) parses to the right `Decision`; empty input/EOF/unrecognized input defaults to deny (BR12); a session choice calls `store.RecordSession`; an always choice calls `store.RecordAlways` only when a repo root is available, and "always" isn't even offered in the prompt text when it isn't (BR10). Confirm compile failure.

- [x] **Step 12 - Implement `InteractivePermissionGate`**
  Create `cmd/permissionprompt.go` per `domain-entities.md`'s shape. Run Step 11 tests, confirm green.

- [x] **Step 13 - Thread the gate through `runChatTurnWithTools` and its call sites**
  Add a `gate tools.PermissionGate` parameter to `runChatTurnWithTools` (`cmd/toolloop.go`), pass it through to the `registry.Dispatch(ctx, call, gate)` call. Update `cmd/chat.go`'s registry-construction block: build an `ApprovalStore` (via `os.Getwd()` + `utils.FindGitBoundary` + `fm.ConfigPath`) and an `InteractivePermissionGate` (`os.Stdin`/`os.Stdout`) alongside the existing `registry := tools.NewRegistry()` line, pass `gate` into both `runChatTurnWithTools` call sites. Update `cmd/toolturn_test.go`'s 2 call sites with a no-op fake gate (mirroring how `onReasoning` no-op callbacks were added in Initiative 1 Unit 5).

- [x] **Step 14 - Full verification**
  `make test`, `make lint`, `make test-coverage` (confirm no regression in `tools`/`cmd`/`utils`), `make cli && go test -tags=integration -v .`. No behavior change expected yet from a user's perspective (`read_file` is still the only registered tool, still non-destructive, still `--tools`-gated) - this step confirms the plumbing is correct and inert, ready for Unit 7 to actually exercise it.
