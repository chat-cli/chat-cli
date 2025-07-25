/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chat-cli",
	Short: "Chat with LLMs from Amazon Bedrock!",
	Long: `Chat-CLI is a command line tool that allows you to chat with LLMs from Amazon Bedrock!

 ██████╗██╗  ██╗ █████╗ ████████╗      ██████╗██╗     ██╗
██╔════╝██║  ██║██╔══██╗╚══██╔══╝     ██╔════╝██║     ██║
██║     ███████║███████║   ██║        ██║     ██║     ██║
██║     ██╔══██║██╔══██║   ██║        ██║     ██║     ██║
╚██████╗██║  ██║██║  ██║   ██║███████╗╚██████╗███████╗██║
 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝╚══════╝ ╚═════╝╚══════╝╚═╝	

Running 'chat-cli' without any commands starts an interactive chat session.
Use 'chat-cli [command]' to access other features like prompts, image generation, and configuration.

To quit a chat session, type "quit" or "/quit"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// When root command is called without subcommands, run the chat command
		chatCmd.Run(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("region", "r", "us-east-1", "set the AWS region")

	// Add chat-specific flags to root command so they work when running chat-cli directly
	rootCmd.PersistentFlags().StringP("model-id", "m", "anthropic.claude-3-5-sonnet-20240620-v1:0", "set the model id")
	rootCmd.PersistentFlags().String("custom-arn", "", "pass a custom arn from bedrock marketplace or cross-region inference")
	rootCmd.PersistentFlags().String("chat-id", "", "pass a valid chat-id to load a previous conversation")
	rootCmd.PersistentFlags().Float32("temperature", 1.0, "temperature setting")
	rootCmd.PersistentFlags().Float32("topP", 0.999, "topP setting")
	rootCmd.PersistentFlags().Int32("max-tokens", 500, "max tokens")
}
