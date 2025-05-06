package cmd

import (
	"testing"
	"os"
)

func TestExecute(t *testing.T) {
	// Save original args and restore them at the end of the test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test basic execution (success case)
	os.Args = []string{"chat-cli"}
	Execute()
}

func TestInit(t *testing.T) {
	// Check if the region flag exists and has the correct default value
	if rootCmd.PersistentFlags().Lookup("region") == nil {
		t.Error("Expected 'region' flag to be defined")
	}

	region, _ := rootCmd.PersistentFlags().GetString("region")
	if region != "us-east-1" {
		t.Errorf("Expected default region to be 'us-east-1', got '%s'", region)
	}
}