# Design Document

## Overview

The improved error handling system will replace the current ad-hoc error handling approach with a structured, user-friendly system that provides appropriate error messages for different user types and contexts. The design focuses on creating a centralized error handling framework that can gracefully degrade functionality while providing clear guidance to users.

The system will introduce custom error types, a centralized error handler, configurable logging levels, and early validation mechanisms to catch issues before they impact the user experience.

## Architecture

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Commands  │───▶│  Error Handler  │───▶│   User Output   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │     Logger      │
                       └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Log Files     │
                       └─────────────────┘
```

### Error Flow

1. **Error Detection**: Errors are caught at their source (AWS SDK, database, file operations, etc.)
2. **Error Classification**: Errors are wrapped with context and classified by type and severity
3. **Error Handling**: The centralized handler determines the appropriate response based on error type and user configuration
4. **User Communication**: User-friendly messages are displayed while technical details are logged
5. **Graceful Degradation**: Non-critical failures allow the application to continue with reduced functionality

## Components and Interfaces

### 1. Custom Error Types

```go
// ErrorType represents different categories of errors
type ErrorType int

const (
    ErrorTypeAWS ErrorType = iota
    ErrorTypeDatabase
    ErrorTypeConfiguration
    ErrorTypeValidation
    ErrorTypeNetwork
    ErrorTypeFileSystem
    ErrorTypeModel
)

// AppError represents a structured application error
type AppError struct {
    Type        ErrorType
    Code        string
    Message     string
    UserMessage string
    Cause       error
    Context     map[string]interface{}
    Recoverable bool
}

// ErrorHandler interface for handling different error scenarios
type ErrorHandler interface {
    Handle(err *AppError) error
    SetVerbose(verbose bool)
    SetDebug(debug bool)
}
```

### 2. Error Handler Implementation

The error handler will:
- Determine appropriate user messages based on error type
- Log technical details at appropriate levels
- Decide whether to terminate or continue execution
- Provide recovery suggestions where applicable

### 3. Validation Framework

```go
// Validator interface for early validation
type Validator interface {
    Validate() error
}

// ConfigValidator validates application configuration
type ConfigValidator struct {
    Config *config.FileManager
    AWSConfig aws.Config
}

// ModelValidator validates model availability and compatibility
type ModelValidator struct {
    ModelID string
    Region  string
    Client  *bedrock.Client
}
```

### 4. Enhanced Logging System

```go
// LogLevel represents different logging levels
type LogLevel int

const (
    LogLevelError LogLevel = iota
    LogLevelWarn
    LogLevelInfo
    LogLevelDebug
)

// Logger interface for structured logging
type Logger interface {
    Error(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Debug(msg string, fields ...Field)
    SetLevel(level LogLevel)
}
```

## Data Models

### Error Context Structure

```go
type ErrorContext struct {
    Operation   string                 // What operation was being performed
    Component   string                 // Which component generated the error
    UserID      string                 // User context (if applicable)
    ChatID      string                 // Chat session context (if applicable)
    Timestamp   time.Time              // When the error occurred
    Metadata    map[string]interface{} // Additional context
}
```

### Configuration Extensions

```go
type ErrorConfig struct {
    VerboseErrors bool   `yaml:"verbose_errors"`
    DebugMode     bool   `yaml:"debug_mode"`
    LogLevel      string `yaml:"log_level"`
    LogFile       string `yaml:"log_file"`
    MaxLogSize    int    `yaml:"max_log_size_mb"`
    MaxLogFiles   int    `yaml:"max_log_files"`
}
```

## Error Handling

### Error Classification Strategy

1. **Critical Errors**: Require immediate termination (invalid AWS credentials, missing required dependencies)
2. **Recoverable Errors**: Allow graceful degradation (chat history loading failure, non-critical config issues)
3. **User Errors**: Input validation failures that can be corrected (invalid model ID, malformed chat ID)
4. **System Errors**: Infrastructure issues that may be temporary (network timeouts, temporary AWS service issues)

### Recovery Mechanisms

- **Fallback Values**: Use defaults when configuration is invalid
- **Retry Logic**: Automatic retries for transient network issues
- **Alternative Paths**: Skip optional features when they fail
- **User Guidance**: Clear instructions for resolving user errors

### Error Message Templates

```go
var ErrorMessages = map[ErrorType]map[string]string{
    ErrorTypeAWS: {
        "credentials": "AWS credentials not found or invalid. Please run 'aws configure' or set AWS_PROFILE environment variable.",
        "region": "Invalid AWS region '%s'. Please check your region setting with 'aws configure get region'.",
        "permissions": "Insufficient AWS permissions for Bedrock. Please ensure your AWS user has bedrock:InvokeModel permissions.",
    },
    ErrorTypeModel: {
        "not_found": "Model '%s' not found. Use 'chat-cli models list' to see available models.",
        "not_text": "Model '%s' doesn't support text generation. Please choose a text-capable model.",
        "no_streaming": "Model '%s' doesn't support streaming. Please choose a streaming-capable model.",
    },
    // ... more error templates
}
```

## Testing Strategy

### Unit Testing

1. **Error Type Tests**: Verify custom error types are created correctly with proper context
2. **Handler Tests**: Test error handler responses for different error types and configurations
3. **Validation Tests**: Ensure validators catch invalid configurations and inputs
4. **Message Tests**: Verify user-friendly messages are generated correctly

### Integration Testing

1. **End-to-End Error Flows**: Test complete error scenarios from detection to user output
2. **Graceful Degradation**: Verify application continues working when non-critical components fail
3. **Configuration Testing**: Test error handling with various configuration states
4. **AWS Integration**: Test error handling with actual AWS service responses

### Error Simulation

1. **Mock AWS Errors**: Simulate various AWS SDK errors and responses
2. **Database Failures**: Test database connection and operation failures
3. **Network Issues**: Simulate network timeouts and connectivity problems
4. **File System Errors**: Test file permission and access issues

## Implementation Phases

### Phase 1: Core Error Infrastructure
- Implement custom error types and error handler
- Add basic validation framework
- Enhance logging system

### Phase 2: Command Integration
- Update all cobra commands to use new error handling
- Replace log.Fatal() calls with graceful error handling
- Add early validation to command initialization

### Phase 3: User Experience Enhancements
- Implement user-friendly error messages
- Add verbose and debug modes
- Create error recovery mechanisms

### Phase 4: Advanced Features
- Add structured logging with rotation
- Implement retry logic for transient errors
- Add error reporting and metrics collection

## Migration Strategy

The migration will be incremental to avoid breaking existing functionality:

1. **Introduce New Components**: Add error handling infrastructure alongside existing code
2. **Gradual Replacement**: Replace log.Fatal() calls one command at a time
3. **Backward Compatibility**: Ensure existing behavior is preserved during transition
4. **Testing at Each Step**: Validate each component replacement before proceeding

This approach ensures the application remains functional throughout the migration process while gradually improving the error handling experience.