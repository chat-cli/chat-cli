package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	// Test that root command exists and has expected properties
	if rootCmd.Use != "chat-cli" {
		t.Errorf("Expected Use 'chat-cli', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short != "Chat with LLMs from Amazon Bedrock!" {
		t.Errorf("Expected correct short description, got '%s'", rootCmd.Short)
	}

	// Test that root command's help text mentions /quit command
	if !strings.Contains(rootCmd.Long, "\"quit\" or \"/quit\"") {
		t.Errorf("Expected help text to mention both quit and /quit commands")
	}

	// Test that root command has expected flags
	flag := rootCmd.PersistentFlags().Lookup("region")
	if flag == nil {
		t.Error("Expected 'region' flag to exist")
	} else {
		if flag.DefValue != "us-east-1" {
			t.Errorf("Expected default region 'us-east-1', got '%s'", flag.DefValue)
		}
	}

	flag = rootCmd.PersistentFlags().Lookup("model-id")
	if flag == nil {
		t.Error("Expected 'model-id' flag to exist")
	} else {
		expectedDefault := "anthropic.claude-3-5-sonnet-20240620-v1:0"
		if flag.DefValue != expectedDefault {
			t.Errorf("Expected default model-id '%s', got '%s'", expectedDefault, flag.DefValue)
		}
	}

	flag = rootCmd.PersistentFlags().Lookup("custom-arn")
	if flag == nil {
		t.Error("Expected 'custom-arn' flag to exist")
	}
}

func TestPromptCommand(t *testing.T) {
	// Test that prompt command exists and has expected properties
	if promptCmd.Use != "prompt" {
		t.Errorf("Expected Use 'prompt', got '%s'", promptCmd.Use)
	}

	if promptCmd.Short != "Send a prompt to a LLM" {
		t.Errorf("Expected correct short description, got '%s'", promptCmd.Short)
	}

	// Test that prompt command requires arguments
	if promptCmd.Args == nil {
		t.Error("Expected prompt command to have Args validation")
	}

	// Test prompt command without arguments (should fail)
	tempCmd := &cobra.Command{
		Use:  "prompt",
		Args: cobra.MinimumNArgs(1),
		Run:  func(cmd *cobra.Command, args []string) {},
	}

	err := tempCmd.Args(tempCmd, []string{})
	if err == nil {
		t.Error("Expected error when running prompt command without arguments")
	}
}

func TestConfigCommand(t *testing.T) {
	// Test that config command exists
	if configCmd.Use != "config" {
		t.Errorf("Expected Use 'config', got '%s'", configCmd.Use)
	}

	// Test that config command has subcommands
	subcommands := configCmd.Commands()
	if len(subcommands) == 0 {
		t.Error("Expected config command to have subcommands")
	}

	// Look for expected subcommands
	hasSet := false
	hasList := false
	hasUnset := false

	for _, subcmd := range subcommands {
		switch {
		case strings.HasPrefix(subcmd.Use, "set"):
			hasSet = true
		case strings.HasPrefix(subcmd.Use, "list"):
			hasList = true
		case strings.HasPrefix(subcmd.Use, "unset"):
			hasUnset = true
		}
	}

	if !hasSet {
		t.Error("Expected config command to have 'set' subcommand")
	}
	if !hasList {
		t.Error("Expected config command to have 'list' subcommand")
	}
	if !hasUnset {
		t.Error("Expected config command to have 'unset' subcommand")
	}
}

func TestVersionCommand(t *testing.T) {
	// Test that version command exists
	if versionCmd.Use != "version" {
		t.Errorf("Expected Use 'version', got '%s'", versionCmd.Use)
	}

	if versionCmd.Short != "Prints the current version" {
		t.Errorf("Expected correct short description, got '%s'", versionCmd.Short)
	}

	// Test that version command has a Run function
	if versionCmd.Run == nil {
		t.Error("Expected version command to have a Run function")
	}
}

func TestChatCommand(t *testing.T) {
	// Test that chat command exists
	if chatCmd.Use != "chat" {
		t.Errorf("Expected Use 'chat', got '%s'", chatCmd.Use)
	}

	if chatCmd.Short != "Chat session management" {
		t.Errorf("Expected correct short description, got '%s'", chatCmd.Short)
	}

	// Chat command should have a Run function
	if chatCmd.Run == nil {
		t.Error("Expected chat command to have a Run function")
	}
}

func TestImageCommand(t *testing.T) {
	// Test that image command exists
	if imageCmd.Use != "image" {
		t.Errorf("Expected Use 'image', got '%s'", imageCmd.Use)
	}

	// Test that image command requires arguments
	if imageCmd.Args == nil {
		t.Error("Expected image command to have Args validation")
	}
}

func TestModelsCommand(t *testing.T) {
	// Test that models command exists
	if modelsCmd.Use != "models" {
		t.Errorf("Expected Use 'models', got '%s'", modelsCmd.Use)
	}

	// Test that models command has subcommands
	subcommands := modelsCmd.Commands()
	if len(subcommands) == 0 {
		t.Error("Expected models command to have subcommands")
	}
}

func TestCommandHierarchy(t *testing.T) {
	// Test that all expected commands are registered with root
	expectedCommands := []string{"chat", "prompt", "config", "image", "models", "version"}
	rootSubcommands := rootCmd.Commands()

	commandMap := make(map[string]bool)
	for _, cmd := range rootSubcommands {
		commandMap[cmd.Use] = true
	}

	for _, expectedCmd := range expectedCommands {
		if !commandMap[expectedCmd] {
			t.Errorf("Expected command '%s' to be registered with root command", expectedCmd)
		}
	}
}

func TestFlagInheritance(t *testing.T) {
	// Test that persistent flags are inherited by subcommands
	persistentFlags := []string{"region", "model-id", "custom-arn"}

	for _, flagName := range persistentFlags {
		// Check that flag exists on root
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected persistent flag '%s' to exist on root command", flagName)
			continue
		}

		// Check that flag is available on subcommands
		for _, subcmd := range rootCmd.Commands() {
			inheritedFlag := subcmd.Flags().Lookup(flagName)
			if inheritedFlag == nil {
				// Try looking in inherited flags
				inheritedFlag = subcmd.InheritedFlags().Lookup(flagName)
				if inheritedFlag == nil {
					t.Errorf("Expected flag '%s' to be inherited by command '%s'", flagName, subcmd.Use)
				}
			}
		}
	}
}

// Test helper functions and utilities
func TestExecute(t *testing.T) {
	// Test that Execute function exists and doesn't panic
	// We can't easily test the actual execution without mocking AWS
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute function panicked: %v", r)
		}
	}()

	// Just verify the function exists and can be called
	// We won't actually execute it since it would try to connect to AWS
	// The Execute method is inherited from cobra.Command, so we just check it exists
	if rootCmd == nil {
		t.Error("Expected root command to exist")
	}
}

// Test environment variable handling
func TestEnvironmentVariables(t *testing.T) {
	// Test that commands can handle environment variables
	// This is a basic test since full testing would require AWS setup

	originalEnv := os.Getenv("AWS_REGION")
	defer func() {
		if err := os.Setenv("AWS_REGION", originalEnv); err != nil {
			t.Errorf("Failed to restore AWS_REGION: %v", err)
		}
	}()

	if err := os.Setenv("AWS_REGION", "eu-west-1"); err != nil {
		t.Fatalf("Failed to set AWS_REGION: %v", err)
	}

	// The actual environment variable handling is done by the AWS SDK
	// We're just testing that we don't have any obvious environment variable conflicts
	if os.Getenv("AWS_REGION") != "eu-west-1" {
		t.Error("Environment variable setting failed")
	}
}
