package validation

import (
	"context"
	"testing"

	"github.com/chat-cli/chat-cli/errors"
)

// mockValidator is a test validator that can be configured to pass or fail
type mockValidator struct {
	shouldFail bool
	errorCode  string
	errorMsg   string
}

func (m *mockValidator) Validate(ctx context.Context) error {
	if m.shouldFail {
		return errors.NewValidationError(
			m.errorCode,
			m.errorMsg,
			"Mock validation failed",
			nil,
		)
	}
	return nil
}

func TestNewValidationResult(t *testing.T) {
	result := NewValidationResult()
	
	if !result.Valid {
		t.Error("Expected new validation result to be valid")
	}
	
	if len(result.Errors) != 0 {
		t.Error("Expected new validation result to have no errors")
	}
	
	if result.Context == nil {
		t.Error("Expected new validation result to have initialized context")
	}
}

func TestValidationResult_AddError(t *testing.T) {
	result := NewValidationResult()
	
	err := errors.NewValidationError(
		"test_error",
		"Test error",
		"Test error message",
		nil,
	)
	
	result.AddError(err)
	
	if result.Valid {
		t.Error("Expected validation result to be invalid after adding error")
	}
	
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
	
	if result.Errors[0] != err {
		t.Error("Expected added error to match")
	}
}

func TestValidationResult_AddContext(t *testing.T) {
	result := NewValidationResult()
	
	result.AddContext("test_key", "test_value")
	
	if result.Context["test_key"] != "test_value" {
		t.Error("Expected context to be added correctly")
	}
}

func TestValidationResult_GetFirstError(t *testing.T) {
	result := NewValidationResult()
	
	// Test with no errors
	if result.GetFirstError() != nil {
		t.Error("Expected nil when no errors present")
	}
	
	// Test with errors
	err1 := errors.NewValidationError("error1", "Error 1", "Error 1", nil)
	err2 := errors.NewValidationError("error2", "Error 2", "Error 2", nil)
	
	result.AddError(err1)
	result.AddError(err2)
	
	firstError := result.GetFirstError()
	if firstError != err1 {
		t.Error("Expected first error to be returned")
	}
}

func TestCombineResults(t *testing.T) {
	// Create valid result
	validResult := NewValidationResult()
	validResult.AddContext("valid_key", "valid_value")
	
	// Create invalid result
	invalidResult := NewValidationResult()
	err := errors.NewValidationError("test_error", "Test error", "Test error", nil)
	invalidResult.AddError(err)
	invalidResult.AddContext("invalid_key", "invalid_value")
	
	// Combine results
	combined := CombineResults(validResult, invalidResult)
	
	// Should be invalid due to one invalid result
	if combined.Valid {
		t.Error("Expected combined result to be invalid")
	}
	
	// Should contain error from invalid result
	if len(combined.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(combined.Errors))
	}
	
	// Should contain context from both results
	if combined.Context["valid_key"] != "valid_value" {
		t.Error("Expected valid context to be preserved")
	}
	
	if combined.Context["invalid_key"] != "invalid_value" {
		t.Error("Expected invalid context to be preserved")
	}
}

func TestNewValidatorGroup(t *testing.T) {
	group := NewValidatorGroup(true)
	
	if group == nil {
		t.Error("Expected validator group to be created")
	}
	
	if !group.stopOnFirstError {
		t.Error("Expected stopOnFirstError to be true")
	}
	
	if len(group.validators) != 0 {
		t.Error("Expected empty validators list")
	}
}

func TestValidatorGroup_Add(t *testing.T) {
	group := NewValidatorGroup(false)
	validator := &mockValidator{}
	
	group.Add(validator)
	
	if len(group.validators) != 1 {
		t.Errorf("Expected 1 validator, got %d", len(group.validators))
	}
}

func TestValidatorGroup_Validate_AllPass(t *testing.T) {
	group := NewValidatorGroup(false)
	
	// Add passing validators
	group.Add(&mockValidator{shouldFail: false})
	group.Add(&mockValidator{shouldFail: false})
	
	ctx := context.Background()
	err := group.Validate(ctx)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidatorGroup_Validate_OneFails(t *testing.T) {
	group := NewValidatorGroup(false)
	
	// Add one passing and one failing validator
	group.Add(&mockValidator{shouldFail: false})
	group.Add(&mockValidator{
		shouldFail: true,
		errorCode:  "test_error",
		errorMsg:   "Test error",
	})
	
	ctx := context.Background()
	err := group.Validate(ctx)
	
	if err == nil {
		t.Error("Expected error when validator fails")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "test_error" {
		t.Errorf("Expected error code 'test_error', got %s", appErr.Code)
	}
}

func TestValidatorGroup_Validate_StopOnFirstError(t *testing.T) {
	group := NewValidatorGroup(true) // Stop on first error
	
	// Add two failing validators
	group.Add(&mockValidator{
		shouldFail: true,
		errorCode:  "first_error",
		errorMsg:   "First error",
	})
	group.Add(&mockValidator{
		shouldFail: true,
		errorCode:  "second_error",
		errorMsg:   "Second error",
	})
	
	ctx := context.Background()
	err := group.Validate(ctx)
	
	if err == nil {
		t.Error("Expected error when validator fails")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	// Should only get the first error due to stopOnFirstError
	if appErr.Code != "first_error" {
		t.Errorf("Expected first error code, got %s", appErr.Code)
	}
}

func TestValidatorGroup_Validate_ContinueOnError(t *testing.T) {
	group := NewValidatorGroup(false) // Continue on error
	
	// Add two failing validators
	group.Add(&mockValidator{
		shouldFail: true,
		errorCode:  "first_error",
		errorMsg:   "First error",
	})
	group.Add(&mockValidator{
		shouldFail: true,
		errorCode:  "second_error",
		errorMsg:   "Second error",
	})
	
	ctx := context.Background()
	err := group.Validate(ctx)
	
	if err == nil {
		t.Error("Expected error when validators fail")
	}
	
	// Should get the first error (but both validators should have run)
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Error("Expected AppError type")
	}
	
	if appErr.Code != "first_error" {
		t.Errorf("Expected first error code, got %s", appErr.Code)
	}
}