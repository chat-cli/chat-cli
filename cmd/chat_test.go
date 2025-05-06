package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestChatCommand(t *testing.T) {
	// Test if chat command is properly registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "chat" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'chat' command to be registered to rootCmd")
	}
}

// Helper function to capture stdout for testing
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestChatCommandFlags(t *testing.T) {
	// Test if flags are properly defined
	if chatCmd.Flags().Lookup("model") == nil {
		t.Error("Expected 'model' flag to be defined")
	}

	if chatCmd.Flags().Lookup("temperature") == nil {
		t.Error("Expected 'temperature' flag to be defined")
	}
}