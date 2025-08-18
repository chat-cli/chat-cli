# Error Handling Package

This package provides a structured error handling system for the chat-cli application with user-friendly messages, contextual information, and graceful degradation capabilities.

## Features

- **Custom Error Types**: Structured errors with type classification (AWS, Database, Configuration, etc.)
- **Severity Levels**: Critical, High, Medium, and Low severity levels
- **User-Friendly Messages**: Automatic generation of user-friendly error messages with suggestions
- **Contextual Information**: Rich context including operation, component, chat ID, and metadata
- **Error Recovery**: Support for recoverable errors and retry logic
- **Centralized Handling**: Global error handler with configurable verbosity and debug modes

## Quick Start

### Creating Errors

```go
import "github.com/chat-cli/chat-cli/errors"

// Create an AWS credentials error
awsErr := errors.NewAWSError(
    "credentials_not_found",
    "AWS credentials not configured",
    "", // User message will be auto-generated
    nil,
).WithOperation("LoadAWSConfig").WithComponent("aws-config")

// Enhance with user-friendly message
awsErr = errors.EnhanceErrorWithMessage(awsErr)
```

### Handling Errors

```go
// Set up global error handler
handler := errors.NewDefaultErrorHandler()
handler.SetVerbose(true)  // Show technical details
handler.SetDebug(true)    // Show debug information
errors.SetGlobalHandler(handler)

// Handle an error
errors.Handle(awsErr)
```

### Wrapping Standard Errors

```go
// Wrap a standard error
if err != nil {
    appErr := errors.WrapAWSError(err, "GetFoundationModel")
    return errors.Handle(appErr)
}
```

## Error Types

- `ErrorTypeAWS`: AWS service related errors
- `ErrorTypeDatabase`: Database connection and query errors
- `ErrorTypeConfiguration`: Configuration file and setting errors
- `ErrorTypeValidation`: Input validation errors
- `ErrorTypeNetwork`: Network connectivity errors
- `ErrorTypeFileSystem`: File access and permission errors
- `ErrorTypeModel`: AI model related errors

## Severity Levels

- `ErrorSeverityCritical`: Application must terminate
- `ErrorSeverityHigh`: Significant impact, may require termination
- `ErrorSeverityMedium`: Moderate impact, recoverable
- `ErrorSeverityLow`: Minor impact, informational

## Builder Functions

### Basic Constructors
- `NewAppError(type, code, message, userMessage, cause)`: Generic error
- `NewAWSError(code, message, userMessage, cause)`: AWS-specific error
- `NewDatabaseError(code, message, userMessage, cause)`: Database error
- `NewConfigurationError(code, message, userMessage, cause)`: Config error
- `NewValidationError(code, message, userMessage, cause)`: Validation error
- `NewNetworkError(code, message, userMessage, cause)`: Network error
- `NewFileSystemError(code, message, userMessage, cause)`: File system error
- `NewModelError(code, message, userMessage, cause)`: Model error
- `NewCriticalError(type, code, message, userMessage, cause)`: Critical error

### Wrapper Functions
- `WrapError(err, type, code, message)`: Wrap any error
- `WrapAWSError(err, operation)`: Wrap AWS SDK errors
- `WrapDatabaseError(err, operation)`: Wrap database errors
- `WrapNetworkError(err, operation)`: Wrap network errors
- `WrapFileSystemError(err, filePath, operation)`: Wrap file system errors
- `WrapConfigurationError(err, configKey, operation)`: Wrap config errors

## Fluent API

Errors support method chaining for easy configuration:

```go
err := errors.NewAWSError("credentials_invalid", "Invalid credentials", "", cause).
    WithSeverity(errors.ErrorSeverityCritical).
    WithOperation("AuthenticateUser").
    WithComponent("auth").
    WithChatID("chat-123").
    WithMetadata("region", "us-east-1")
```

## Error Recovery and Retry

```go
// Check if error is retryable
if errors.IsRetryableError(err) {
    delay := errors.GetRetryDelay(err, attemptNumber)
    time.Sleep(time.Duration(delay) * time.Second)
    // Retry operation
}
```

## User Messages

The package includes comprehensive error message templates:

```go
// Get user-friendly message
userMsg := errors.GetUserMessage(errors.ErrorTypeAWS, "credentials_not_found")
// Returns: "AWS credentials not found. Please run 'aws configure' or set the AWS_PROFILE environment variable."

// Get suggestion
suggestion := errors.GetSuggestion(errors.ErrorTypeAWS, "credentials_not_found")
// Returns: "Run 'aws configure' to set up your credentials, or set the AWS_PROFILE environment variable."
```

## Global Error Handler

```go
// Configure global handler
handler := errors.NewDefaultErrorHandler()
handler.SetVerbose(true)
handler.SetDebug(true)
errors.SetGlobalHandler(handler)

// Use convenience function
errors.Handle(myError)
```

## Testing

The package includes comprehensive tests demonstrating usage patterns:

```bash
go test ./errors -v
```

## Integration

To integrate with existing code, gradually replace `log.Fatal()` calls:

```go
// Before
if err != nil {
    log.Fatal(err)
}

// After
if err != nil {
    appErr := errors.WrapAWSError(err, "LoadConfig")
    return errors.Handle(appErr)
}
```

This provides better user experience while maintaining the same error handling behavior.