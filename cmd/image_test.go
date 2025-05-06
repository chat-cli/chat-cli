package cmd

import (
	"testing"
)

func TestImageCommand(t *testing.T) {
	// Test if image command is properly registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "image" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'image' command to be registered to rootCmd")
	}
}

func TestImageCommandFlags(t *testing.T) {
	// Test if flags are properly defined
	if imageCmd.Flags().Lookup("model") == nil {
		t.Error("Expected 'model' flag to be defined")
	}

	if imageCmd.Flags().Lookup("width") == nil {
		t.Error("Expected 'width' flag to be defined")
	}

	if imageCmd.Flags().Lookup("height") == nil {
		t.Error("Expected 'height' flag to be defined")
	}

	if imageCmd.Flags().Lookup("cfg-scale") == nil {
		t.Error("Expected 'cfg-scale' flag to be defined")
	}

	if imageCmd.Flags().Lookup("steps") == nil {
		t.Error("Expected 'steps' flag to be defined")
	}

	if imageCmd.Flags().Lookup("seed") == nil {
		t.Error("Expected 'seed' flag to be defined")
	}

	if imageCmd.Flags().Lookup("print") == nil {
		t.Error("Expected 'print' flag to be defined")
	}
}

func TestImageCommandArgs(t *testing.T) {
	// Test Args validation function (MinimumNArgs(1))
	if imageCmd.Args(imageCmd, []string{}) == nil {
		t.Error("Expected error when no arguments are provided")
	}
	
	if imageCmd.Args(imageCmd, []string{"test prompt"}) != nil {
		t.Error("Expected no error when at least one argument is provided")
	}
}