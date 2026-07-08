# Integration Test Instructions

## Purpose
chat-cli is a single binary, not a set of services — there's no service-to-service integration to test. "Integration" here means two things: (1) the existing binary-level integration test suite (`integration_test.go`), and (2) verifying the 5 units from this initiative actually **compose** correctly when their flags are combined in one invocation, since each unit was built and tested in isolation.

## Existing Binary-Level Integration Tests

### 1. Build the Binary First
```bash
make cli
```

### 2. Execute the Integration Test Suite
```bash
go test -tags=integration -v .
```

### 3. Expected Results
All 7 tests pass: `TestCLIVersion`, `TestCLIHelp`, `TestCLIConfigHelp`, `TestCLIPromptNoArgs`, `TestCLIImageNoArgs`, `TestCLIFlagsExist`, `TestCLIModelsSubcommands`. `TestCLIFlagsExist` was re-verified to still pass with all 5 units' new flags present (`--system`, `--tools`, `--thinking`, `--thinking-budget`, `--document` alongside the pre-existing `--region`/`--model-id`/`--custom-arn`).

## Cross-Unit Composition Scenarios (this Build and Test pass)

Since no AWS credentials are available in this environment, these scenarios verify flag parsing, config resolution, and local validation logic compose correctly across units up to the point of the actual Bedrock call — a real Bedrock round-trip was **not** possible to test here (see `build-and-test-summary.md`).

### Scenario 1: All flags visible together, no registration conflicts
- **Description**: Confirm 5 units' worth of new flags (`--system`, `--tools`, `--thinking`/`--thinking-budget`, `--document`) don't collide with each other or pre-existing flags when registered on the same commands.
- **Test Steps**: `chat-cli --help` and `chat-cli prompt --help`; grep for each flag name.
- **Expected Results**: Every flag appears exactly once, with its documented default.
- **Result**: ✅ Pass — all flags present, no duplicates.

### Scenario 2: `prompt` with Units 1+3+4+5 combined (`--system`, piped document caching, `--document`, `--thinking`)
- **Description**: One `prompt` invocation exercising system prompt resolution (Unit 1), a document attachment (Unit 4), and extended thinking (Unit 5) together — the maximum plausible flag combination for `prompt`.
- **Setup**: `chat-cli config set system-prompt "..."` (Unit 1's config layer), a local `test-doc.txt`.
- **Test Steps**: `chat-cli prompt "summarize" --system "override" --document test-doc.txt --thinking --thinking-budget 2048`
- **Expected Results**: No panic; execution proceeds through config resolution, document read/validation, and reasoning-config construction, failing only at the expected AWS-credentials boundary (this sandbox has none).
- **Result**: ✅ Pass — reached `error: operation error Bedrock: GetFoundationModel ... no EC2 IMDS role found` (the expected, clean failure point), no crash.

### Scenario 3: `chat` with Units 1+2+5 combined (`--system`, `--tools`, `--thinking`)
- **Description**: One `chat` invocation exercising system prompt (Unit 1), tool registry construction (Unit 2), and reasoning config (Unit 5) together.
- **Test Steps**: `chat-cli --system "test" --tools --thinking --thinking-budget 2048`
- **Expected Results**: No panic; same clean AWS-credentials failure as Scenario 2.
- **Result**: ✅ Pass.

### Cleanup
Scratch config/data directories and test files created for these scenarios were removed after verification; no artifacts committed to the repository.

## Not Covered By These Scenarios
A live round-trip against real Bedrock (tool dispatch actually being called by a model, a real cache hit/miss, actual reasoning output, a real document being summarized) requires AWS credentials this environment doesn't have. This is the same gap noted in every individual unit's summary — see `build-and-test-summary.md` for the consolidated list of what still needs real-credential verification.
