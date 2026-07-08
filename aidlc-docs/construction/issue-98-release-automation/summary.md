# Build and Test Summary — Issue #98 Release Automation

## Deliverables
- Version injection via `-ldflags` (`cmd/version.go`, `Makefile`, `.goreleaser.yaml`)
- `.github/workflows/ci.yml` — test/lint/integration on PR and main
- `.github/workflows/tag-release.yml` — workflow_dispatch semver bump
- `.github/workflows/release.yml` — GoReleaser on `v*` tag push
- `docs/release.md` — maintainer documentation
- Removed `.goreleaser.yaml.backup`

## AI-DLC Artifacts
- Requirements: `aidlc-docs/inception/requirements/issue-98-release-automation.md`
- Plan: `aidlc-docs/inception/plans/issue-98-execution-plan.md`

## Manual Setup Required After Merge
1. Add `HOMEBREW_TAP_GITHUB_TOKEN` secret (PAT with push to `chat-cli/homebrew-chat-cli`)
2. Optionally enable branch protection requiring CI on `main`

## Verification
- `make test` / `make lint` locally
- `make cli && ./bin/chat-cli version` → `dev`
- First live release: run Tag Release workflow with `patch` bump after secrets configured

Closes #98.
