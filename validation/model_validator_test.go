package validation

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/chat-cli/chat-cli/errors"
)

func TestNewModelValidator(t *testing.T) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("test-model", "", "us-east-1", config)
	
	if validator == nil {
		t.Error("Expected validator to be created")
	}
	
	if validator.ModelID != "test-model" {
		t.Errorf("Expected model ID 'test-model', got %s", validator.ModelID)
	}
	
	if validator.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got %s", validator.Region)
	}
}

func TestModelValidator_validateCustomARN_Empty(t *testing.T) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("", "", "us-east-1", config)
	
	err := validator.validateCustomARN()
	
	if err == nil {
		t.Error("Expected error for empty custom ARN")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "custom_arn_empty" {
		t.Errorf("Expected error code 'custom_arn_empty', got %s", appErr.Code)
	}
}

func TestModelValidator_validateCustomARN_InvalidPrefix(t *testing.T) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("", "invalid:arn:format", "us-east-1", config)
	
	err := validator.validateCustomARN()
	
	if err == nil {
		t.Error("Expected error for invalid ARN prefix")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "custom_arn_invalid_format" {
		t.Errorf("Expected error code 'custom_arn_invalid_format', got %s", appErr.Code)
	}
}

func TestModelValidator_validateCustomARN_InvalidStructure(t *testing.T) {
	testCases := []struct {
		arn         string
		description string
		expectedCode string
	}{
		{"arn:aws:bedrock", "too few parts", "custom_arn_invalid_format"},
		{"arn:aws:bedrock:us-east-1", "missing account and resource", "custom_arn_invalid_structure"},
		{"arn:aws:bedrock:us-east-1:", "missing resource", "custom_arn_invalid_structure"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			config := aws.Config{Region: "us-east-1"}
			validator := NewModelValidator("", tc.arn, "us-east-1", config)
			
			err := validator.validateCustomARN()
			
			if err == nil {
				t.Errorf("Expected error for invalid ARN structure: %s", tc.arn)
			}
			
			appErr, ok := err.(*errors.AppError)
			if !ok {
				t.Error("Expected AppError type")
			}
			
			if appErr.Code != tc.expectedCode {
				t.Errorf("Expected error code '%s', got %s", tc.expectedCode, appErr.Code)
			}
		})
	}
}

func TestModelValidator_validateCustomARN_NotFoundationModel(t *testing.T) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("", "arn:aws:bedrock:us-east-1::custom-model/test", "us-east-1", config)
	
	err := validator.validateCustomARN()
	
	if err == nil {
		t.Error("Expected error for non-foundation-model ARN")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "custom_arn_not_foundation_model" {
		t.Errorf("Expected error code 'custom_arn_not_foundation_model', got %s", appErr.Code)
	}
}

func TestModelValidator_validateCustomARN_Valid(t *testing.T) {
	validARNs := []string{
		"arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-3-5-sonnet-20240620-v1:0",
		"arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-text-express-v1",
		"arn:aws:bedrock:eu-west-1:123456789012:foundation-model/meta.llama2-70b-chat-v1",
	}
	
	for _, arn := range validARNs {
		t.Run(arn, func(t *testing.T) {
			config := aws.Config{Region: "us-east-1"}
			validator := NewModelValidator("", arn, "us-east-1", config)
			
			err := validator.validateCustomARN()
			
			if err != nil {
				t.Errorf("Expected no error for valid ARN %s, got %v", arn, err)
			}
		})
	}
}

func TestModelValidator_validateModelIDFormat_Empty(t *testing.T) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("", "", "us-east-1", config)
	
	err := validator.validateModelIDFormat()
	
	if err == nil {
		t.Error("Expected error for empty model ID")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "model_id_empty" {
		t.Errorf("Expected error code 'model_id_empty', got %s", appErr.Code)
	}
}

func TestModelValidator_validateModelIDFormat_TooShort(t *testing.T) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("ab", "", "us-east-1", config)
	
	err := validator.validateModelIDFormat()
	
	if err == nil {
		t.Error("Expected error for too short model ID")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "model_id_too_short" {
		t.Errorf("Expected error code 'model_id_too_short', got %s", appErr.Code)
	}
}

func TestModelValidator_validateModelIDFormat_Valid(t *testing.T) {
	validModelIDs := []string{
		"anthropic.claude-3-5-sonnet-20240620-v1:0",
		"amazon.titan-text-express-v1",
		"meta.llama2-70b-chat-v1",
		"ai21.j2-ultra-v1",
		"cohere.command-text-v14",
	}
	
	for _, modelID := range validModelIDs {
		t.Run(modelID, func(t *testing.T) {
			config := aws.Config{Region: "us-east-1"}
			validator := NewModelValidator(modelID, "", "us-east-1", config)
			
			err := validator.validateModelIDFormat()
			
			if err != nil {
				t.Errorf("Expected no error for valid model ID %s, got %v", modelID, err)
			}
		})
	}
}

func TestModelValidator_Validate_CustomARNPath(t *testing.T) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("", "arn:aws:bedrock:us-east-1::foundation-model/test-model", "us-east-1", config)
	
	ctx := context.Background()
	err := validator.Validate(ctx)
	
	// Should only validate ARN format, not make AWS calls
	if err != nil {
		t.Errorf("Expected no error for valid custom ARN, got %v", err)
	}
}

func TestModelValidator_Validate_ModelIDPath_FormatValidation(t *testing.T) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("ab", "", "us-east-1", config) // Too short
	
	ctx := context.Background()
	err := validator.Validate(ctx)
	
	if err == nil {
		t.Error("Expected error for invalid model ID format")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "model_id_too_short" {
		t.Errorf("Expected error code 'model_id_too_short', got %s", appErr.Code)
	}
}

func TestModelValidator_GetSuggestedModels(t *testing.T) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("test", "", "us-east-1", config)
	
	models := validator.GetSuggestedModels()
	
	if len(models) == 0 {
		t.Error("Expected suggested models to be returned")
	}
	
	// Check that common models are included
	expectedModels := []string{
		"anthropic.claude-3-5-sonnet-20240620-v1:0",
		"amazon.titan-text-express-v1",
	}
	
	for _, expected := range expectedModels {
		found := false
		for _, model := range models {
			if model == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model %s to be in suggested models", expected)
		}
	}
}

func TestModelValidator_ValidateModelList(t *testing.T) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("", "", "us-east-1", config)
	
	// Test with only format validation (no AWS calls)
	modelIDs := []string{
		"ab", // Too short - should fail format validation
	}
	
	ctx := context.Background()
	err := validator.ValidateModelList(ctx, modelIDs)
	
	// Should fail because one model ID is invalid
	if err == nil {
		t.Error("Expected error when validating list with invalid model ID")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "model_id_too_short" {
		t.Errorf("Expected error code 'model_id_too_short', got %s", appErr.Code)
	}
}

// Benchmark tests
func BenchmarkModelValidator_validateCustomARN(b *testing.B) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("", "arn:aws:bedrock:us-east-1::foundation-model/test-model", "us-east-1", config)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.validateCustomARN()
	}
}

func BenchmarkModelValidator_validateModelIDFormat(b *testing.B) {
	config := aws.Config{Region: "us-east-1"}
	validator := NewModelValidator("anthropic.claude-3-5-sonnet-20240620-v1:0", "", "us-east-1", config)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.validateModelIDFormat()
	}
}