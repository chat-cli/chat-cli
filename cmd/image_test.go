package cmd

import (
	"testing"
)

func TestImageCmd(t *testing.T) {
	if imageCmd.Use == "" {
		t.Error("Expected imageCmd to have non-empty Use")
	}
	
	// Test command flags
	flags := imageCmd.Flags()
	
	modelFlag := flags.Lookup("model")
	if modelFlag == nil {
		t.Error("Expected 'model' flag to exist")
	}
}