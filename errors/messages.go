package errors

import (
	"fmt"
)

// ErrorMessages contains user-friendly error message templates
var ErrorMessages = map[ErrorType]map[string]string{
	ErrorTypeAWS: {
		"credentials_not_found": "AWS credentials not found. Please run 'aws configure' or set the AWS_PROFILE environment variable.",
		"credentials_invalid":   "AWS credentials are invalid. Please check your access key and secret key with 'aws configure'.",
		"region_invalid":        "Invalid AWS region '%s'. Please check your region setting with 'aws configure get region'.",
		"permissions_denied":    "Insufficient AWS permissions for Bedrock. Please ensure your AWS user has bedrock:InvokeModel permissions.",
		"service_unavailable":   "AWS Bedrock service is currently unavailable. Please try again later.",
		"rate_limited":          "AWS API rate limit exceeded. Please wait a moment and try again.",
		"connection_failed":     "Failed to connect to AWS services. Please check your internet connection and AWS configuration.",
	},
	ErrorTypeModel: {
		"not_found":       "Model '%s' not found. Use 'chat-cli models list' to see available models.",
		"not_text":        "Model '%s' doesn't support text generation. Please choose a text-capable model.",
		"no_streaming":    "Model '%s' doesn't support streaming. Please choose a streaming-capable model.",
		"invalid_arn":     "Invalid custom ARN '%s'. Please check the ARN format and ensure it's accessible.",
		"validation_failed": "Model validation failed. Please verify the model ID or custom ARN is correct.",
	},
	ErrorTypeDatabase: {
		"connection_failed":  "Failed to connect to the database. Chat history may not be available.",
		"migration_failed":   "Database migration failed. Some features may not work properly.",
		"query_failed":       "Database query failed. This operation could not be completed.",
		"chat_not_found":     "Chat with ID '%s' not found. Please check the chat ID or start a new conversation.",
		"save_failed":        "Failed to save chat message. Your conversation may not be preserved.",
	},
	ErrorTypeConfiguration: {
		"file_not_found":    "Configuration file not found. Using default settings.",
		"file_invalid":      "Configuration file is invalid. Please check the YAML syntax.",
		"permission_denied": "Permission denied accessing configuration file. Please check file permissions.",
		"value_invalid":     "Invalid configuration value for '%s'. Please check your config.yaml file.",
		"directory_failed":  "Failed to create configuration directory. Please check permissions.",
	},
	ErrorTypeValidation: {
		"chat_id_invalid":    "Invalid chat ID format. Please provide a valid UUID.",
		"model_id_empty":     "Model ID cannot be empty. Please specify a model ID.",
		"temperature_range":  "Temperature must be between 0.0 and 1.0.",
		"max_tokens_range":   "Max tokens must be greater than 0.",
		"top_p_range":        "Top P must be between 0.0 and 1.0.",
		"region_empty":       "AWS region cannot be empty. Please specify a region.",
	},
	ErrorTypeNetwork: {
		"connection_timeout": "Connection timeout. Please check your internet connection and try again.",
		"dns_resolution":     "DNS resolution failed. Please check your network settings.",
		"connection_refused": "Connection refused. The service may be temporarily unavailable.",
		"ssl_error":          "SSL/TLS error occurred. Please check your network security settings.",
	},
	ErrorTypeFileSystem: {
		"file_not_found":     "File '%s' not found. Please check the file path.",
		"permission_denied":  "Permission denied accessing file '%s'. Please check file permissions.",
		"disk_full":          "Disk space full. Please free up space and try again.",
		"file_too_large":     "File '%s' is too large to process.",
		"invalid_format":     "Invalid file format for '%s'. Please check the file type.",
	},
}

// GetUserMessage returns a user-friendly error message for the given error type and code
func GetUserMessage(errorType ErrorType, code string, args ...interface{}) string {
	if messages, exists := ErrorMessages[errorType]; exists {
		if template, exists := messages[code]; exists {
			// If template contains placeholders, format with provided arguments
			if len(args) > 0 {
				return fmt.Sprintf(template, args...)
			}
			return template
		}
	}
	
	// Fallback to a generic message if no specific template is found
	return fmt.Sprintf("An error occurred in %s: %s", errorType.String(), code)
}

// GetSuggestion returns a helpful suggestion for resolving the error
func GetSuggestion(errorType ErrorType, code string) string {
	suggestions := map[ErrorType]map[string]string{
		ErrorTypeAWS: {
			"credentials_not_found": "Run 'aws configure' to set up your credentials, or set the AWS_PROFILE environment variable.",
			"credentials_invalid":   "Verify your credentials with 'aws sts get-caller-identity'.",
			"region_invalid":        "Set a valid region with 'aws configure set region us-east-1'.",
			"permissions_denied":    "Contact your AWS administrator to grant Bedrock permissions.",
		},
		ErrorTypeModel: {
			"not_found":    "Run 'chat-cli models list' to see all available models.",
			"not_text":     "Choose a model that supports text generation from the models list.",
			"no_streaming": "Select a model that supports streaming responses.",
		},
		ErrorTypeDatabase: {
			"connection_failed": "Check if the database file is accessible and not corrupted.",
			"chat_not_found":    "Use 'chat-cli chat list' to see available conversations.",
		},
		ErrorTypeConfiguration: {
			"file_invalid":   "Validate your YAML syntax or delete the config file to regenerate defaults.",
			"value_invalid":  "Check the configuration documentation for valid values.",
		},
		ErrorTypeValidation: {
			"chat_id_invalid":   "Provide a valid UUID format for the chat ID.",
			"temperature_range": "Use a temperature value between 0.0 (focused) and 1.0 (creative).",
		},
	}
	
	if typeSuggestions, exists := suggestions[errorType]; exists {
		if suggestion, exists := typeSuggestions[code]; exists {
			return suggestion
		}
	}
	
	return "Please check the documentation or contact support for assistance."
}

// EnhanceErrorWithMessage enhances an AppError with user-friendly messages and suggestions
func EnhanceErrorWithMessage(err *AppError, args ...interface{}) *AppError {
	if err == nil {
		return nil
	}
	
	// Set user message if not already set
	if err.UserMessage == "" {
		err.UserMessage = GetUserMessage(err.Type, err.Code, args...)
	}
	
	// Add suggestion to metadata
	suggestion := GetSuggestion(err.Type, err.Code)
	if suggestion != "" {
		err.WithMetadata("suggestion", suggestion)
	}
	
	return err
}

// FormatErrorWithSuggestion formats an error message with its suggestion
func FormatErrorWithSuggestion(err *AppError) string {
	message := err.GetUserMessage()
	
	if suggestion, exists := err.Metadata["suggestion"]; exists {
		if suggestionStr, ok := suggestion.(string); ok {
			return fmt.Sprintf("%s\n\nSuggestion: %s", message, suggestionStr)
		}
	}
	
	return message
}

// IsRetryableError determines if an error is retryable based on its type and code
func IsRetryableError(err *AppError) bool {
	if err == nil {
		return false
	}
	
	retryableErrors := map[ErrorType][]string{
		ErrorTypeNetwork: {"connection_timeout", "dns_resolution", "connection_refused"},
		ErrorTypeAWS:     {"service_unavailable", "rate_limited", "connection_failed"},
		ErrorTypeDatabase: {"connection_failed", "query_failed"},
	}
	
	if codes, exists := retryableErrors[err.Type]; exists {
		for _, code := range codes {
			if err.Code == code {
				return true
			}
		}
	}
	
	return false
}

// GetRetryDelay returns the recommended retry delay for retryable errors
func GetRetryDelay(err *AppError, attempt int) int {
	if err == nil || !IsRetryableError(err) {
		return 0
	}
	
	// Exponential backoff: 1s, 2s, 4s, 8s, max 30s
	delay := 1 << attempt
	if delay > 30 {
		delay = 30
	}
	
	// Special cases for rate limiting
	if err.Type == ErrorTypeAWS && err.Code == "rate_limited" {
		return delay * 2 // Longer delays for rate limiting
	}
	
	return delay
}