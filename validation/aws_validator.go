package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/chat-cli/chat-cli/errors"
)

// AWSConfigValidator validates AWS configuration including credentials and region
type AWSConfigValidator struct {
	Region string
	Config aws.Config
}

// NewAWSConfigValidator creates a new AWS configuration validator
func NewAWSConfigValidator(region string) *AWSConfigValidator {
	return &AWSConfigValidator{
		Region: region,
	}
}

// Validate validates AWS configuration
func (v *AWSConfigValidator) Validate(ctx context.Context) error {
	// Validate region format first
	if err := v.validateRegion(); err != nil {
		return err
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(v.Region))
	if err != nil {
		return errors.NewAWSError(
			"config_load_failed",
			fmt.Sprintf("Failed to load AWS configuration: %v", err),
			"Unable to load AWS configuration. Please ensure AWS CLI is configured or environment variables are set.",
			err,
		).WithOperation("LoadAWSConfig").WithComponent("aws-validator")
	}

	v.Config = cfg

	// Validate credentials by making a test call to STS
	if err := v.validateCredentials(ctx); err != nil {
		return err
	}

	return nil
}

// validateRegion validates the AWS region format and availability
func (v *AWSConfigValidator) validateRegion() error {
	if v.Region == "" {
		return errors.NewValidationError(
			"region_empty",
			"AWS region is empty",
			"AWS region must be specified. Please set a valid region like 'us-east-1'.",
			nil,
		).WithOperation("ValidateRegion").WithComponent("aws-validator")
	}

	// Validate region format (e.g., us-east-1, eu-west-1, ap-southeast-2)
	regionPattern := regexp.MustCompile(`(?i)^[a-z]{2,3}-[a-z]+-\d+$`)
	if !regionPattern.MatchString(v.Region) {
		return errors.NewValidationError(
			"region_invalid_format",
			fmt.Sprintf("Invalid AWS region format: %s", v.Region),
			fmt.Sprintf("Invalid AWS region '%s'. Please use a valid region format like 'us-east-1' or 'eu-west-1'.", v.Region),
			nil,
		).WithOperation("ValidateRegion").WithComponent("aws-validator").
			WithMetadata("provided_region", v.Region)
	}

	// Check if region is a known AWS region
	knownRegions := []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1", "eu-north-1",
		"ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2", "ap-south-1",
		"ca-central-1", "sa-east-1", "af-south-1", "me-south-1",
		"ap-east-1", "ap-southeast-3", "eu-south-1", "me-central-1",
	}

	regionFound := false
	for _, region := range knownRegions {
		if strings.EqualFold(v.Region, region) {
			regionFound = true
			break
		}
	}

	if !regionFound {
		return errors.NewValidationError(
			"region_unknown",
			fmt.Sprintf("Unknown AWS region: %s", v.Region),
			fmt.Sprintf("Unknown AWS region '%s'. Please use a valid AWS region like 'us-east-1'.", v.Region),
			nil,
		).WithOperation("ValidateRegion").WithComponent("aws-validator").
			WithMetadata("provided_region", v.Region).
			WithMetadata("suggested_regions", []string{"us-east-1", "us-west-2", "eu-west-1"})
	}

	return nil
}

// validateCredentials validates AWS credentials by making a test STS call
func (v *AWSConfigValidator) validateCredentials(ctx context.Context) error {
	stsClient := sts.NewFromConfig(v.Config)

	// Make a GetCallerIdentity call to validate credentials
	_, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		// Parse AWS error to provide specific guidance
		errStr := err.Error()
		
		if strings.Contains(errStr, "NoCredentialsError") || strings.Contains(errStr, "no credentials") {
			return errors.NewAWSError(
				"credentials_not_found",
				"AWS credentials not found",
				"AWS credentials not found. Please run 'aws configure' or set AWS_PROFILE environment variable.",
				err,
			).WithOperation("ValidateCredentials").WithComponent("aws-validator")
		}
		
		if strings.Contains(errStr, "InvalidUserID.NotFound") || strings.Contains(errStr, "SignatureDoesNotMatch") {
			return errors.NewAWSError(
				"credentials_invalid",
				"AWS credentials are invalid",
				"AWS credentials are invalid. Please check your access key and secret key with 'aws configure'.",
				err,
			).WithOperation("ValidateCredentials").WithComponent("aws-validator")
		}
		
		if strings.Contains(errStr, "TokenRefreshRequired") || strings.Contains(errStr, "ExpiredToken") {
			return errors.NewAWSError(
				"credentials_expired",
				"AWS credentials have expired",
				"AWS credentials have expired. Please refresh your credentials or re-run 'aws configure'.",
				err,
			).WithOperation("ValidateCredentials").WithComponent("aws-validator")
		}
		
		if strings.Contains(errStr, "AccessDenied") {
			return errors.NewAWSError(
				"permissions_denied",
				"Insufficient AWS permissions",
				"Insufficient AWS permissions. Please ensure your AWS user has the necessary permissions.",
				err,
			).WithOperation("ValidateCredentials").WithComponent("aws-validator")
		}

		// Generic AWS error
		return errors.NewAWSError(
			"credentials_validation_failed",
			fmt.Sprintf("AWS credentials validation failed: %v", err),
			"Unable to validate AWS credentials. Please check your AWS configuration.",
			err,
		).WithOperation("ValidateCredentials").WithComponent("aws-validator")
	}

	return nil
}

// GetConfig returns the loaded AWS configuration after successful validation
func (v *AWSConfigValidator) GetConfig() aws.Config {
	return v.Config
}