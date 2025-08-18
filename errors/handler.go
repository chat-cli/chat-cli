package errors

import (
	"fmt"
	"log"
	"os"
)

// ErrorHandler interface for handling different error scenarios
type ErrorHandler interface {
	Handle(err *AppError) error
	SetVerbose(verbose bool)
	SetDebug(debug bool)
	SetLogger(logger Logger)
}

// Logger interface for structured logging
type Logger interface {
	Error(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// DefaultErrorHandler is the default implementation of ErrorHandler
type DefaultErrorHandler struct {
	verbose bool
	debug   bool
	logger  Logger
}

// NewDefaultErrorHandler creates a new instance of DefaultErrorHandler
func NewDefaultErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{
		verbose: false,
		debug:   false,
		logger:  &DefaultLogger{}, // Use default logger if none provided
	}
}

// Handle processes an AppError and determines the appropriate response
func (h *DefaultErrorHandler) Handle(err *AppError) error {
	if err == nil {
		return nil
	}

	// Log the error with appropriate level
	h.logError(err)

	// Display user message
	h.displayUserMessage(err)

	// If critical error, terminate the application
	if err.IsCritical() {
		os.Exit(1)
	}

	return err
}

// SetVerbose enables or disables verbose error reporting
func (h *DefaultErrorHandler) SetVerbose(verbose bool) {
	h.verbose = verbose
}

// SetDebug enables or disables debug mode
func (h *DefaultErrorHandler) SetDebug(debug bool) {
	h.debug = debug
}

// SetLogger sets a custom logger
func (h *DefaultErrorHandler) SetLogger(logger Logger) {
	h.logger = logger
}

// logError logs the error with appropriate detail level
func (h *DefaultErrorHandler) logError(err *AppError) {
	fields := []Field{
		{Key: "type", Value: err.Type.String()},
		{Key: "code", Value: err.Code},
		{Key: "severity", Value: err.Severity.String()},
		{Key: "recoverable", Value: err.Recoverable},
	}

	// Add context fields if available
	if err.Context != nil {
		if err.Context.Operation != "" {
			fields = append(fields, Field{Key: "operation", Value: err.Context.Operation})
		}
		if err.Context.Component != "" {
			fields = append(fields, Field{Key: "component", Value: err.Context.Component})
		}
		if err.Context.ChatID != "" {
			fields = append(fields, Field{Key: "chat_id", Value: err.Context.ChatID})
		}
		if err.Context.UserID != "" {
			fields = append(fields, Field{Key: "user_id", Value: err.Context.UserID})
		}
	}

	// Add metadata fields
	for key, value := range err.Metadata {
		fields = append(fields, Field{Key: key, Value: value})
	}

	// Add cause if available and in debug mode
	if h.debug && err.Cause != nil {
		fields = append(fields, Field{Key: "cause", Value: err.Cause.Error()})
	}

	// Log at appropriate level based on severity
	switch err.Severity {
	case ErrorSeverityCritical, ErrorSeverityHigh:
		h.logger.Error(err.Message, fields...)
	case ErrorSeverityMedium:
		h.logger.Warn(err.Message, fields...)
	case ErrorSeverityLow:
		h.logger.Info(err.Message, fields...)
	}
}

// displayUserMessage shows the user-friendly error message
func (h *DefaultErrorHandler) displayUserMessage(err *AppError) {
	userMsg := err.GetUserMessage()
	
	if h.verbose {
		// In verbose mode, show more technical details
		fmt.Fprintf(os.Stderr, "Error [%s:%s]: %s\n", err.Type.String(), err.Code, userMsg)
		
		if err.Context != nil && err.Context.Operation != "" {
			fmt.Fprintf(os.Stderr, "Operation: %s\n", err.Context.Operation)
		}
		
		if h.debug && err.Cause != nil {
			fmt.Fprintf(os.Stderr, "Technical details: %v\n", err.Cause)
		}
	} else {
		// In normal mode, show just the user-friendly message
		fmt.Fprintf(os.Stderr, "Error: %s\n", userMsg)
	}
}

// DefaultLogger is a simple logger implementation that uses the standard log package
type DefaultLogger struct{}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, fields ...Field) {
	l.logWithFields("ERROR", msg, fields...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string, fields ...Field) {
	l.logWithFields("WARN", msg, fields...)
}

// Info logs an info message
func (l *DefaultLogger) Info(msg string, fields ...Field) {
	l.logWithFields("INFO", msg, fields...)
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, fields ...Field) {
	l.logWithFields("DEBUG", msg, fields...)
}

// logWithFields logs a message with structured fields
func (l *DefaultLogger) logWithFields(level, msg string, fields ...Field) {
	logMsg := fmt.Sprintf("[%s] %s", level, msg)
	
	if len(fields) > 0 {
		logMsg += " |"
		for _, field := range fields {
			logMsg += fmt.Sprintf(" %s=%v", field.Key, field.Value)
		}
	}
	
	log.Println(logMsg)
}

// Global error handler instance
var globalHandler ErrorHandler = NewDefaultErrorHandler()

// SetGlobalHandler sets the global error handler
func SetGlobalHandler(handler ErrorHandler) {
	globalHandler = handler
}

// GetGlobalHandler returns the global error handler
func GetGlobalHandler() ErrorHandler {
	return globalHandler
}

// Handle is a convenience function that uses the global error handler
func Handle(err *AppError) error {
	return globalHandler.Handle(err)
}