# Unit Test Execution

## Run Unit Tests

### 1. Execute All Unit Tests
```bash
make test
# equivalent to: go test ./... -v
```

### 2. Review Test Results
- **Expected**: All tests pass, 0 failures (2 tests intentionally `SKIP` — `TestBubbleInput` and `TestStringPrompt` — both require live stdin/TTY interaction and are explicitly deferred to manual/integration testing, a pre-existing pattern from before this initiative)
- **Test Coverage**: Run `make test-coverage` for the full report (`coverage.out`/`coverage.html`)
  - `cmd`: 23.6% (up from 7.4% before this initiative)
  - `config`: 77.2% (unchanged)
  - `repository`: 77.5% (unchanged)
  - `tools`: 90.0% (new package, introduced in Unit 2)
  - `utils`: 48.2% (up from 46.6% before this initiative)
  - **Total**: 66.3% (up from 52.6% before this initiative)
- **Test Report Location**: `coverage.out`/`coverage.html` at the workspace root (gitignored, not committed)

### 3. Fix Failing Tests
No failing tests to fix as of this Build and Test pass — all 5 units were individually verified with `make test` before merging, and the full suite was re-run clean at the start of this stage.

## New Test Files Introduced This Initiative
- `cmd/systemprompt_test.go` (Unit 1)
- `tools/registry_test.go`, `tools/readfile_test.go`, `cmd/toolloop_test.go`, `cmd/streamaccumulate_test.go`, `cmd/toolturn_test.go` (Unit 2, `cmd/streamaccumulate_test.go` and `cmd/toolturn_test.go` extended again in Unit 5)
- `cmd/promptcache_test.go` (Unit 3)
- `cmd/documentinput_test.go`, plus new cases in `utils/utils_test.go` (Unit 4)
- `cmd/reasoning_test.go` (Unit 5)
