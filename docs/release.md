(release)=
# Release Process

chat-cli uses **GitHub Actions** for continuous integration and **GoReleaser** for publishing releases. Releases update two repositories:

- [chat-cli/chat-cli](https://github.com/chat-cli/chat-cli) — binaries and GitHub Release assets
- [chat-cli/homebrew-chat-cli](https://github.com/chat-cli/homebrew-chat-cli) — Homebrew formula

## Continuous integration

Every pull request and push to `main` runs `.github/workflows/ci.yml`:

- `make test`
- `make lint`
- `make cli` + integration tests

## How to cut a release

Releases are **tag-driven**, not automatic on every merge to `main`.

### Option A — GitHub Actions (recommended)

1. Merge changes to `main` and confirm CI is green.
2. Open **Actions → Tag Release → Run workflow**.
3. Choose bump level: `patch`, `minor`, or `major`.
4. The workflow runs tests, tags locally, publishes with GoReleaser, then pushes the `v*` tag.
5. Manual tag pushes (Option B) still trigger the separate **Release** workflow.

### Option B — Manual tag

```bash
git checkout main
git pull
make test
git tag v0.5.4
git push origin v0.5.4
```

Pushing a `v*` tag triggers the Release workflow.

## Version strings

Released binaries embed the tag via GoReleaser ldflags (`chat-cli version` shows e.g. `v0.5.4`).

Local builds without ldflags report `dev`:

```bash
make cli
./bin/chat-cli version
# chat-cli dev, darwin/arm64
```

You do **not** need to edit `cmd/version.go` before releasing.

## Repository secrets

Configure these in **Settings → Secrets and variables → Actions**:

| Secret | Purpose |
|--------|---------|
| `GITHUB_TOKEN` | Provided automatically; publishes release assets to this repo |
| `HOMEBREW_TAP_DEPLOY_KEY` | SSH private key (write deploy key on `chat-cli/homebrew-chat-cli`) |

Without `HOMEBREW_TAP_DEPLOY_KEY`, GoReleaser still publishes GitHub Release binaries but cannot update the Homebrew tap.

### Homebrew tap deploy key (one-time setup)

Generate a dedicated key (no passphrase):

```bash
ssh-keygen -t ed25519 -C "goreleaser-homebrew-tap" -f ~/.ssh/chat-cli-homebrew-tap -N ""
```

Add the public key to the tap repo with write access:

```bash
gh repo deploy-key add ~/.ssh/chat-cli-homebrew-tap.pub \
  --repo chat-cli/homebrew-chat-cli \
  --title "GoReleaser release automation" \
  --allow-write
```

Store the private key as a repository secret:

```bash
gh secret set HOMEBREW_TAP_DEPLOY_KEY --repo chat-cli/chat-cli < ~/.ssh/chat-cli-homebrew-tap
```

## Local GoReleaser (optional)

```bash
# Snapshot build (no publish)
goreleaser release --snapshot --clean
```

## Related files

- `.goreleaser.yaml` — build targets, archives, Homebrew tap config
- `.github/workflows/ci.yml` — CI on PR/main
- `.github/workflows/tag-release.yml` — semver bump dispatch
- `.github/workflows/release.yml` — GoReleaser on tag push

Closes workflow design for [#98](https://github.com/chat-cli/chat-cli/issues/98).
