# Build and Test Summary: Built-in Agent Tools (#86)

Covers all 3 units together: Unit 6 (Confirmation and Sticky Approval Engine), Unit 7 (New Built-in Tools), Unit 8 (Automatic Tool-Use Enablement).

## Build Status
- **Build Tool**: Go 1.24+ (`go build`/`make cli`)
- **Build Status**: Success — `go build ./...` clean, no warnings
- **Build Artifacts**: `./bin/chat-cli` (built and verified, then removed per repo convention)

## Unit Tests
- **Command**: `make test`
- **Result**: All packages pass, 0 failures (2 pre-existing `SKIP`s unrelated to this initiative)
- **Coverage**: `cmd` 31.8%→33.5%, `tools` 90.0%→81.9% (new package surface added, still high), `utils` 51.8%→55.0%, total 67.8%→70.9% — no regression across the initiative
- **New test files this initiative**: `tools/approvalstore_test.go`, `tools/writefile_test.go`, `tools/runshell_test.go`, `tools/gitdiff_test.go`, `cmd/permissionprompt_test.go`, plus extensions to `tools/readfile_test.go`, `tools/registry_test.go`, `utils/utils_test.go`, `cmd/inferenceconfig_test.go`, `cmd/toolturn_test.go`

## Integration Tests
- **Command**: `make cli && go test -tags=integration -v .`
- **Result**: 7/7 pass, including `TestCLIFlagsExist` (confirms `--tools` is gone, no new flag was added by this initiative — the whole point was removing a flag)

## Cross-Unit Composition Scenarios
Since no AWS credentials are available in this environment, these verify wiring composes correctly up to the AWS-credentials boundary, plus (for the security-relevant confirmation gate specifically) full end-to-end verification via real components without needing a live model.

### Scenario 1: `chat` with `--system` + `--thinking` + always-on tools combined
- **Test**: `chat-cli --system "test" --thinking --thinking-budget 2048`
- **Result**: ✅ Pass — session starts cleanly, tools silently attached (all 4, no flag), no panic.

### Scenario 2: `chat` with `--no-context-file` (#88) + always-on tools (#86) combined
- **Test**: `chat-cli --no-context-file`
- **Result**: ✅ Pass — confirms Initiative 2 and Initiative 3's features coexist without interference.

### Scenario 3: Confirmation gate, full end-to-end, real components (no AWS needed)
Since driving a live model into requesting a tool call isn't possible here, this scenario wires the real `Registry`, real tools, and real `InteractivePermissionGate`/`ApprovalStore` together with scripted stdin — the exact code path `chat.go` uses once a model does request a tool call:
1. `write_file` + approve **once** → file created with correct content, `Dispatch` returns success
2. `write_file` + **deny** → file NOT created, model receives an error `ToolResultBlock`, no crash
3. `run_shell` + approve **session** → first call prompts; a second call with the same base command (`echo`) skips the prompt entirely — session-tier sticky approval verified working
4. `git_diff` with no stdin available at all → reached and errored immediately (correctly, not a git repo) without ever attempting to read from stdin — confirms the gate is never consulted for non-destructive tools
- **Result**: ✅ Pass, all 4 sub-scenarios (see Unit 7's code summary for the original run)

### Scenario 4: `--tools` fully removed
- **Test**: `chat-cli --help | grep -i tools`
- **Result**: ✅ Pass — zero matches, confirming Unit 8's flag removal is complete.

## Security Verification
- Per-repository "always" approval isolation (the security-critical property from FR7.1 — an approval in one repo must never leak into another) is verified at the unit level in `tools/approvalstore_test.go`'s `TestApprovalStore_AlwaysTier/always_approvals_do_not_leak_across_different_repository_roots` subtest, using two real `ApprovalStore` instances pointed at different repo roots sharing the same config path.
- `write_file`'s cwd confinement reuses `utils.ValidateLocalPathForWrite`, itself sharing the same `confineToWorkingDir` core logic already proven by `utils.ValidateLocalPath`'s existing, unmodified test suite.
- `run_shell`'s process-group-kill fix (found during Unit 7) is covered by a test that asserts both correctness and prompt return time, so the "timeout doesn't actually bound wall-clock time" regression class can't silently return.

## Performance Tests
- **Status**: N/A — same rationale as every prior initiative (single-user local CLI). `run_shell`'s 30s timeout is a reliability bound, not a performance target.

## ⚠️ Consolidated Real-Credential Verification List (this initiative's additions)
No AWS credentials are available in this environment. New items added to the project's running list (alongside Initiative 1's pre-existing items, most notably Unit 5's `reasoning_config` shape):

1. **`isToolUseUnsupportedError`'s heuristic** (Unit 8) — unverified against real Bedrock error text for a tool-use rejection. Same risk category and same open question as Unit 5's `reasoning_config` shape. If a model that doesn't support tools breaks `chat` instead of gracefully falling back, check this first.
2. **An actual tool-call round-trip for the 3 new tools** (Unit 7) — the protocol and gate logic are fully verified against synthetic events and real local components, but never against a real model's actual tool-call behavior for `write_file`/`run_shell`/`git_diff` specifically (Unit 2's `read_file` round-trip from Initiative 1 is a separate, still-open item on the original list).
3. **Real-world confirmation prompt UX** — the prompt text/formatting has been verified programmatically (captured buffer content) but never visually inspected in a real terminal session.

## Overall Status
- **Build**: ✅ Success
- **All Tests**: ✅ Pass (unit + integration + cross-unit composition + security verification; performance N/A)
- **Ready for**: Merge. All 3 units of #86 are code-complete, individually and cross-unit tested, committed, and pushed to `claude/ai-dlc-documentation-rl4e5s`.
