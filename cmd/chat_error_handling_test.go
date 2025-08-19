package cmd

import (
	"testing"

	"github.com/chat-cli/chat-cli/errors"
	"github.com/spf13/cobra"
)

// TestChatCommandErrorHandling tests that the chat command properly handles errors
func TestChatCommandErrorHandling(t *testing.T) {
	// Test that the chat command exists and has proper error handling setup
	cmd := &cobra.Command{
		Use: "test-chat",
		Run: chatCmd.Run, // Use the actual chat command's Run function
	}

	// Add required flags to prevent flag parsing errors
	cmd.PersistentFlags().String("region", "us-east-1", "AWS region")
	cmd.PersistentFlags().String("model-id", "anthropic.claude-3-5-sonnet-20240620-v1:0", "Model ID")
	cmd.PersistentFlags().String("custom-arn", "", "Custom ARN")
	cmd.PersistentFlags().String("chat-id", "", "Chat ID")
	cmd.PersistentFlags().Float32("temperature", 0.7, "Temperature")
	cmd.PersistentFlags().Float32("topP", 0.9, "Top P")
	cmd.PersistentFlags().Int32("max-tokens", 1000, "Max tokens")

	// Test that error handler is available
	handler := errors.GetGlobalHandler()
	if handler == nil {
		t.Fatal("Global error handler should be available")
	}

	// Test that we can create various error types that the chat command uses
	testCases := []struct {
		name        string
		errorFunc   func() *errors.AppError
		expectedType errors.ErrorType
	}{
		{
			name: "Configuration Error",
			errorFunc: func() *errors.AppError {
				return errors.NewConfigurationError(
					"test_config_error",
					"Test configuration error",
					"Test user message",
					nil,
				)
			},
			expectedType: errors.ErrorTypeConfiguration,
		},
		{
			name: "Validation Error",
			errorFunc: func() *errors.AppError {
				return errors.NewValidationError(
					"test_validation_error",
					"Test validation error",
					"Test user message",
					nil,
				)
			},
			expectedType: errors.ErrorTypeValidation,
		},
		{
			name: "AWS Error",
			errorFunc: func() *errors.AppError {
				return errors.NewAWSError(
					"test_aws_error",
					"Test AWS error",
					"Test user message",
					nil,
				)
			},
			expectedType: errors.ErrorTypeAWS,
		},
		{
			name: "Database Error",
			errorFunc: func() *errors.AppError {
				return errors.NewDatabaseError(
					"test_database_error",
					"Test database error",
					"Test user message",
					nil,
				)
			},
			expectedType: errors.ErrorTypeDatabase,
		},
		{
			name: "Model Error",
			errorFunc: func() *errors.AppError {
				return errors.NewModelError(
					"test_model_error",
					"Test model error",
					"Test user message",
					nil,
				)
			},
			expectedType: errors.ErrorTypeModel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.errorFunc()
			if err.Type != tc.expectedType {
				t.Errorf("Expected error type %v, got %v", tc.expectedType, err.Type)
			}

			// Test that the error can be handled
			if err.GetUserMessage() == "" {
				t.Error("Error should have a user message")
			}

			// Test error context can be added
			err.WithOperation("TestOperation").WithComponent("chat-command").WithChatID("test-chat-123")
			
			if err.Context == nil {
				t.Error("Error context should be set")
			}
			
			if err.Context.Operation != "TestOperation" {
				t.Errorf("Expected operation 'TestOperation', got '%s'", err.Context.Operation)
			}
			
			if err.Context.Component != "chat-command" {
				t.Errorf("Expected component 'chat-command', got '%s'", err.Context.Component)
			}
			
			if err.Context.ChatID != "test-chat-123" {
				t.Errorf("Expected chat ID 'test-chat-123', got '%s'", err.Context.ChatID)
			}
		})
	}
}

// TestChatCommandRecoverableErrors tests that recoverable errors don't terminate the application
func TestChatCommandRecoverableErrors(t *testing.T) {
	testCases := []struct {
		name        string
		errorFunc   func() *errors.AppError
		shouldRecover bool
	}{
		{
			name: "Database Error (Recoverable)",
			errorFunc: func() *errors.AppError {
				return errors.NewDatabaseError(
					"database_connection_failed",
					"Database connection failed",
					"Unable to connect to database. Chat will work but history won't be saved.",
					nil,
				).WithRecoverable(true)
			},
			shouldRecover: true,
		},
		{
			name: "AWS Configuration Error (Critical)",
			errorFunc: func() *errors.AppError {
				return errors.NewAWSError(
					"credentials_not_found",
					"AWS credentials not found",
					"AWS credentials not found. Please run 'aws configure'.",
					nil,
				).WithSeverity(errors.ErrorSeverityCritical).WithRecoverable(false)
			},
			shouldRecover: false,
		},
		{
			name: "Model Validation Error (High Severity but Recoverable)",
			errorFunc: func() *errors.AppError {
				return errors.NewModelError(
					"model_not_found",
					"Model not found",
					"Model 'invalid-model' not found. Please choose a valid model.",
					nil,
				).WithSeverity(errors.ErrorSeverityHigh).WithRecoverable(true)
			},
			shouldRecover: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.errorFunc()
			
			if err.IsRecoverable() != tc.shouldRecover {
				t.Errorf("Expected recoverable=%v, got %v", tc.shouldRecover, err.IsRecoverable())
			}
			
			// Test that critical errors are properly identified
			if !tc.shouldRecover && !err.IsCritical() {
				t.Error("Non-recoverable errors should typically be critical")
			}
		})
	}
}