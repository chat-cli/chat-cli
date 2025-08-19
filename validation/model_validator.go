package validation

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/chat-cli/chat-cli/errors"
)

// ModelValidator validates Bedrock model availability and compatibility
type ModelValidator struct {
	ModelID   string
	CustomARN string
	Region    string
	Config    aws.Config
	client    *bedrock.Client
}

// NewModelValidator creates a new model validator
func NewModelValidator(modelID, customARN, region string, config aws.Config) *ModelValidator {
	return &ModelValidator{
		ModelID:   modelID,
		CustomARN: customARN,
		Region:    region,
		Config:    config,
		client:    bedrock.NewFromConfig(config),
	}
}

// Validate validates the model configuration
func (v *ModelValidator) Validate(ctx context.Context) error {
	// If custom ARN is provided, validate ARN format but skip model availability check
	if v.CustomARN != "" {
		return v.validateCustomARN()
	}

	// Validate model ID format
	if err := v.validateModelIDFormat(); err != nil {
		return err
	}

	// Validate model availability and compatibility
	return v.validateModelAvailability(ctx)
}

// validateCustomARN validates the format of a custom ARN
func (v *ModelValidator) validateCustomARN() error {
	if v.CustomARN == "" {
		return errors.NewValidationError(
			"custom_arn_empty",
			"Custom ARN is empty",
			"Custom ARN cannot be empty when specified.",
			nil,
		).WithOperation("ValidateCustomARN").WithComponent("model-validator")
	}

	// Validate ARN format: arn:aws:bedrock:region::foundation-model/model-id
	// or arn:aws:bedrock:region:account:foundation-model/model-id
	if !strings.HasPrefix(v.CustomARN, "arn:aws:bedrock:") {
		return errors.NewValidationError(
			"custom_arn_invalid_format",
			fmt.Sprintf("Invalid custom ARN format: %s", v.CustomARN),
			fmt.Sprintf("Invalid custom ARN format '%s'. Expected format: arn:aws:bedrock:region::foundation-model/model-id", v.CustomARN),
			nil,
		).WithOperation("ValidateCustomARN").WithComponent("model-validator").
			WithMetadata("provided_arn", v.CustomARN)
	}

	// Split ARN to validate components
	arnParts := strings.Split(v.CustomARN, ":")
	if len(arnParts) < 6 {
		// Check if it's just an invalid format first
		if len(arnParts) < 4 || arnParts[0] != "arn" || arnParts[1] != "aws" || arnParts[2] != "bedrock" {
			return errors.NewValidationError(
				"custom_arn_invalid_format",
				fmt.Sprintf("Invalid custom ARN format: %s", v.CustomARN),
				fmt.Sprintf("Invalid custom ARN format '%s'. Expected format: arn:aws:bedrock:region::foundation-model/model-id", v.CustomARN),
				nil,
			).WithOperation("ValidateCustomARN").WithComponent("model-validator").
				WithMetadata("provided_arn", v.CustomARN)
		}
		
		return errors.NewValidationError(
			"custom_arn_invalid_structure",
			fmt.Sprintf("Invalid custom ARN structure: %s", v.CustomARN),
			fmt.Sprintf("Invalid custom ARN structure '%s'. Expected format: arn:aws:bedrock:region::foundation-model/model-id", v.CustomARN),
			nil,
		).WithOperation("ValidateCustomARN").WithComponent("model-validator").
			WithMetadata("provided_arn", v.CustomARN)
	}

	// Validate that it's a foundation-model ARN
	if len(arnParts) >= 6 && !strings.HasPrefix(arnParts[5], "foundation-model/") {
		return errors.NewValidationError(
			"custom_arn_not_foundation_model",
			fmt.Sprintf("Custom ARN is not a foundation model: %s", v.CustomARN),
			fmt.Sprintf("Custom ARN '%s' is not a foundation model. Expected format: arn:aws:bedrock:region::foundation-model/model-id", v.CustomARN),
			nil,
		).WithOperation("ValidateCustomARN").WithComponent("model-validator").
			WithMetadata("provided_arn", v.CustomARN)
	}

	return nil
}

// validateModelIDFormat validates the format of a model ID
func (v *ModelValidator) validateModelIDFormat() error {
	if v.ModelID == "" {
		return errors.NewValidationError(
			"model_id_empty",
			"Model ID is empty",
			"Model ID cannot be empty. Please specify a valid Bedrock model ID.",
			nil,
		).WithOperation("ValidateModelID").WithComponent("model-validator")
	}

	// Basic format validation for common model ID patterns
	// Most Bedrock model IDs follow patterns like:
	// - anthropic.claude-3-5-sonnet-20240620-v1:0
	// - amazon.titan-text-express-v1
	// - meta.llama2-70b-chat-v1
	if len(v.ModelID) < 3 {
		return errors.NewValidationError(
			"model_id_too_short",
			fmt.Sprintf("Model ID too short: %s", v.ModelID),
			fmt.Sprintf("Model ID '%s' is too short. Please provide a valid Bedrock model ID.", v.ModelID),
			nil,
		).WithOperation("ValidateModelID").WithComponent("model-validator").
			WithMetadata("provided_model_id", v.ModelID)
	}

	return nil
}

// validateModelAvailability validates that the model exists and is compatible
func (v *ModelValidator) validateModelAvailability(ctx context.Context) error {
	// Get foundation model details
	model, err := v.client.GetFoundationModel(ctx, &bedrock.GetFoundationModelInput{
		ModelIdentifier: &v.ModelID,
	})
	if err != nil {
		// Parse AWS error to provide specific guidance
		errStr := err.Error()
		
		if strings.Contains(errStr, "ValidationException") && strings.Contains(errStr, "not found") {
			return errors.NewModelError(
				"model_not_found",
				fmt.Sprintf("Model not found: %s", v.ModelID),
				fmt.Sprintf("Model '%s' not found. Use 'chat-cli models list' to see available models.", v.ModelID),
				err,
			).WithOperation("ValidateModelAvailability").WithComponent("model-validator").
				WithMetadata("provided_model_id", v.ModelID)
		}
		
		if strings.Contains(errStr, "AccessDeniedException") {
			return errors.NewModelError(
				"model_access_denied",
				fmt.Sprintf("Access denied for model: %s", v.ModelID),
				fmt.Sprintf("Access denied for model '%s'. Please ensure the model is enabled in your AWS Bedrock console.", v.ModelID),
				err,
			).WithOperation("ValidateModelAvailability").WithComponent("model-validator").
				WithMetadata("provided_model_id", v.ModelID)
		}

		// Generic model error
		return errors.NewModelError(
			"model_validation_failed",
			fmt.Sprintf("Model validation failed: %v", err),
			fmt.Sprintf("Unable to validate model '%s'. Please check the model ID and your AWS configuration.", v.ModelID),
			err,
		).WithOperation("ValidateModelAvailability").WithComponent("model-validator").
			WithMetadata("provided_model_id", v.ModelID)
	}

	// Validate model capabilities
	if err := v.validateModelCapabilities(model.ModelDetails); err != nil {
		return err
	}

	return nil
}

// validateModelCapabilities validates that the model supports required capabilities
func (v *ModelValidator) validateModelCapabilities(modelDetails *types.FoundationModelDetails) error {
	// Check if this is a text model
	if !slices.Contains(modelDetails.OutputModalities, "TEXT") {
		return errors.NewModelError(
			"model_not_text",
			fmt.Sprintf("Model %s is not a text model", *modelDetails.ModelId),
			fmt.Sprintf("Model '%s' doesn't support text generation. Please choose a text-capable model.", *modelDetails.ModelId),
			nil,
		).WithOperation("ValidateModelCapabilities").WithComponent("model-validator").
			WithMetadata("model_id", *modelDetails.ModelId).
			WithMetadata("output_modalities", modelDetails.OutputModalities)
	}

	// Check if model supports streaming (required for chat functionality)
	if modelDetails.ResponseStreamingSupported == nil || !*modelDetails.ResponseStreamingSupported {
		return errors.NewModelError(
			"model_no_streaming",
			fmt.Sprintf("Model %s does not support streaming", *modelDetails.ModelId),
			fmt.Sprintf("Model '%s' doesn't support streaming. Please choose a streaming-capable model.", *modelDetails.ModelId),
			nil,
		).WithOperation("ValidateModelCapabilities").WithComponent("model-validator").
			WithMetadata("model_id", *modelDetails.ModelId).
			WithMetadata("streaming_supported", *modelDetails.ResponseStreamingSupported)
	}

	return nil
}

// GetSuggestedModels returns a list of commonly available models for the current region
func (v *ModelValidator) GetSuggestedModels() []string {
	// Common models that are typically available across regions
	return []string{
		"anthropic.claude-3-5-sonnet-20240620-v1:0",
		"anthropic.claude-3-haiku-20240307-v1:0",
		"anthropic.claude-3-opus-20240229-v1:0",
		"amazon.titan-text-express-v1",
		"amazon.titan-text-lite-v1",
	}
}

// ValidateModelList validates a list of model IDs
func (v *ModelValidator) ValidateModelList(ctx context.Context, modelIDs []string) error {
	validationGroup := NewValidatorGroup(false) // Don't stop on first error
	
	for _, modelID := range modelIDs {
		validator := NewModelValidator(modelID, "", v.Region, v.Config)
		validationGroup.Add(validator)
	}
	
	return validationGroup.Validate(ctx)
}