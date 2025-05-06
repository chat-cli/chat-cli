package cmd

import (
	"testing"
)

func TestPromptCommand(t *testing.T) {
	// Test if prompt command is properly registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "prompt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'prompt' command to be registered to rootCmd")
	}
}

func TestPromptCommandFlags(t *testing.T) {
	// Test if flags are properly defined
	if promptCmd.Flags().Lookup("model") == nil {
		t.Error("Expected 'model' flag to be defined")
	}

	if promptCmd.Flags().Lookup("temperature") == nil {
		t.Error("Expected 'temperature' flag to be defined")
	}
}

func TestPromptCommandArgs(t *testing.T) {
	// Test Args validation function (MinimumNArgs(1))
	if promptCmd.Args(promptCmd, []string{}) == nil {
		t.Error("Expected error when no arguments are provided")
	}
	
	if promptCmd.Args(promptCmd, []string{"test prompt"}) != nil {
		t.Error("Expected no error when at least one argument is provided")
	}
}