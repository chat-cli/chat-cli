package errors

import (
	"errors"
	"testing"
)

func TestNewAppError(t *testing.T) {
	cause := errors.New("underlying error")
	appErr := NewAppError(ErrorTypeAWS, "test_code", "test message", "user message", cause)

	if appErr.Type != ErrorTypeAWS {
		t.Errorf("NewAppError() Type = %v, want %v", appErr.Type, ErrorTypeAWS)
	}

	if appErr.Code != "test_code" {
		t.Errorf("NewAppError() Code = %v, want %v", appErr.Code, "test_code")
	}

	if appErr.Message != "test message" {
		t.Errorf("NewAppError() Message = %v, want %v", appErr.Message, "test message")
	}

	if appErr.UserMessage != "user message" {
		t.Errorf("NewAppError() UserMessage = %v, want %v", appErr.UserMessage, "user message")
	}

	if appErr.Cause != cause {
		t.Errorf("NewAppError() Cause = %v, want %v", appErr.Cause, cause)
	}

	if appErr.Severity != ErrorSeverityMedium {
		t.Errorf("NewAppError() Severity = %v, want %v", appErr.Severity, ErrorSeverityMedium)
	}

	if !appErr.Recoverable {
		t.Errorf("NewAppError() Recoverable = %v, want %v", appErr.Recoverable, true)
	}

	if appErr.Context == nil {
		t.Errorf("NewAppError() Context should not be nil")
	}

	if appErr.Metadata == nil {
		t.Errorf("NewAppError() Metadata should not be nil")
	}
}

func TestNewAWSError(t *testing.T) {
	cause := errors.New("aws error")
	appErr := NewAWSError("credentials_invalid", "AWS credentials invalid", "Please check your credentials", cause)

	if appErr.Type != ErrorTypeAWS {
		t.Errorf("NewAWSError() Type = %v, want %v", appErr.Type, ErrorTypeAWS)
	}

	if appErr.Severity != ErrorSeverityHigh {
		t.Errorf("NewAWSError() Severity = %v, want %v", appErr.Severity, ErrorSeverityHigh)
	}
}

func TestNewDatabaseError(t *testing.T) {
	cause := errors.New("db error")
	appErr := NewDatabaseError("connection_failed", "Database connection failed", "Cannot connect to database", cause)

	if appErr.Type != ErrorTypeDatabase {
		t.Errorf("NewDatabaseError() Type = %v, want %v", appErr.Type, ErrorTypeDatabase)
	}

	if appErr.Severity != ErrorSeverityMedium {
		t.Errorf("NewDatabaseError() Severity = %v, want %v", appErr.Severity, ErrorSeverityMedium)
	}
}

func TestNewConfigurationError(t *testing.T) {
	cause := errors.New("config error")
	appErr := NewConfigurationError("file_invalid", "Config file invalid", "Configuration file has errors", cause)

	if appErr.Type != ErrorTypeConfiguration {
		t.Errorf("NewConfigurationError() Type = %v, want %v", appErr.Type, ErrorTypeConfiguration)
	}

	if appErr.Severity != ErrorSeverityHigh {
		t.Errorf("NewConfigurationError() Severity = %v, want %v", appErr.Severity, ErrorSeverityHigh)
	}
}

func TestNewValidationError(t *testing.T) {
	cause := errors.New("validation error")
	appErr := NewValidationError("invalid_input", "Input validation failed", "Please check your input", cause)

	if appErr.Type != ErrorTypeValidation {
		t.Errorf("NewValidationError() Type = %v, want %v", appErr.Type, ErrorTypeValidation)
	}

	if appErr.Severity != ErrorSeverityMedium {
		t.Errorf("NewValidationError() Severity = %v, want %v", appErr.Severity, ErrorSeverityMedium)
	}

	if !appErr.Recoverable {
		t.Errorf("NewValidationError() Recoverable = %v, want %v", appErr.Recoverable, true)
	}
}

func TestNewNetworkError(t *testing.T) {
	cause := errors.New("network error")
	appErr := NewNetworkError("connection_timeout", "Network timeout", "Connection timed out", cause)

	if appErr.Type != ErrorTypeNetwork {
		t.Errorf("NewNetworkError() Type = %v, want %v", appErr.Type, ErrorTypeNetwork)
	}

	if appErr.Severity != ErrorSeverityMedium {
		t.Errorf("NewNetworkError() Severity = %v, want %v", appErr.Severity, ErrorSeverityMedium)
	}

	if !appErr.Recoverable {
		t.Errorf("NewNetworkError() Recoverable = %v, want %v", appErr.Recoverable, true)
	}
}

func TestNewFileSystemError(t *testing.T) {
	cause := errors.New("fs error")
	appErr := NewFileSystemError("file_not_found", "File not found", "The file could not be found", cause)

	if appErr.Type != ErrorTypeFileSystem {
		t.Errorf("NewFileSystemError() Type = %v, want %v", appErr.Type, ErrorTypeFileSystem)
	}

	if appErr.Severity != ErrorSeverityMedium {
		t.Errorf("NewFileSystemError() Severity = %v, want %v", appErr.Severity, ErrorSeverityMedium)
	}
}

func TestNewModelError(t *testing.T) {
	cause := errors.New("model error")
	appErr := NewModelError("not_found", "Model not found", "The specified model was not found", cause)

	if appErr.Type != ErrorTypeModel {
		t.Errorf("NewModelError() Type = %v, want %v", appErr.Type, ErrorTypeModel)
	}

	if appErr.Severity != ErrorSeverityHigh {
		t.Errorf("NewModelError() Severity = %v, want %v", appErr.Severity, ErrorSeverityHigh)
	}
}

func TestNewCriticalError(t *testing.T) {
	cause := errors.New("critical error")
	appErr := NewCriticalError(ErrorTypeAWS, "fatal_error", "Fatal error occurred", "A critical error occurred", cause)

	if appErr.Severity != ErrorSeverityCritical {
		t.Errorf("NewCriticalError() Severity = %v, want %v", appErr.Severity, ErrorSeverityCritical)
	}

	if appErr.Recoverable {
		t.Errorf("NewCriticalError() Recoverable = %v, want %v", appErr.Recoverable, false)
	}
}

func TestAppError_WithSeverity(t *testing.T) {
	appErr := NewAppError(ErrorTypeAWS, "test", "test", "test", nil)
	result := appErr.WithSeverity(ErrorSeverityCritical)

	if result.Severity != ErrorSeverityCritical {
		t.Errorf("WithSeverity() Severity = %v, want %v", result.Severity, ErrorSeverityCritical)
	}

	if result != appErr {
		t.Errorf("WithSeverity() should return the same instance")
	}
}

func TestAppError_WithRecoverable(t *testing.T) {
	appErr := NewAppError(ErrorTypeAWS, "test", "test", "test", nil)
	result := appErr.WithRecoverable(false)

	if result.Recoverable {
		t.Errorf("WithRecoverable() Recoverable = %v, want %v", result.Recoverable, false)
	}

	if result != appErr {
		t.Errorf("WithRecoverable() should return the same instance")
	}
}

func TestAppError_WithOperation(t *testing.T) {
	appErr := NewAppError(ErrorTypeAWS, "test", "test", "test", nil)
	operation := "test_operation"
	result := appErr.WithOperation(operation)

	if result.Context.Operation != operation {
		t.Errorf("WithOperation() Operation = %v, want %v", result.Context.Operation, operation)
	}

	if result != appErr {
		t.Errorf("WithOperation() should return the same instance")
	}
}

func TestAppError_WithComponent(t *testing.T) {
	appErr := NewAppError(ErrorTypeAWS, "test", "test", "test", nil)
	component := "test_component"
	result := appErr.WithComponent(component)

	if result.Context.Component != component {
		t.Errorf("WithComponent() Component = %v, want %v", result.Context.Component, component)
	}

	if result != appErr {
		t.Errorf("WithComponent() should return the same instance")
	}
}

func TestAppError_WithChatID(t *testing.T) {
	appErr := NewAppError(ErrorTypeAWS, "test", "test", "test", nil)
	chatID := "test_chat_id"
	result := appErr.WithChatID(chatID)

	if result.Context.ChatID != chatID {
		t.Errorf("WithChatID() ChatID = %v, want %v", result.Context.ChatID, chatID)
	}

	if result != appErr {
		t.Errorf("WithChatID() should return the same instance")
	}
}

func TestAppError_WithUserID(t *testing.T) {
	appErr := NewAppError(ErrorTypeAWS, "test", "test", "test", nil)
	userID := "test_user_id"
	result := appErr.WithUserID(userID)

	if result.Context.UserID != userID {
		t.Errorf("WithUserID() UserID = %v, want %v", result.Context.UserID, userID)
	}

	if result != appErr {
		t.Errorf("WithUserID() should return the same instance")
	}
}