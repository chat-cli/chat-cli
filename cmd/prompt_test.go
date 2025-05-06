package cmd

import (
	"testing"
)

func TestPromptCmd(t *testing.T) {
	if promptCmd.Use == "" {
		t.Error("Expected promptCmd to have non-empty Use")
	}
	
	// Test command flags
	flags := promptCmd.Flags()
	
	modelFlag := flags.Lookup("model")
	if modelFlag == nil {
		t.Error("Expected 'model' flag to exist")
	}
}