# Execution Plan: Release Automation (#98)

## Scope
Single unit — DevOps / CI. No application logic changes beyond version ldflags.

## Deliverables
1. `cmd/version.go` — injectable `version` var (default `dev`)
2. `Makefile` — `-ldflags` for local builds
3. `.goreleaser.yaml` — ldflags + Homebrew tap token env
4. `.github/workflows/ci.yml` — test + lint on PR/main
5. `.github/workflows/tag-release.yml` — workflow_dispatch semver bump
6. `.github/workflows/release.yml` — GoReleaser on tag push
7. `docs/release.md` — maintainer guide
8. Delete `.goreleaser.yaml.backup`

## Implementation Order
1. Version ldflags (code + Makefile + goreleaser)
2. CI workflow
3. Tag + release workflows
4. Documentation
5. Verify locally: `make test`, `make cli`, `goreleaser release --snapshot` (if goreleaser installed)

## Post-merge Setup (manual, documented)
- Add repo secret `HOMEBREW_TAP_DEPLOY_KEY` (write deploy key on `homebrew-chat-cli`)
- Optional: enable branch protection requiring CI on `main`

## Out of Scope
- release-please / conventional commits automation
- Auto-tag on every push to main
- Changes to `homebrew-chat-cli` repo structure
