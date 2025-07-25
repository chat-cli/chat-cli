//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Integration tests require the CLI to be built first
// Run with: go test -tags=integration

func TestCLIVersion(t *testing.T) {
	// Build the CLI if it doesn't exist
	if _, err := os.Stat("./bin/chat-cli"); os.IsNotExist(err) {
		cmd := exec.Command("make", "cli")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build CLI: %v", err)
		}
	}

	// Test version command
	cmd := exec.Command("./bin/chat-cli", "version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "chat-cli version") {
		t.Errorf("Expected version output to contain 'chat-cli version', got: %s", outputStr)
	}
}

func TestCLIHelp(t *testing.T) {
	// Build the CLI if it doesn't exist
	if _, err := os.Stat("./bin/chat-cli"); os.IsNotExist(err) {
		cmd := exec.Command("make", "cli")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build CLI: %v", err)
		}
	}

	// Test help command
	cmd := exec.Command("./bin/chat-cli", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	outputStr := string(output)
	expectedStrings := []string{
		"Chat with LLMs from Amazon Bedrock!",
		"Available Commands:",
		"chat",
		"prompt",
		"config",
		"image",
		"models",
		"version",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected help output to contain '%s', got: %s", expected, outputStr)
		}
	}
}

func TestCLIConfigHelp(t *testing.T) {
	// Build the CLI if it doesn't exist
	if _, err := os.Stat("./bin/chat-cli"); os.IsNotExist(err) {
		cmd := exec.Command("make", "cli")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build CLI: %v", err)
		}
	}

	// Test config help
	cmd := exec.Command("./bin/chat-cli", "config", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Config help command failed: %v", err)
	}

	outputStr := string(output)
	expectedSubcommands := []string{
		"set",
		"get",
		"unset",
	}

	for _, expected := range expectedSubcommands {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected config help to contain '%s', got: %s", expected, outputStr)
		}
	}
}

func TestCLIPromptNoArgs(t *testing.T) {
	// Build the CLI if it doesn't exist
	if _, err := os.Stat("./bin/chat-cli"); os.IsNotExist(err) {
		cmd := exec.Command("make", "cli")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build CLI: %v", err)
		}
	}

	// Test prompt command without arguments (should fail)
	cmd := exec.Command("./bin/chat-cli", "prompt")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected prompt command to fail without arguments")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "requires at least 1 arg") {
		t.Errorf("Expected error about missing arguments, got: %s", outputStr)
	}
}

func TestCLIImageNoArgs(t *testing.T) {
	// Build the CLI if it doesn't exist
	if _, err := os.Stat("./bin/chat-cli"); os.IsNotExist(err) {
		cmd := exec.Command("make", "cli")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build CLI: %v", err)
		}
	}

	// Test image command without arguments (should fail)
	cmd := exec.Command("./bin/chat-cli", "image")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected image command to fail without arguments")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "requires at least 1 arg") {
		t.Errorf("Expected error about missing arguments, got: %s", outputStr)
	}
}

func TestCLIFlagsExist(t *testing.T) {
	// Build the CLI if it doesn't exist
	if _, err := os.Stat("./bin/chat-cli"); os.IsNotExist(err) {
		cmd := exec.Command("make", "cli")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build CLI: %v", err)
		}
	}

	// Test that expected flags exist
	cmd := exec.Command("./bin/chat-cli", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	outputStr := string(output)
	expectedFlags := []string{
		"--region",
		"--model-id",
		"--custom-arn",
	}

	for _, flag := range expectedFlags {
		if !strings.Contains(outputStr, flag) {
			t.Errorf("Expected help output to contain flag '%s', got: %s", flag, outputStr)
		}
	}
}

func TestCLIModelsSubcommands(t *testing.T) {
	// Build the CLI if it doesn't exist
	if _, err := os.Stat("./bin/chat-cli"); os.IsNotExist(err) {
		cmd := exec.Command("make", "cli")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build CLI: %v", err)
		}
	}

	// Test models help
	cmd := exec.Command("./bin/chat-cli", "models", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Models help command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "list") {
		t.Errorf("Expected models help to contain 'list' subcommand, got: %s", outputStr)
	}
}
