---
inclusion: always
---

# Chat-CLI Project Guide

This steering document provides comprehensive guidance for working with the chat-cli codebase, a Go-based CLI application for interacting with Amazon Bedrock LLMs.

## Project Overview

Chat-CLI is a terminal-based program built with Go and Cobra that provides an interface to Amazon Bedrock LLMs. It supports interactive chat sessions, one-shot prompts, image generation, and persistent configuration management.

**Key Technologies:**
- Go 1.23.4+
- Cobra (CLI framework)
- Viper (configuration management)
- SQLite (chat history persistence)
- AWS SDK v2 (Bedrock integration)
- Charm libraries (TUI components: bubbletea, bubbles, lipgloss)

## Architecture Principles

### Clean Separation of Concerns

The codebase follows a layered architecture:

1. **CLI Layer** (`/cmd/`) - Cobra commands and user interaction
2. **Configuration Layer** (`/config/`) - OS-specific config management with Viper
3. **Database Layer** (`/db/` and `/repository/`) - SQLite persistence with repository pattern
4. **Factory Layer** (`/factory/`) - Database factory pattern
5. **Utils Layer** (`/utils/`) - Utility functions for document loading

### Key Design Patterns

**Root Command Behavior:**
- Running `chat-cli` without arguments launches interactive chat directly
- All chat flags (--model-id, --custom-arn, --chat-id) work at root level
- Subcommands provide additional functionality (prompt, config, image, models)

**Configuration Precedence:**
1. CLI flags (highest priority)
2. Config file values (set via `config set`)
3. Built-in defaults (lowest priority)

**Model Selection Priority:**
1. `--custom-arn` flag (highest)
2. `--model-id` flag
3. Config file values
4. Default: `anthropic.claude-3-5-sonnet-20240620-v1:0`

## Development Workflow

### Building and Running

```bash
# Build the CLI binary
make

# Run from build directory
./bin/chat-cli <command> <args> <flags>

# Run directly with Go (development)
go run main.go <command> <args> <flags>
```

### Testing Requirements

**CRITICAL: Always follow Test-Driven Development (TDD)**

Before implementing any feature:
1. Write tests first that define expected behavior
2. Run tests and watch them fail
3. Write minimal code to make tests pass
4. Run full test suite to verify
5. Refactor while keeping tests green

**Test Commands:**
```bash
# Run all tests
make test

# Run tests in short mode (skip integration tests)
make test-short

# Run tests with coverage
make test-coverage

# Run integration tests (requires built CLI)
make cli && go test -tags=integration -v

# Run linting
make lint
```

**Coverage Goals:**
- New functions: 80%+ coverage
- Critical paths: 90%+ coverage
- Test both success and failure scenarios
- Include edge cases and boundary conditions

**Current Coverage:**
- Repository: 80.6%
- Config: 77.2%
- Utils: 46.6%
- CMD: 7.4%

### Code Quality Standards

**Before Committing:**
```bash
make test                    # Ensure all tests pass
make test-coverage          # Verify coverage hasn't decreased
make lint                   # Check code quality
make cli && go test -tags=integration -v  # Test CLI integration
```

## File Organization

### Core Files
- `main.go` - Entry point, delegates to cmd package
- `go.mod` / `go.sum` - Go module dependencies
- `Makefile` - Build and test automation
- `integration_test.go` - Integration tests

### Directory Structure
- `/cmd/` - All Cobra command implementations
- `/config/` - Configuration management
- `/db/` - Database layer and migrations
- `/repository/` - Repository pattern implementations
- `/factory/` - Factory pattern for database
- `/utils/` - Utility functions
- `/docs/` - Sphinx documentation (Python-based)
- `/bin/` - Compiled binaries (gitignored)
- `/dist/` - Release artifacts (gitignored)

### Command Files
- `root.go` - Base command, launches chat when no subcommands
- `chat.go` - Interactive chat sessions with persistence
- `chatList.go` - List recent chat sessions
- `prompt.go` - One-shot prompts with stdin support
- `config.go` - Configuration management (set/get/unset)
- `image.go` - Image generation commands
- `models.go` / `modelsList.go` - Model listing and management
- `version.go` - Version information

## Key Features and Implementation Details

### Chat Session Management
- Auto-saves all interactions to SQLite
- Sessions identified by UUID
- `chat list` shows 10 most recent sessions with timestamps
- `--chat-id` flag resumes specific conversations
- Works with root command: `chat-cli --chat-id <uuid>`

### Document Input
- Stdin piping: `cat file.go | chat-cli prompt "explain"`
- Wraps documents in `<document></document>` tags
- Optimized for Anthropic Claude models

### Image Support
- Attachments via `--image` flag
- Supports PNG/JPG, <5MB limit
- Model-dependent feature

### Streaming Responses
- Default behavior for supported models
- Disable with `--no-stream` flag (prompt only)
- Chat command requires streaming-capable models

### AWS Integration
- Direct AWS SDK v2 integration
- Supports foundation models and custom ARNs
- Region configuration (default: us-east-1)
- Requires AWS CLI configured with credentials
- Bedrock model access must be enabled in AWS Console

## Documentation Guidelines

**IMPORTANT: Never create documentation in project root**

### Existing Documentation Structure
- `docs/index.md` - Project overview and getting started
- `docs/setup.md` - Installation and setup
- `docs/usage.md` - Command usage and examples
- `docs/models.md` - Supported AI models
- `docs/marketplace.md` - AWS Marketplace integration
- `docs/testing.md` - Testing guide and best practices

### Documentation Rules
1. **Always update docs in `docs/` directory** - Never create `.md` files in root
2. **Prefer editing existing docs** over creating new files
3. **Use lowercase with hyphens** for new files (e.g., `user-guide.md`)
4. **Link from index.md** to ensure discoverability
5. **Build with Sphinx** - See `docs/requirements.txt` for Python dependencies

## Common Development Tasks

### Adding a New Command
1. Create `cmd/newcommand.go` with Cobra command structure
2. Add command to root in `cmd/root.go` init function
3. Write tests in `cmd/newcommand_test.go`
4. Update integration tests if needed
5. Document in `docs/usage.md`
6. Run full test suite before committing

### Adding Configuration Options
1. Update `config/config.go` with new config keys
2. Add Viper bindings and validation
3. Write tests in `config/config_test.go`
4. Update `cmd/config.go` for CLI access
5. Document in `docs/setup.md` or `docs/usage.md`

### Modifying Database Schema
1. Create migration in `db/migrations.go`
2. Update repository interfaces in `/repository/`
3. Write comprehensive tests
4. Test migration path from previous versions
5. Document breaking changes if any

### Adding Model Support
1. Update model list in `cmd/models.go`
2. Add model-specific handling if needed
3. Test with actual AWS Bedrock access
4. Document in `docs/models.md`
5. Update README.md if it's a significant addition

## Testing Patterns

### Unit Tests
- Place in `*_test.go` files alongside code
- Use Go's `testing` package
- Mock external dependencies (AWS SDK, database)
- Test one thing at a time with descriptive names

### Integration Tests
- Located in `integration_test.go`
- Require built CLI binary
- Test end-to-end command execution
- Use build tag: `// +build integration`

### Test Organization
```go
func TestFeatureName(t *testing.T) {
    // Arrange - setup test data
    // Act - execute the code
    // Assert - verify results
}
```

## Release Process

The project uses GoReleaser for automated releases:
- Configuration in `.goreleaser.yaml`
- Builds binaries for multiple OS/architecture combinations
- Publishes to GitHub releases
- Homebrew tap: `chat-cli/chat-cli`

## AWS Requirements

Users need:
1. AWS account with credentials configured
2. AWS CLI installed and configured (`aws config`)
3. Bedrock model access enabled in AWS Console
4. Region-specific model availability varies

## Common Pitfalls to Avoid

1. **Don't create markdown files in root** - Use `docs/` directory
2. **Don't skip tests** - Always write tests before implementation
3. **Don't decrease coverage** - Maintain or improve test coverage
4. **Don't forget integration tests** - Update when adding commands
5. **Don't hardcode AWS regions** - Use configuration system
6. **Don't ignore linting** - Run `make lint` before committing
7. **Don't break backward compatibility** - Consider migration paths

## Quick Reference

**Default Model:** `anthropic.claude-3-5-sonnet-20240620-v1:0`
**Default Region:** `us-east-1`
**Config Location:** OS-specific (managed by Viper)
**Database:** SQLite in user's config directory
**Entry Point:** `main.go` → `cmd.Execute()` → `rootCmd`

## Getting Help

- Use `--help` flag on any command
- Check `docs/` directory for detailed documentation
- Review test files for usage examples
- See `CLAUDE.md` for AI assistant guidance
