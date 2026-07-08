# Unplanned Prerequisite: AWS SDK Upgrade

**Trigger**: Discovered while starting Unit 3 (Prompt Caching, #83) that the pinned `github.com/aws/aws-sdk-go-v2/service/bedrockruntime` (v1.23.0) has no cache-point types at all (`CachePointBlock`, `ContentBlockMemberCachePoint`, `SystemContentBlockMemberCachePoint` don't exist). Bisected: v1.27.0 still lacks them, v1.28.0 has them. Also checked ahead and found v1.23.0 has zero reasoning-content types needed for Unit 5 (Extended Thinking, #85) — same wall, different unit.

**Decision** (user-confirmed): upgrade to latest (v1.55.0) now rather than the bare minimum, since Unit 5 needs a newer SDK anyway and staying current matches this whole initiative's goal.

## Changes
- `go.mod`: `github.com/aws/aws-sdk-go-v2/service/bedrockruntime` v1.23.0 → v1.55.0, with compatible bumps to `github.com/aws/aws-sdk-go-v2` (v1.32.6 → v1.42.1), `github.com/aws/smithy-go` (v1.22.1 → v1.27.3), and their internal transitive dependencies (`aws/protocol/eventstream`, `internal/configsources`, `internal/endpoints/v2`).
- **Side effect**: `go.mod`'s `go` directive moved from `1.23.4` to `1.24` with an explicit `toolchain go1.24.7` line — the newer SDK requires a newer Go toolchain to build. This environment already has go1.24.7 installed, but contributors on older Go versions will need to upgrade (or let `go`'s automatic toolchain download handle it, which needs network access).
- `README.md`, `CLAUDE.md`: updated stated Go version requirement (1.22.1/1.23.4 → 1.24+) to match, since both were already stale relative to `go.mod` even before this bump (previously flagged in issue #95).
- `github.com/aws/aws-sdk-go-v2/service/bedrock` (control plane, used for `GetFoundationModel`/`ListFoundationModels`) was left at v1.25.0 — unaffected by this change, no new capability from that service is used yet.

## Verification
- `go build ./...`: clean, no breaking API changes affected our usage (`GetFoundationModel`, `Converse`, `ConverseStream`, `InvokeModel`, `ListFoundationModels` all compile unchanged)
- `make test`: all packages pass, no regressions
- `make lint`: clean
- `go test -tags=integration -v .`: all 7 pass
- Scanned the SDK changelog between v1.23.0 and v1.55.0 for entries tagged breaking/removed/deprecated relevant to our usage — nothing concerning found (one breaking change bumps the SDK's own minimum Go version to 1.19, well below what this project already requires)
