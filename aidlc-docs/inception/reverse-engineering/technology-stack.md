# Technology Stack

## Programming Languages
- Go - 1.23.4 (per `go.mod`; README build instructions mention 1.22.1 as the minimum — see code-quality-assessment.md for this drift)

## Frameworks
- Cobra (`github.com/spf13/cobra` v1.8.1) - CLI command/flag framework, root of the entire `cmd` package
- Viper (`github.com/spf13/viper` v1.19.0) - Configuration file management (YAML) with precedence support
- Bubble Tea (`github.com/charmbracelet/bubbletea` v1.3.6) - TUI event loop framework for the interactive input widget
- Bubbles (`github.com/charmbracelet/bubbles` v0.21.0) - Pre-built TUI components (`textarea`) used by the input widget
- Lip Gloss (`github.com/charmbracelet/lipgloss` v1.1.0) - Terminal styling (borders, colors) for the input widget

## Infrastructure
- Amazon Bedrock - Foundation model hosting; both control-plane (`bedrock`) and runtime/inference (`bedrockruntime`) APIs are used
- AWS SDK for Go v2 (`github.com/aws/aws-sdk-go-v2` v1.32.6 + `config` v1.28.6) - AWS credential resolution and service clients
- SQLite (via `modernc.org/sqlite` v1.38.2, pure-Go/CGO-free) - Local chat history persistence; no external database server
- GoReleaser (`.goreleaser.yaml`) - Cross-platform binary builds and Homebrew tap publishing
- ReadTheDocs / Sphinx (`.readthedocs.yaml`, `docs/`) - Hosted documentation site build

## Build Tools
- Go Modules - Dependency management (`go.mod`/`go.sum`)
- Make - Build/test/lint orchestration (`Makefile`)
- golangci-lint (config in `.golangci.yml`) - Extended static analysis (used in CI; not directly invoked by `make lint`, which runs plain `go vet` + `go fmt`)

## Testing Tools
- Go `testing` package (standard library) - All unit and integration tests
- Go build tags (`integration`) - Separates integration tests (`integration_test.go`) from unit tests, gated behind a compiled binary
- `go tool cover` - Coverage report generation (`make test-coverage` → `coverage.out`/`coverage.html`)
