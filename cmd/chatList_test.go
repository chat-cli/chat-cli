package cmd

import (
	"strings"
	"testing"
)

func TestChatListCommand(t *testing.T) {
	// Test if chatList command is properly registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'list' command to be registered to rootCmd")
	}
}

// Testing without actual database connection since we would need to mock it for proper testing
func TestChatListCommandFlags(t *testing.T) {
	// Test if flags are properly defined
	if chatListCmd.Flags().Lookup("sort") == nil {
		t.Error("Expected 'sort' flag to be defined")
	}
	
	if chatListCmd.Flags().Lookup("direction") == nil {
		t.Error("Expected 'direction' flag to be defined")
	}
}

// Note: Full testing of the Run function would require mocking the database connection
// or setting up a test database. For now, we'll just test the command structure.
func TestChatListInitialization(t *testing.T) {
	// Verify command is properly set up
	if chatListCmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got %s", chatListCmd.Use)
	}
	
	if !strings.Contains(chatListCmd.Short, "List") {
		t.Errorf("Expected Short description to contain 'List', got %s", chatListCmd.Short)
	}
}