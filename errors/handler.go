package errors

import (
	"fmt"
	"os"

	"github.com/chat-cli/chat-cli/logging"
)

// ErrorHandler interface for handling different error scenarios
type ErrorHandler interface {
	Handle(err *AppError) error
	SetVerbose(verbose bool)
	SetDebug(debug bool)
	SetLogger(logger logging.Logger)
}

// Import the logging package types
// Logger interface is now defined in the logging package

// DefaultErrorHandler is the default implementation of ErrorHandler
type DefaultErrorHandler struct {
	verbose bool
	debug   bool
	logger  logging.Logger
}

// NewDefaultErrorHandler creates a new instance of DefaultErrorHandler
func NewDefaultErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{
		verbose: false,
		debug:   false,
		logger:  logging.GetGlobalLogger(), // Use global logger from logging package
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
func (h *DefaultErrorHandler) SetLogger(logger logging.Logger) {
	h.logger = logger
}

// logError logs the error with appropriate detail level
func (h *DefaultErrorHandler) logError(err *AppError) {
	fields := []logging.Field{
		logging.NewField("type", err.Type.String()),
		logging.NewField("code", err.Code),
		logging.NewField("severity", err.Severity.String()),
		logging.NewField("recoverable", err.Recoverable),
	}

	// Add context fields if available
	if err.Context != nil {
		if err.Context.Operation != "" {
			fields = append(fields, logging.NewField("operation", err.Context.Operation))
		}
		if err.Context.Component != "" {
			fields = append(fields, logging.NewField("component", err.Context.Component))
		}
		if err.Context.ChatID != "" {
			fields = append(fields, logging.NewField("chat_id", err.Context.ChatID))
		}
		if err.Context.UserID != "" {
			fields = append(fields, logging.NewField("user_id", err.Context.UserID))
		}
	}

	// Add metadata fields
	for key, value := range err.Metadata {
		fields = append(fields, logging.NewField(key, value))
	}

	// Add cause if available and in debug mode
	if h.debug && err.Cause != nil {
		fields = append(fields, logging.NewField("cause", err.Cause.Error()))
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