package errors

import (
	"errors"
	"testing"
	"time"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeAWS, "AWS"},
		{ErrorTypeDatabase, "Database"},
		{ErrorTypeConfiguration, "Configuration"},
		{ErrorTypeValidation, "Validation"},
		{ErrorTypeNetwork, "Network"},
		{ErrorTypeFileSystem, "FileSystem"},
		{ErrorTypeModel, "Model"},
		{ErrorTypeUnknown, "Unknown"},
	}

	for _, test := range tests {
		if got := test.errorType.String(); got != test.expected {
			t.Errorf("ErrorType.String() = %v, want %v", got, test.expected)
		}
	}
}

func TestErrorSeverity_String(t *testing.T) {
	tests := []struct {
		severity ErrorSeverity
		expected string
	}{
		{ErrorSeverityLow, "Low"},
		{ErrorSeverityMedium, "Medium"},
		{ErrorSeverityHigh, "High"},
		{ErrorSeverityCritical, "Critical"},
	}

	for _, test := range tests {
		if got := test.severity.String(); got != test.expected {
			t.Errorf("ErrorSeverity.String() = %v, want %v", got, test.expected)
		}
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "error with cause",
			appError: &AppError{
				Type:    ErrorTypeAWS,
				Code:    "test_code",
				Message: "test message",
				Cause:   errors.New("underlying error"),
			},
			expected: "[AWS:test_code] test message: underlying error",
		},
		{
			name: "error without cause",
			appError: &AppError{
				Type:    ErrorTypeDatabase,
				Code:    "db_error",
				Message: "database failed",
			},
			expected: "[Database:db_error] database failed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.appError.Error(); got != test.expected {
				t.Errorf("AppError.Error() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := &AppError{
		Cause: originalErr,
	}

	if got := appErr.Unwrap(); got != originalErr {
		t.Errorf("AppError.Unwrap() = %v, want %v", got, originalErr)
	}
}

func TestAppError_GetUserMessage(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "with user message",
			appError: &AppError{
				Message:     "technical message",
				UserMessage: "user-friendly message",
			},
			expected: "user-friendly message",
		},
		{
			name: "without user message",
			appError: &AppError{
				Message: "technical message",
			},
			expected: "technical message",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.appError.GetUserMessage(); got != test.expected {
				t.Errorf("AppError.GetUserMessage() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestAppError_IsCritical(t *testing.T) {
	tests := []struct {
		name     string
		severity ErrorSeverity
		expected bool
	}{
		{"critical", ErrorSeverityCritical, true},
		{"high", ErrorSeverityHigh, false},
		{"medium", ErrorSeverityMedium, false},
		{"low", ErrorSeverityLow, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			appErr := &AppError{Severity: test.severity}
			if got := appErr.IsCritical(); got != test.expected {
				t.Errorf("AppError.IsCritical() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestAppError_IsRecoverable(t *testing.T) {
	tests := []struct {
		name        string
		recoverable bool
		expected    bool
	}{
		{"recoverable", true, true},
		{"not recoverable", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			appErr := &AppError{Recoverable: test.recoverable}
			if got := appErr.IsRecoverable(); got != test.expected {
				t.Errorf("AppError.IsRecoverable() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestAppError_WithContext(t *testing.T) {
	appErr := &AppError{}
	ctx := &ErrorContext{
		Operation: "test_operation",
		Component: "test_component",
		Timestamp: time.Now(),
	}

	result := appErr.WithContext(ctx)

	if result.Context != ctx {
		t.Errorf("AppError.WithContext() did not set context correctly")
	}

	if result != appErr {
		t.Errorf("AppError.WithContext() should return the same instance")
	}
}

func TestAppError_WithMetadata(t *testing.T) {
	appErr := &AppError{}
	key := "test_key"
	value := "test_value"

	result := appErr.WithMetadata(key, value)

	if result.Metadata == nil {
		t.Errorf("AppError.WithMetadata() did not initialize metadata map")
	}

	if result.Metadata[key] != value {
		t.Errorf("AppError.WithMetadata() = %v, want %v", result.Metadata[key], value)
	}

	if result != appErr {
		t.Errorf("AppError.WithMetadata() should return the same instance")
	}
}