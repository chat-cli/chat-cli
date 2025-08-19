package validation_test

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/chat-cli/chat-cli/validation"
)

// ExampleAWSConfigValidator demonstrates how to validate AWS configuration
func ExampleAWSConfigValidator() {
	// Create AWS configuration validator
	validator := validation.NewAWSConfigValidator("us-east-1")
	
	ctx := context.Background()
	
	// Validate AWS configuration
	if err := validator.Validate(ctx); err != nil {
		fmt.Printf("AWS validation failed: %v\n", err)
		return
	}
	
	fmt.Println("AWS configuration is valid")
	
	// Get the validated configuration
	config := validator.GetConfig()
	fmt.Printf("Validated region: %s\n", config.Region)
}

// ExampleModelValidator demonstrates how to validate Bedrock models
func ExampleModelValidator() {
	// Load AWS config (in real usage, this would be validated first)
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return
	}
	
	// Create model validator for a standard model ID
	validator := validation.NewModelValidator(
		"anthropic.claude-3-5-sonnet-20240620-v1:0",
		"",
		"us-east-1",
		cfg,
	)
	
	ctx := context.Background()
	
	// Validate model (this would make AWS API calls in real usage)
	if err := validator.Validate(ctx); err != nil {
		fmt.Printf("Model validation failed: %v\n", err)
		return
	}
	
	fmt.Println("Model is valid and available")
}

// ExampleModelValidator_customARN demonstrates custom ARN validation
func ExampleModelValidator_customARN() {
	cfg := aws.Config{Region: "us-east-1"}
	
	// Create model validator for a custom ARN
	validator := validation.NewModelValidator(
		"",
		"arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-3-5-sonnet-20240620-v1:0",
		"us-east-1",
		cfg,
	)
	
	ctx := context.Background()
	
	// Validate custom ARN (only validates format, no AWS API calls)
	if err := validator.Validate(ctx); err != nil {
		fmt.Printf("Custom ARN validation failed: %v\n", err)
		return
	}
	
	fmt.Println("Custom ARN format is valid")
}

// ExampleValidatorGroup demonstrates how to run multiple validators together
func ExampleValidatorGroup() {
	cfg := aws.Config{Region: "us-east-1"}
	
	// Create validator group
	group := validation.NewValidatorGroup(false) // Continue on errors
	
	// Add AWS configuration validator
	awsValidator := validation.NewAWSConfigValidator("us-east-1")
	group.Add(awsValidator)
	
	// Add model validator
	modelValidator := validation.NewModelValidator(
		"anthropic.claude-3-5-sonnet-20240620-v1:0",
		"",
		"us-east-1",
		cfg,
	)
	group.Add(modelValidator)
	
	ctx := context.Background()
	
	// Run all validators
	if err := group.Validate(ctx); err != nil {
		fmt.Printf("Validation group failed: %v\n", err)
		return
	}
	
	fmt.Println("All validations passed")
}

// ExampleValidationResult demonstrates working with validation results
func ExampleValidationResult() {
	// Create validation result
	result := validation.NewValidationResult()
	
	// Add context information
	result.AddContext("operation", "model_validation")
	result.AddContext("model_id", "test-model")
	
	fmt.Printf("Initial state - Valid: %t, Errors: %d\n", result.Valid, len(result.Errors))
	
	// Simulate adding an error
	// (In real usage, this would be done by validators)
	// result.AddError(someError)
	
	fmt.Printf("Context: %+v\n", result.Context)
	
	// Output:
	// Initial state - Valid: true, Errors: 0
	// Context: map[model_id:test-model operation:model_validation]
}