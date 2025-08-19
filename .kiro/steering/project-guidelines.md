---
inclusion: always
---

# Chat-CLI Project Guidelines

This steering document provides essential guidance for working with the chat-cli codebase, based on the project's CLAUDE.md documentation.

## Build and Development Commands

### Building the Project
```bash
# Build the CLI binary
make

# Alternative: Direct Go build
go build -o ./bin/chat-cli main.go
```

### Running the Application
```bash
# Run from build directory
./bin/chat-cli <command> <args> <flags>

# Run directly with Go (for development)
go run main.go <command> <args> <flags>
```

### Testing Requirements

**CRITICAL: Always Update Tests When Adding Features**

When adding new functionality or modifying existing code, you MUST:

1. **Add corresponding unit tests** in `*_test.go` files alongside your code
2. **Update integration tests** in `integration_test.go` if adding new commands or flags
3. **Run the full test suite** before committing: `make test && make test-coverage`
4. **Maintain or improve coverage** - don't let test coverage decrease
5. **Update test documentation** in `docs/testing.md` if adding new test patterns

**Test Commands:**
```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run integration tests (requires built CLI)
make cli && go test -tags=integration -v

# Run linting and formatting
make lint
```

**Coverage Goals:**
- **New functions**: Aim for 80%+ test coverage
- **Critical paths**: 90%+ coverage for core business logic
- **Error handling**: Test both success and failure scenarios
- **Edge cases**: Include boundary condition testing

## Architecture Overview

This is a Go CLI application built with Cobra that provides an interface to Amazon Bedrock LLMs.

### Core Components

**CLI Layer (`/cmd/`)**
- `root.go` - Base Cobra command that launches chat functionality directly
- `chat.go` - Interactive chat sessions with persistent storage
- `prompt.go` - One-shot prompt commands with stdin support
- `config.go` - Configuration management
- `image.go` - Image generation commands
- `models*.go` - Model listing and management

**Configuration (`/config/`)**
- `config.go` - OS-specific file management using Viper
- Handles config.yaml storage in user's config directory

**Database Layer (`/db/` and `/repository/`)**
- SQLite-based persistence for chat history
- Repository pattern with base and chat-specific implementations
- Migration system for database schema management

**Error Handling (`/errors/`)**
- Structured error handling with custom types
- User-friendly messages and contextual information
- Centralized error handler with severity levels

### Key Patterns

**Model Selection Priority:**
1. `--custom-arn` flag (highest)
2. `--model-id` flag
3. Config file values
4. Default: `anthropic.claude-3-5-sonnet-20240620-v1:0`

**Chat Session Management:**
- Auto-saves all chat interactions to SQLite
- `--chat-id` flag resumes specific conversations
- UUIDs track individual chat sessions

**Error Handling:**
- Use the `/errors` package for all error handling
- Wrap standard errors with appropriate error types
- Provide user-friendly messages with suggestions
- Use severity levels and recovery patterns

## Development Guidelines

### Code Quality
- Follow Go conventions and best practices
- Use the repository pattern for data access
- Implement proper error handling using the errors package
- Write comprehensive tests for all new functionality

### Documentation
- Update documentation in the `docs/` directory
- Never create `.md` files in project root
- Update existing docs rather than creating new files
- Ensure new docs are linked from index.md

### AWS Integration
- Use AWS SDK v2 for Bedrock services
- Support both foundation models and custom ARNs
- Handle region configuration properly
- Implement proper credential management

### File Organization
- `main.go` - Entry point
- `cmd/` - CLI commands
- `config/` - Configuration management
- `db/` and `repository/` - Data persistence
- `errors/` - Error handling infrastructure
- `factory/` - Factory patterns
- `utils/` - Utility functions
- `docs/` - Documentation

## Testing Strategy

### Unit Tests
- Test individual functions and methods
- Mock external dependencies
- Test both success and failure scenarios
- Aim for high coverage on business logic

### Integration Tests
- Test CLI commands end-to-end
- Verify flag parsing and configuration
- Test database operations
- Validate AWS integration (where possible)

### Error Handling Tests
- Test all error paths
- Verify user-friendly messages
- Test error recovery scenarios
- Validate severity levels and context

## Common Patterns

### Error Handling
```go
// Wrap AWS errors
if err != nil {
    appErr := errors.WrapAWSError(err, "GetFoundationModel")
    return errors.Handle(appErr)
}

// Create specific error types
dbErr := errors.NewDatabaseError("connection_failed", "DB connection failed", "", err).
    WithOperation("ConnectDatabase").
    WithChatID(chatID)
```

### Repository Pattern
```go
// Use repositories for data access
chatRepo := repository.NewChatRepository(database)
if err := chatRepo.Create(chat); err != nil {
    return errors.WrapDatabaseError(err, "CreateChat")
}
```

### Configuration Management
```go
// Use FileManager for config operations
fm, err := conf.NewFileManager("chat-cli")
if err != nil {
    return errors.WrapConfigurationError(err, "app_name", "NewFileManager")
}
```

This steering document should be referenced for all development work on the chat-cli project to ensure consistency and quality.