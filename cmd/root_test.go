package cmd

import (
	"testing"
)

func TestRootCmdHasUse(t *testing.T) {
	if rootCmd.Use == "" {
		t.Error("Expected rootCmd to have non-empty Use")
	}
}

func TestExecute(t *testing.T) {
	// This test ensures Execute doesn't panic
	// In a real test we would capture stdout/stderr to check output
	Execute()
}