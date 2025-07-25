# Testing Guide

This document describes the testing strategy and setup for the chat-cli project.

## Test Structure

The project uses Go's built-in testing framework with the following test categories:

### Unit Tests
- **Location**: `*_test.go` files alongside source code
- **Purpose**: Test individual functions and components in isolation
- **Coverage**: Core utilities, configuration management, repository layer, and command structure

### Integration Tests  
- **Location**: `integration_test.go` (root level)
- **Purpose**: Test CLI commands end-to-end
- **Requirements**: Built CLI binary (`./bin/chat-cli`)

## Running Tests

### All Tests
```bash
make test
```

### Unit Tests Only
```bash
go test ./... -short
```

### Integration Tests Only
```bash
# Requires built CLI binary
make cli
go test -tags=integration -v
```

### With Coverage
```bash
make test-coverage
# Opens coverage.html in browser to view detailed coverage report
```

### Benchmarks
```bash
make benchmark
```

## Test Components

### Utils Package (`utils/utils_test.go`)
Tests core utility functions:
- `DecodeImage()` - Base64 image decoding
- `ReadImage()` - File reading with security checks
- `LoadDocument()` - Document loading and formatting
- `ProcessStreamingOutput()` - AWS streaming response handling

### Config Package (`config/config_test.go`)
Tests configuration management:
- `NewFileManager()` - File manager initialization
- OS-specific path handling (Windows, macOS, Linux)
- Viper configuration setup
- Environment variable handling
- Configuration precedence (flags > config > defaults)

### Repository Package (`repository/chat_test.go`)
Tests database operations:
- Chat creation and retrieval
- Message listing and filtering
- Database connection handling
- SQLite operations with in-memory testing

### Command Package (`cmd/cmd_test.go`)
Tests CLI command structure:
- Command registration and hierarchy
- Flag inheritance and validation
- Argument validation
- Help text and descriptions

### Integration Tests (`integration_test.go`)
Tests full CLI functionality:
- Command execution with built binary
- Help text output verification
- Flag and argument validation
- Error handling for invalid inputs

## Test Utilities

### Mock Database
The repository tests use an in-memory SQLite database for fast, isolated testing:

```go
type MockDatabase struct {
    db *sql.DB
}

func setupTestDB(t *testing.T) *MockDatabase {
    db, err := sql.Open("sqlite3", ":memory:")
    // ... setup code
}
```

### Command Testing Helpers
Command tests use helper functions to execute commands and capture output:

```go
func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetErr(buf)
    cmd.SetArgs(args)
    
    err := cmd.Execute()
    return buf.String(), err
}
```

## Test Coverage Goals

- **Target**: 80%+ overall coverage
- **Critical paths**: 90%+ coverage for core utilities and repository layer
- **CLI commands**: Focus on structure and validation (AWS integration mocked)

## Continuous Integration

The project uses GitHub Actions for automated testing:

### Test Matrix
- Go versions: 1.21, 1.22, 1.23
- Platforms: Ubuntu (Linux)
- Test types: Unit, integration, linting, security

### Pipeline Steps
1. **Setup**: Go installation, dependency caching
2. **Code Quality**: `go vet`, `go fmt`, `golangci-lint`
3. **Testing**: Unit tests with race detection and coverage
4. **Security**: `gosec` security scanning
5. **Integration**: Full CLI testing with built binary
6. **Artifacts**: Coverage reports uploaded to Codecov

## Development Workflow

### Before Committing
```bash
# Run all quality checks
make lint
make test
make test-coverage

# Build and test CLI
make cli
go test -tags=integration -v
```

### Adding New Tests

1. **Unit tests**: Add `*_test.go` files alongside source code
2. **Integration tests**: Add test functions to `integration_test.go`
3. **Mocking**: Use in-memory databases and mock AWS services where possible
4. **Coverage**: Aim for comprehensive test coverage of new functionality

### Test Guidelines

- **Isolation**: Tests should not depend on external services (except integration tests)
- **Fast**: Unit tests should run quickly (< 1s per test)
- **Deterministic**: Tests should produce consistent results
- **Clear**: Test names should clearly describe what is being tested
- **Coverage**: Test both success and error paths

## AWS Service Testing

Since the CLI integrates with AWS Bedrock, testing follows these patterns:

### Unit Level
- Mock AWS SDK types and responses
- Test data transformation and error handling
- Focus on application logic, not AWS integration

### Integration Level  
- Test CLI command structure and validation
- Verify help text and argument parsing
- Skip actual AWS calls (would require credentials and real services)

### Manual Testing
- Full AWS integration requires manual testing with valid credentials
- Test with real Bedrock models for complete validation
- Use different AWS regions and model configurations

## Troubleshooting Tests

### Common Issues

1. **Integration tests fail**: Ensure CLI is built (`make cli`)
2. **Coverage low**: Add tests for untested functions and error paths
3. **Flaky tests**: Check for race conditions or external dependencies
4. **Slow tests**: Profile and optimize, consider using `testing.Short()`

### Debugging

```bash
# Verbose test output
go test ./... -v

# Run specific test
go test ./utils -run TestDecodeImage -v

# Test with race detection
go test ./... -race

# Generate detailed coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```