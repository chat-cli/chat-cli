package errors

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
)

// WrapError wraps a standard error into an AppError with the specified type and code
func WrapError(err error, errorType ErrorType, code, message string) *AppError {
	if err == nil {
		return nil
	}
	
	appErr := NewAppError(errorType, code, message, "", err)
	return EnhanceErrorWithMessage(appErr)
}

// WrapAWSError wraps an AWS SDK error into an AppError
func WrapAWSError(err error, operation string) *AppError {
	if err == nil {
		return nil
	}
	
	// Determine error code and user message based on error content
	errStr := err.Error()
	var code, userMessage string
	
	switch {
	case strings.Contains(errStr, "NoCredentialsProvided") || strings.Contains(errStr, "SharedConfigProfileNotExist"):
		code = "credentials_not_found"
	case strings.Contains(errStr, "InvalidUserID.NotFound") || strings.Contains(errStr, "SignatureDoesNotMatch"):
		code = "credentials_invalid"
	case strings.Contains(errStr, "InvalidRegion"):
		code = "region_invalid"
	case strings.Contains(errStr, "UnauthorizedOperation") || strings.Contains(errStr, "AccessDenied"):
		code = "permissions_denied"
	case strings.Contains(errStr, "ServiceUnavailable") || strings.Contains(errStr, "InternalError"):
		code = "service_unavailable"
	case strings.Contains(errStr, "Throttling") || strings.Contains(errStr, "RequestLimitExceeded"):
		code = "rate_limited"
	case strings.Contains(errStr, "RequestTimeout") || strings.Contains(errStr, "connection"):
		code = "connection_failed"
	default:
		code = "unknown"
		userMessage = fmt.Sprintf("AWS operation failed: %s", operation)
	}
	
	appErr := NewAWSError(code, fmt.Sprintf("AWS %s failed: %v", operation, err), userMessage, err).
		WithOperation(operation).
		WithComponent("aws")
	
	return EnhanceErrorWithMessage(appErr)
}

// WrapDatabaseError wraps a database error into an AppError
func WrapDatabaseError(err error, operation string) *AppError {
	if err == nil {
		return nil
	}
	
	errStr := err.Error()
	var code string
	
	switch {
	case strings.Contains(errStr, "no such file") || strings.Contains(errStr, "database is locked"):
		code = "connection_failed"
	case strings.Contains(errStr, "no such table") || strings.Contains(errStr, "syntax error"):
		code = "migration_failed"
	case strings.Contains(errStr, "no rows"):
		code = "not_found"
	case strings.Contains(errStr, "constraint"):
		code = "constraint_violation"
	default:
		code = "query_failed"
	}
	
	appErr := NewDatabaseError(code, fmt.Sprintf("Database %s failed: %v", operation, err), "", err).
		WithOperation(operation).
		WithComponent("database")
	
	return EnhanceErrorWithMessage(appErr)
}

// WrapNetworkError wraps a network error into an AppError
func WrapNetworkError(err error, operation string) *AppError {
	if err == nil {
		return nil
	}
	
	var code string
	
	// Check for specific network error types
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			code = "connection_timeout"
		} else {
			code = "connection_failed"
		}
	} else if strings.Contains(err.Error(), "no such host") {
		code = "dns_resolution"
	} else if strings.Contains(err.Error(), "connection refused") {
		code = "connection_refused"
	} else if strings.Contains(err.Error(), "tls") || strings.Contains(err.Error(), "ssl") {
		code = "ssl_error"
	} else {
		code = "connection_failed"
	}
	
	appErr := NewNetworkError(code, fmt.Sprintf("Network %s failed: %v", operation, err), "", err).
		WithOperation(operation).
		WithComponent("network")
	
	return EnhanceErrorWithMessage(appErr)
}

// WrapFileSystemError wraps a file system error into an AppError
func WrapFileSystemError(err error, filePath, operation string) *AppError {
	if err == nil {
		return nil
	}
	
	var code string
	
	if os.IsNotExist(err) {
		code = "file_not_found"
	} else if os.IsPermission(err) {
		code = "permission_denied"
	} else if strings.Contains(err.Error(), "no space left") {
		code = "disk_full"
	} else if strings.Contains(err.Error(), "file too large") {
		code = "file_too_large"
	} else {
		code = "access_failed"
	}
	
	appErr := NewFileSystemError(code, fmt.Sprintf("File %s failed for %s: %v", operation, filePath, err), "", err).
		WithOperation(operation).
		WithComponent("filesystem").
		WithMetadata("file_path", filePath)
	
	return EnhanceErrorWithMessage(appErr, filePath)
}

// WrapConfigurationError wraps a configuration error into an AppError
func WrapConfigurationError(err error, configKey, operation string) *AppError {
	if err == nil {
		return nil
	}
	
	errStr := err.Error()
	var code string
	
	switch {
	case strings.Contains(errStr, "not found") || os.IsNotExist(err):
		code = "file_not_found"
	case strings.Contains(errStr, "permission denied") || os.IsPermission(err):
		code = "permission_denied"
	case strings.Contains(errStr, "yaml") || strings.Contains(errStr, "unmarshal"):
		code = "file_invalid"
	case strings.Contains(errStr, "invalid") || strings.Contains(errStr, "parse"):
		code = "value_invalid"
	default:
		code = "load_failed"
	}
	
	appErr := NewConfigurationError(code, fmt.Sprintf("Configuration %s failed: %v", operation, err), "", err).
		WithOperation(operation).
		WithComponent("config").
		WithMetadata("config_key", configKey)
	
	return EnhanceErrorWithMessage(appErr, configKey)
}

// IsContextCanceled checks if an error is due to context cancellation
func IsContextCanceled(err error) bool {
	return err == context.Canceled || err == context.DeadlineExceeded
}

// IsConnectionError checks if an error is a connection-related error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for network errors
	if _, ok := err.(net.Error); ok {
		return true
	}
	
	// Check for syscall errors
	if opErr, ok := err.(*net.OpError); ok {
		if syscallErr, ok := opErr.Err.(*os.SyscallError); ok {
			errno := syscallErr.Err
			return errno == syscall.ECONNREFUSED || errno == syscall.ECONNRESET || errno == syscall.ETIMEDOUT
		}
	}
	
	errStr := strings.ToLower(err.Error())
	connectionKeywords := []string{
		"connection", "network", "timeout", "refused", "reset", "unreachable",
		"no route", "host", "dns", "resolve",
	}
	
	for _, keyword := range connectionKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	
	return false
}

// HandlePanic recovers from a panic and converts it to an AppError
func HandlePanic() *AppError {
	if r := recover(); r != nil {
		var err error
		if e, ok := r.(error); ok {
			err = e
		} else {
			err = fmt.Errorf("panic: %v", r)
		}
		
		return NewCriticalError(ErrorTypeUnknown, "panic", "Unexpected error occurred", "An unexpected error occurred. Please try again.", err).
			WithComponent("panic_handler")
	}
	return nil
}

// Must is a helper function that panics if an error is not nil
// This should only be used during initialization where errors are truly fatal
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustAppError is a helper function that panics if an AppError is not nil
func MustAppError(err *AppError) {
	if err != nil {
		panic(err)
	}
}

// Chain combines multiple errors into a single AppError
func Chain(errors ...*AppError) *AppError {
	var nonNilErrors []*AppError
	for _, err := range errors {
		if err != nil {
			nonNilErrors = append(nonNilErrors, err)
		}
	}
	
	if len(nonNilErrors) == 0 {
		return nil
	}
	
	if len(nonNilErrors) == 1 {
		return nonNilErrors[0]
	}
	
	// Create a combined error with the first error as the base
	primary := nonNilErrors[0]
	var messages []string
	
	for _, err := range nonNilErrors {
		messages = append(messages, err.GetUserMessage())
	}
	
	combinedMessage := strings.Join(messages, "; ")
	
	return NewAppError(primary.Type, "multiple_errors", "Multiple errors occurred", combinedMessage, primary.Cause).
		WithSeverity(primary.Severity).
		WithRecoverable(primary.Recoverable).
		WithMetadata("error_count", len(nonNilErrors))
}