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
	rootCmd.PersistentFlags().StringP("model-id", "m", DefaultModelID, "set the model id or inference profile id")
	rootCmd.PersistentFlags().String("custom-arn", "", "pass a custom arn from bedrock marketplace or cross-region inference")
	rootCmd.PersistentFlags().String("chat-id", "", "pass a valid chat-id to load a previous conversation")
	rootCmd.PersistentFlags().String("system", "", "set a system prompt")
	rootCmd.PersistentFlags().Bool("no-context-file", false, "disable automatic project-context file discovery (AGENTS.md/CLAUDE.md/etc., chat only)")
	rootCmd.PersistentFlags().Bool("thinking", false, "enable extended thinking / reasoning mode")
	rootCmd.PersistentFlags().Int32("thinking-budget", 1024, "token budget for extended thinking on legacy models (requires --thinking)")
	rootCmd.PersistentFlags().String("thinking-effort", defaultThinkingEffort, "reasoning effort for adaptive models: low, medium, or high (requires --thinking)")
	rootCmd.PersistentFlags().Float32("temperature", 1.0, "optional temperature (0-1); omitted from the request unless set")
	rootCmd.PersistentFlags().Float32("topP", 0.999, "optional top-P (0-1); omitted from the request unless set")
	rootCmd.PersistentFlags().Int32("max-tokens", 500, "max tokens")
}
