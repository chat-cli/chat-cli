package errors

import (
	"fmt"
	"time"
)

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
	ErrorTypeUnknown
)

// String returns the string representation of ErrorType
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeAWS:
		return "AWS"
	case ErrorTypeDatabase:
		return "Database"
	case ErrorTypeConfiguration:
		return "Configuration"
	case ErrorTypeValidation:
		return "Validation"
	case ErrorTypeNetwork:
		return "Network"
	case ErrorTypeFileSystem:
		return "FileSystem"
	case ErrorTypeModel:
		return "Model"
	default:
		return "Unknown"
	}
}

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	ErrorSeverityLow ErrorSeverity = iota
	ErrorSeverityMedium
	ErrorSeverityHigh
	ErrorSeverityCritical
)

// String returns the string representation of ErrorSeverity
func (es ErrorSeverity) String() string {
	switch es {
	case ErrorSeverityLow:
		return "Low"
	case ErrorSeverityMedium:
		return "Medium"
	case ErrorSeverityHigh:
		return "High"
	case ErrorSeverityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// ErrorContext provides additional context about where and when an error occurred
type ErrorContext struct {
	Operation   string                 `json:"operation"`   // What operation was being performed
	Component   string                 `json:"component"`   // Which component generated the error
	UserID      string                 `json:"user_id"`     // User context (if applicable)
	ChatID      string                 `json:"chat_id"`     // Chat session context (if applicable)
	Timestamp   time.Time              `json:"timestamp"`   // When the error occurred
	Metadata    map[string]interface{} `json:"metadata"`    // Additional context
}

// AppError represents a structured application error with context and user-friendly messages
type AppError struct {
	Type        ErrorType              `json:"type"`
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`      // Technical message for developers
	UserMessage string                 `json:"user_message"` // User-friendly message
	Cause       error                  `json:"-"`            // Original error (not serialized)
	Context     *ErrorContext          `json:"context"`
	Severity    ErrorSeverity          `json:"severity"`
	Recoverable bool                   `json:"recoverable"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Error implements the error interface
func (ae *AppError) Error() string {
	if ae.Cause != nil {
		return fmt.Sprintf("[%s:%s] %s: %v", ae.Type.String(), ae.Code, ae.Message, ae.Cause)
	}
	return fmt.Sprintf("[%s:%s] %s", ae.Type.String(), ae.Code, ae.Message)
}

// Unwrap returns the underlying error for error unwrapping
func (ae *AppError) Unwrap() error {
	return ae.Cause
}

// GetUserMessage returns the user-friendly message, falling back to the technical message if not set
func (ae *AppError) GetUserMessage() string {
	if ae.UserMessage != "" {
		return ae.UserMessage
	}
	return ae.Message
}

// IsCritical returns true if the error is critical and should terminate the application
func (ae *AppError) IsCritical() bool {
	return ae.Severity == ErrorSeverityCritical
}

// IsRecoverable returns true if the error can be recovered from
func (ae *AppError) IsRecoverable() bool {
	return ae.Recoverable
}

// WithContext adds context to the error
func (ae *AppError) WithContext(ctx *ErrorContext) *AppError {
	ae.Context = ctx
	return ae
}

// WithMetadata adds metadata to the error
func (ae *AppError) WithMetadata(key string, value interface{}) *AppError {
	if ae.Metadata == nil {
		ae.Metadata = make(map[string]interface{})
	}
	ae.Metadata[key] = value
	return ae
}