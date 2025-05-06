package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	// Test that the root command has the expected properties
	assert.Equal(t, "chat-cli", rootCmd.Use)
	assert.Equal(t, "Chat with LLMs from Amazon Bedrock!", rootCmd.Short)
	assert.Contains(t, rootCmd.Long, "Chat-CLI is a command line tool")
	
	// Test that the root command has the expected flags
	regionFlag := rootCmd.PersistentFlags().Lookup("region")
	assert.NotNil(t, regionFlag)
	assert.Equal(t, "region", regionFlag.Name)
	assert.Equal(t, "r", regionFlag.Shorthand)
	assert.Equal(t, "us-east-1", regionFlag.DefValue)
	assert.Equal(t, "set the AWS region", regionFlag.Usage)
}

func TestExecute(t *testing.T) {
	// Create a command that will succeed for testing Execute
	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			// Do nothing
		},
	}
	
	// Save original rootCmd and restore it after the test
	originalRootCmd := rootCmd
	defer func() { rootCmd = originalRootCmd }()
	
	// Replace rootCmd with our test command
	rootCmd = testCmd
	
	// Test Execute with a command that succeeds
	Execute()
}