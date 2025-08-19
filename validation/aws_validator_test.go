package validation

import (
	"context"
	"testing"

	"github.com/chat-cli/chat-cli/errors"
)

func TestNewAWSConfigValidator(t *testing.T) {
	validator := NewAWSConfigValidator("us-east-1")
	
	if validator == nil {
		t.Error("Expected validator to be created")
	}
	
	if validator.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got %s", validator.Region)
	}
}

func TestAWSConfigValidator_validateRegion_Empty(t *testing.T) {
	validator := NewAWSConfigValidator("")
	
	err := validator.validateRegion()
	
	if err == nil {
		t.Error("Expected error for empty region")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "region_empty" {
		t.Errorf("Expected error code 'region_empty', got %s", appErr.Code)
	}
}

func TestAWSConfigValidator_validateRegion_InvalidFormat(t *testing.T) {
	testCases := []struct {
		region      string
		description string
	}{
		{"invalid", "simple invalid format"},
		{"us-east", "missing number"},
		{"123-east-1", "invalid prefix"},
		{"us_east_1", "underscores instead of hyphens"},
		{"us-east-1-extra", "too many parts"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			validator := NewAWSConfigValidator(tc.region)
			
			err := validator.validateRegion()
			
			if err == nil {
				t.Errorf("Expected error for invalid region format: %s", tc.region)
			}
			
			appErr, ok := err.(*errors.AppError)
			if !ok {
				t.Error("Expected AppError type")
			}
			
			if appErr.Code != "region_invalid_format" {
				t.Errorf("Expected error code 'region_invalid_format', got %s", appErr.Code)
			}
		})
	}
}

func TestAWSConfigValidator_validateRegion_UnknownRegion(t *testing.T) {
	validator := NewAWSConfigValidator("xx-unknown-1")
	
	err := validator.validateRegion()
	
	if err == nil {
		t.Error("Expected error for unknown region")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "region_unknown" {
		t.Errorf("Expected error code 'region_unknown', got %s", appErr.Code)
	}
}

func TestAWSConfigValidator_validateRegion_ValidRegions(t *testing.T) {
	validRegions := []string{
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
		"eu-west-1",
		"eu-west-2",
		"eu-central-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ca-central-1",
	}
	
	for _, region := range validRegions {
		t.Run(region, func(t *testing.T) {
			validator := NewAWSConfigValidator(region)
			
			err := validator.validateRegion()
			
			if err != nil {
				t.Errorf("Expected no error for valid region %s, got %v", region, err)
			}
		})
	}
}

func TestAWSConfigValidator_validateRegion_CaseInsensitive(t *testing.T) {
	testCases := []string{
		"US-EAST-1",
		"Us-East-1",
		"us-EAST-1",
	}
	
	for _, region := range testCases {
		t.Run(region, func(t *testing.T) {
			validator := NewAWSConfigValidator(region)
			
			err := validator.validateRegion()
			
			if err != nil {
				t.Errorf("Expected no error for case variation %s, got %v", region, err)
			}
		})
	}
}

// Note: The following tests would require AWS credentials and network access
// In a real environment, you might want to use AWS SDK mocks or integration test tags

func TestAWSConfigValidator_Validate_RegionValidationFirst(t *testing.T) {
	// Test that region validation happens first, before AWS config loading
	validator := NewAWSConfigValidator("invalid-region")
	
	ctx := context.Background()
	err := validator.Validate(ctx)
	
	if err == nil {
		t.Error("Expected error for invalid region")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	// Should fail on region validation, not AWS config loading
	if appErr.Code != "region_invalid_format" {
		t.Errorf("Expected region validation error, got %s", appErr.Code)
	}
}

func TestAWSConfigValidator_GetConfig(t *testing.T) {
	validator := NewAWSConfigValidator("us-east-1")
	
	// Initially, config should be zero value
	config := validator.GetConfig()
	if config.Region != "" {
		t.Error("Expected empty config before validation")
	}
	
	// Note: After successful validation, GetConfig() would return the loaded config
	// This would require valid AWS credentials to test properly
}

// Benchmark tests for performance validation
func BenchmarkAWSConfigValidator_validateRegion(b *testing.B) {
	validator := NewAWSConfigValidator("us-east-1")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.validateRegion()
	}
}

func BenchmarkAWSConfigValidator_validateRegion_Invalid(b *testing.B) {
	validator := NewAWSConfigValidator("invalid-region")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.validateRegion()
	}
}