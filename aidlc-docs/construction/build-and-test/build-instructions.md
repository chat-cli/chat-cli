# Build Instructions

## Prerequisites
- **Build Tool**: Go 1.24+ (bumped from 1.23.4 during Unit 3's prerequisite SDK upgrade — see `aidlc-docs/construction/sdk-upgrade/summary.md`)
- **Dependencies**: Managed via Go modules (`go.mod`/`go.sum`) — no manual dependency installation needed beyond `go build`/`go mod download`
- **Environment Variables**: None required to build. `AWS_REGION`/AWS credentials are only needed to *run* the CLI against real Bedrock, not to build it.
- **System Requirements**: Any platform Go 1.24 supports (this project cross-compiles for multiple OS/arch via GoReleaser); no special memory/disk requirements.

## Build Steps

### 1. Install Dependencies
```bash
go mod download
```

### 2. Configure Environment
No environment configuration is needed to build. To *run* the built binary against real Bedrock, configure AWS credentials via `aws configure` (see `README.md` Prerequisites).

### 3. Build the Binary
```bash
make cli
# equivalent to: go build -o ./bin/chat-cli main.go
```

### 4. Verify Build Success
- **Expected Output**: `go build -o ./bin/chat-cli main.go` with no errors, producing `./bin/chat-cli`
- **Build Artifacts**: `./bin/chat-cli` (gitignored, not committed)
- **Common Warnings**: None expected — a clean `go build ./...` produces no output at all on success

## Troubleshooting

### Build Fails with Dependency Errors
- **Cause**: Stale or corrupted module cache
- **Solution**: `go clean -modcache && go mod download`

### Build Fails with Compilation Errors
- **Cause**: Should not occur on this branch — `go build ./...` was verified clean after every unit's code generation and again in this Build and Test pass
- **Solution**: If it does occur, check `go version` is 1.24+ (required since the Unit 3 SDK upgrade)
