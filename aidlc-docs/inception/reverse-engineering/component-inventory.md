# Component Inventory

## Application Packages
- `cmd` - CLI command layer (root, chat, chat list, prompt, image, models, models list, config set/unset/list, version)
- `main` - Entry point (`main.go`), delegates to `cmd.Execute()`

## Infrastructure Packages
- None — this is a client application with no IaC/deployment packages in-repo. Release infrastructure is config-only: `.goreleaser.yaml` (binary + Homebrew tap publishing).

## Shared Packages
- `config` - Utilities/FileManager for OS-specific config & data paths, Viper-backed persisted settings
- `db` - Storage-agnostic interfaces (`Database`, `Migration`) and shared `Config` struct
- `db/sqlite` - SQLite driver implementation of the `db` interfaces
- `factory` - Database driver factory (`CreateDatabase`)
- `repository` - Generic `Repository[T]` interface + `ChatRepository` (Chat persistence)
- `utils` - Streaming response processing, image IO, stdin document loading, interactive terminal input widget (Clients: consumed by `cmd` for all Bedrock/image/input interactions)

## Test Packages
- `cmd` (`cmd_test.go`) - Unit tests for CLI command package
- `config` (`config_test.go`) - Unit tests for `FileManager`
- `repository` (`chat_test.go`) - Unit tests for `ChatRepository`
- `utils` (`utils_test.go`, `bubbleinput_test.go`) - Unit tests for utility functions and the input widget
- root (`integration_test.go`) - Build-tag `integration` end-to-end tests against the compiled `chat-cli` binary

## Total Count
- **Total Packages**: 8 (main, cmd, config, db, db/sqlite, factory, repository, utils)
- **Application**: 2 (main, cmd)
- **Infrastructure**: 0
- **Shared**: 5 (config, db, db/sqlite, factory, repository, utils — repository counted here as it's an internal persistence abstraction consumed by cmd)
- **Test**: Test files co-located within cmd, config, repository, utils, plus root-level `integration_test.go` (no separate test packages/directories)
