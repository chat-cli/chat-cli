# Requirements: Release Automation (#98)

## Context
GitHub issue: https://github.com/chat-cli/chat-cli/issues/98

Brownfield DevOps change — no application feature changes. Automate CI and release using GitHub Actions + existing GoReleaser config, updating both `chat-cli/chat-cli` and `chat-cli/homebrew-chat-cli`.

## Functional Requirements

### FR1 — Continuous integration
- Run `make test` and `make lint` on every pull request and push to `main`.
- Failed CI blocks merge (enforced via branch protection — out of repo scope but workflow must exist).

### FR2 — Version injection
- Remove hardcoded version string from `cmd/version.go`.
- Inject version at build time via `-ldflags` from GoReleaser tags.
- Local `make cli` builds report `dev` when no version is injected.

### FR3 — Release pipeline
- Pushing a semver tag (`v*`) triggers GoReleaser to:
  - Build cross-platform binaries
  - Publish a GitHub Release
  - Push updated Homebrew formula to `chat-cli/homebrew-chat-cli`
- Release workflow runs tests before publishing.

### FR4 — One-click tagging
- Provide `workflow_dispatch` to bump semver (patch/minor/major), run tests, create/push tag.
- Tag push triggers FR3 — maintainers do not manually run GoReleaser locally.

## Non-Functional Requirements

### NFR1 — Secrets
- `GITHUB_TOKEN` (default) for release assets in this repo.
- `HOMEBREW_TAP_DEPLOY_KEY` — SSH private key for a write deploy key on `chat-cli/homebrew-chat-cli`.

### NFR2 — No release on every main merge
- Auto-releasing every merge produces too many versions; use explicit tag/dispatch instead.

### NFR3 — Documentation
- Document maintainer workflow in `docs/release.md`.

## Decisions (2026-07-08)
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Release trigger | Tag push `v*` | Standard GoReleaser pattern; pairs with dispatch workflow |
| Semver bump | Manual dispatch (patch/minor/major) | Avoids noisy releases on every merge |
| Version source | Git tag via GoReleaser `{{.Version}}` | Eliminates drift with `cmd/version.go` |

## Acceptance Criteria
- [ ] CI workflow runs on PR and `main`
- [ ] `chat-cli version` shows tag version in released binaries
- [ ] Tag dispatch creates tag after tests pass
- [ ] GoReleaser publishes release + updates Homebrew tap (requires secrets configured)
- [ ] `docs/release.md` documents the process
