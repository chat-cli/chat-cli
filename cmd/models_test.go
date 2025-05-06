package cmd

import (
	"strings"
	"testing"
)

func TestModelsCommand(t *testing.T) {
	// Test if models command is properly registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "models" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'models' command to be registered to rootCmd")
	}
}

func TestModelsCommandRun(t *testing.T) {
	// Test that the models command outputs expected text
	output := captureOutput(func() {
		modelsCmd.Run(modelsCmd, []string{})
	})
	
	if !strings.Contains(output, "models called") {
		t.Errorf("Expected output to contain 'models called', got: %s", output)
	}
}