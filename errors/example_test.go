package errors

import (
	"fmt"
	"testing"
)

// ExampleNewAWSError demonstrates how to use the error handling infrastructure
func ExampleNewAWSError() {
	// Create a simple AWS credentials error
	awsErr := NewAWSError(
		"credentials_not_found",
		"AWS credentials not configured",
		"", // Will be filled by EnhanceErrorWithMessage
		nil,
	).WithOperation("LoadAWSConfig").WithComponent("aws-config")

	// Enhance with user-friendly message
	awsErr = EnhanceErrorWithMessage(awsErr)

	// Handle the error
	handler := NewDefaultErrorHandler()
	handler.SetVerbose(true)
	
	fmt.Printf("Error Type: %s\n", awsErr.Type.String())
	fmt.Printf("Error Code: %s\n", awsErr.Code)
	fmt.Printf("User Message: %s\n", awsErr.GetUserMessage())
	fmt.Printf("Is Critical: %t\n", awsErr.IsCritical())
	fmt.Printf("Is Recoverable: %t\n", awsErr.IsRecoverable())
	
	// Output:
	// Error Type: AWS
	// Error Code: credentials_not_found
	// User Message: AWS credentials not found. Please run 'aws configure' or set the AWS_PROFILE environment variable.
	// Is Critical: false
	// Is Recoverable: true
}

// TestErrorHandlingWorkflow demonstrates a complete error handling workflow
func TestErrorHandlingWorkflow(t *testing.T) {
	// Simulate a database connection error
	dbErr := NewDatabaseError(
		"connection_failed",
		"Failed to connect to SQLite database",
		"",
		fmt.Errorf("database file not found"),
	).WithOperation("ConnectDatabase").
		WithComponent("database").
		WithChatID("test-chat-123")

	// Enhance with user-friendly message
	dbErr = EnhanceErrorWithMessage(dbErr)

	// Verify error properties
	if dbErr.Type != ErrorTypeDatabase {
		t.Errorf("Expected error type Database, got %s", dbErr.Type.String())
	}

	if dbErr.Code != "connection_failed" {
		t.Errorf("Expected error code connection_failed, got %s", dbErr.Code)
	}

	if dbErr.GetUserMessage() == "" {
		t.Error("Expected non-empty user message")
	}

	if dbErr.Context.Operation != "ConnectDatabase" {
		t.Errorf("Expected operation ConnectDatabase, got %s", dbErr.Context.Operation)
	}

	if dbErr.Context.ChatID != "test-chat-123" {
		t.Errorf("Expected chat ID test-chat-123, got %s", dbErr.Context.ChatID)
	}

	// Test error handler
	handler := NewDefaultErrorHandler()
	handler.SetVerbose(true)
	handler.SetDebug(true)

	// This would normally handle the error, but we don't want to exit in tests
	// handler.Handle(dbErr)
}

// TestErrorChaining demonstrates how to chain multiple errors
func TestErrorChaining(t *testing.T) {
	err1 := NewValidationError("invalid_model", "Model ID is invalid", "", nil)
	err2 := NewNetworkError("connection_timeout", "Network timeout", "", nil)
	err3 := NewConfigurationError("file_not_found", "Config file missing", "", nil)

	chainedErr := Chain(err1, err2, err3)

	if chainedErr == nil {
		t.Error("Expected non-nil chained error")
	}

	if chainedErr.Type != ErrorTypeValidation {
		t.Errorf("Expected primary error type Validation, got %s", chainedErr.Type.String())
	}

	if chainedErr.Code != "multiple_errors" {
		t.Errorf("Expected error code multiple_errors, got %s", chainedErr.Code)
	}

	errorCount, exists := chainedErr.Metadata["error_count"]
	if !exists {
		t.Error("Expected error_count in metadata")
	}

	if errorCount != 3 {
		t.Errorf("Expected error count 3, got %v", errorCount)
	}
}

// TestRetryableErrors demonstrates retry logic
func TestRetryableErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       *AppError
		retryable bool
	}{
		{
			name:      "network timeout is retryable",
			err:       NewNetworkError("connection_timeout", "timeout", "", nil),
			retryable: true,
		},
		{
			name:      "aws rate limit is retryable",
			err:       NewAWSError("rate_limited", "rate limited", "", nil),
			retryable: true,
		},
		{
			name:      "validation error is not retryable",
			err:       NewValidationError("invalid_input", "invalid", "", nil),
			retryable: false,
		},
		{
			name:      "model not found is not retryable",
			err:       NewModelError("not_found", "not found", "", nil),
			retryable: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsRetryableError(test.err); got != test.retryable {
				t.Errorf("IsRetryableError() = %v, want %v", got, test.retryable)
			}

			if test.retryable {
				delay := GetRetryDelay(test.err, 1)
				if delay <= 0 {
					t.Errorf("Expected positive retry delay, got %d", delay)
				}
			}
		})
	}
}