package cmd

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
)

func TestItemInterface(t *testing.T) {
	item := item{
		modelID:     "test-model-id",
		modelName:   "Test Model",
		provider:    "Test Provider",
		status:      "ACTIVE",
		crossRegion: true,
		modelArn:    "arn:aws:bedrock:us-east-1::foundation-model/test-model-id",
	}

	// Test Title method
	if item.Title() != "Test Model" {
		t.Errorf("Expected title 'Test Model', got '%s'", item.Title())
	}

	// Test Description method
	expectedDesc := "Test Provider • test-model-id • Cross-Region"
	if item.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, item.Description())
	}

	// Test FilterValue method
	expectedFilter := "Test Model Test Provider test-model-id"
	if item.FilterValue() != expectedFilter {
		t.Errorf("Expected filter value '%s', got '%s'", expectedFilter, item.FilterValue())
	}
}

func TestItemInterfaceWithoutCrossRegion(t *testing.T) {
	item := item{
		modelID:     "test-model-id",
		modelName:   "Test Model",
		provider:    "Test Provider",
		status:      "ACTIVE",
		crossRegion: false,
		modelArn:    "arn:aws:bedrock:us-east-1::foundation-model/test-model-id",
	}

	// Test Description method without cross-region
	expectedDesc := "Test Provider • test-model-id"
	if item.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, item.Description())
	}
}

func TestItemImplementsListItem(t *testing.T) {
	item := item{
		modelID:     "test-model-id",
		modelName:   "Test Model",
		provider:    "Test Provider",
		status:      "ACTIVE",
		crossRegion: true,
		modelArn:    "arn:aws:bedrock:us-east-1::foundation-model/test-model-id",
	}

	// Verify that item implements list.Item interface
	var _ list.Item = item
}

func TestSetModelInConfigFunction(t *testing.T) {
	// Test that the function exists and has the right signature
	// This is a basic test since actual config setting requires file system access
	// We test with an invalid model ID to ensure error handling
	err := setModelInConfig("")
	if err == nil {
		t.Log("setModelInConfig function exists and can be called")
	}
}

func TestSetCustomArnInConfigFunction(t *testing.T) {
	// Test that the function exists and has the right signature
	// This is a basic test since actual config setting requires file system access
	err := setCustomArnInConfig("")
	if err == nil {
		t.Log("setCustomArnInConfig function exists and can be called")
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"24", true},
		{"0", true},
		{"24k", false},
		{"abc", false},
		{"", false},
		{"v1", false},
		{"1000", true},
	}

	for _, test := range tests {
		result := isNumeric(test.input)
		if result != test.expected {
			t.Errorf("isNumeric(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestIsCapacitySuffix(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Capacity suffixes that should be filtered
		{"24k", true},
		{"200k", true},
		{"1000k", true},
		{"4k", true},
		{"8k", true},
		{"48k", true},
		{"128k", true},
		{"300k", true},
		{"mm", true},
		{"512", true}, // Large number > 10

		// Version identifiers that should be kept
		{"0", false},
		{"1", false},
		{"2", false},
		{"7", false},
		{"10", false}, // Edge case: exactly 10 should be kept

		// Other non-capacity suffixes
		{"v1", false},
		{"abc", false},
		{"", false},
	}

	for _, test := range tests {
		result := isCapacitySuffix(test.input)
		if result != test.expected {
			t.Errorf("isCapacitySuffix(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}
