package errors

import (
	"time"
)

// NewAppError creates a new AppError with the specified parameters
func NewAppError(errorType ErrorType, code, message, userMessage string, cause error) *AppError {
	return &AppError{
		Type:        errorType,
		Code:        code,
		Message:     message,
		UserMessage: userMessage,
		Cause:       cause,
		Context: &ErrorContext{
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		},
		Severity:    ErrorSeverityMedium, // Default severity
		Recoverable: true,                // Default to recoverable
		Metadata:    make(map[string]interface{}),
	}
}

// NewAWSError creates a new AWS-related error
func NewAWSError(code, message, userMessage string, cause error) *AppError {
	return NewAppError(ErrorTypeAWS, code, message, userMessage, cause).
		WithSeverity(ErrorSeverityHigh)
}

// NewDatabaseError creates a new database-related error
func NewDatabaseError(code, message, userMessage string, cause error) *AppError {
	return NewAppError(ErrorTypeDatabase, code, message, userMessage, cause).
		WithSeverity(ErrorSeverityMedium)
}

// NewConfigurationError creates a new configuration-related error
func NewConfigurationError(code, message, userMessage string, cause error) *AppError {
	return NewAppError(ErrorTypeConfiguration, code, message, userMessage, cause).
		WithSeverity(ErrorSeverityHigh)
}

// NewValidationError creates a new validation-related error
func NewValidationError(code, message, userMessage string, cause error) *AppError {
	return NewAppError(ErrorTypeValidation, code, message, userMessage, cause).
		WithSeverity(ErrorSeverityMedium).
		WithRecoverable(true)
}

// NewNetworkError creates a new network-related error
func NewNetworkError(code, message, userMessage string, cause error) *AppError {
	return NewAppError(ErrorTypeNetwork, code, message, userMessage, cause).
		WithSeverity(ErrorSeverityMedium).
		WithRecoverable(true)
}

// NewFileSystemError creates a new file system-related error
func NewFileSystemError(code, message, userMessage string, cause error) *AppError {
	return NewAppError(ErrorTypeFileSystem, code, message, userMessage, cause).
		WithSeverity(ErrorSeverityMedium)
}

// NewModelError creates a new model-related error
func NewModelError(code, message, userMessage string, cause error) *AppError {
	return NewAppError(ErrorTypeModel, code, message, userMessage, cause).
		WithSeverity(ErrorSeverityHigh)
}

// NewCriticalError creates a new critical error that should terminate the application
func NewCriticalError(errorType ErrorType, code, message, userMessage string, cause error) *AppError {
	return NewAppError(errorType, code, message, userMessage, cause).
		WithSeverity(ErrorSeverityCritical).
		WithRecoverable(false)
}

// WithSeverity sets the severity level of the error
func (ae *AppError) WithSeverity(severity ErrorSeverity) *AppError {
	ae.Severity = severity
	return ae
}

// WithRecoverable sets whether the error is recoverable
func (ae *AppError) WithRecoverable(recoverable bool) *AppError {
	ae.Recoverable = recoverable
	return ae
}

// WithOperation sets the operation context
func (ae *AppError) WithOperation(operation string) *AppError {
	if ae.Context == nil {
		ae.Context = &ErrorContext{
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}
	}
	ae.Context.Operation = operation
	return ae
}

// WithComponent sets the component context
func (ae *AppError) WithComponent(component string) *AppError {
	if ae.Context == nil {
		ae.Context = &ErrorContext{
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}
	}
	ae.Context.Component = component
	return ae
}

// WithChatID sets the chat ID context
func (ae *AppError) WithChatID(chatID string) *AppError {
	if ae.Context == nil {
		ae.Context = &ErrorContext{
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}
	}
	ae.Context.ChatID = chatID
	return ae
}

// WithUserID sets the user ID context
func (ae *AppError) WithUserID(userID string) *AppError {
	if ae.Context == nil {
		ae.Context = &ErrorContext{
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}
	}
	ae.Context.UserID = userID
	return ae
}