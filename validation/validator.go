package validation

import (
	"context"

	"github.com/chat-cli/chat-cli/errors"
)

// Validator interface for early validation of configurations and inputs
type Validator interface {
	Validate(ctx context.Context) error
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid   bool
	Errors  []*errors.AppError
	Context map[string]interface{}
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:   true,
		Errors:  make([]*errors.AppError, 0),
		Context: make(map[string]interface{}),
	}
}

// AddError adds an error to the validation result and marks it as invalid
func (vr *ValidationResult) AddError(err *errors.AppError) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, err)
}

// AddContext adds context information to the validation result
func (vr *ValidationResult) AddContext(key string, value interface{}) {
	vr.Context[key] = value
}

// GetFirstError returns the first error in the validation result, or nil if valid
func (vr *ValidationResult) GetFirstError() *errors.AppError {
	if len(vr.Errors) > 0 {
		return vr.Errors[0]
	}
	return nil
}

// CombineResults combines multiple validation results into one
func CombineResults(results ...*ValidationResult) *ValidationResult {
	combined := NewValidationResult()
	
	for _, result := range results {
		if !result.Valid {
			combined.Valid = false
			combined.Errors = append(combined.Errors, result.Errors...)
		}
		
		// Merge context
		for key, value := range result.Context {
			combined.Context[key] = value
		}
	}
	
	return combined
}

// ValidatorGroup allows running multiple validators together
type ValidatorGroup struct {
	validators []Validator
	stopOnFirstError bool
}

// NewValidatorGroup creates a new validator group
func NewValidatorGroup(stopOnFirstError bool) *ValidatorGroup {
	return &ValidatorGroup{
		validators: make([]Validator, 0),
		stopOnFirstError: stopOnFirstError,
	}
}

// Add adds a validator to the group
func (vg *ValidatorGroup) Add(validator Validator) {
	vg.validators = append(vg.validators, validator)
}

// Validate runs all validators in the group
func (vg *ValidatorGroup) Validate(ctx context.Context) error {
	result := NewValidationResult()
	
	for _, validator := range vg.validators {
		if err := validator.Validate(ctx); err != nil {
			if appErr, ok := err.(*errors.AppError); ok {
				result.AddError(appErr)
			} else {
				// Wrap non-AppError as validation error
				validationErr := errors.NewValidationError(
					"validation_failed",
					"Validation failed",
					"A validation check failed",
					err,
				).WithOperation("ValidatorGroup.Validate")
				result.AddError(validationErr)
			}
			
			if vg.stopOnFirstError {
				break
			}
		}
	}
	
	if !result.Valid {
		return result.GetFirstError()
	}
	
	return nil
}