package cmd

import (
	"testing"
)

func TestChatCmdHasUse(t *testing.T) {
	if chatCmd.Use == "" {
		t.Error("Expected chatCmd to have non-empty Use")
	}
}

func TestChatCmdFlags(t *testing.T) {
	// Test that required flags are set correctly
	flags := chatCmd.Flags()
	
	modelFlag := flags.Lookup("model")
	if modelFlag == nil {
		t.Error("Expected 'model' flag to exist")
	}
	
	// Add more flag tests as needed
}