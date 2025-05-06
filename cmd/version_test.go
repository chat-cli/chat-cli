package cmd

import (
	"runtime"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Test if version command is properly registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'version' command to be registered to rootCmd")
	}
}

func TestVersionCommandRun(t *testing.T) {
	// Test that the version command outputs expected format
	output := captureOutput(func() {
		versionCmd.Run(versionCmd, []string{})
	})
	
	// Test contains version format
	if !strings.Contains(output, "chat-cli v") {
		t.Errorf("Expected output to contain 'chat-cli v', got: %s", output)
	}
	
	// Test contains OS and architecture
	osArch := runtime.GOOS + "/" + runtime.GOARCH
	if !strings.Contains(output, osArch) {
		t.Errorf("Expected output to contain OS/Arch '%s', got: %s", osArch, output)
	}
}