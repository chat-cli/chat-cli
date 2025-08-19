/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/chat-cli/chat-cli/errors"
	"github.com/chat-cli/chat-cli/logging"
	"github.com/chat-cli/chat-cli/validation"
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize error handling and logging based on flags
		if err := initializeErrorHandling(cmd); err != nil {
			return err
		}

		// Perform early validation for AWS configuration
		if err := performEarlyValidation(cmd); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// When root command is called without subcommands, run the chat command
		chatCmd.Run(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Handle the error using structured error handling
		if appErr, ok := err.(*errors.AppError); ok {
			errors.Handle(appErr)
		} else {
			// Wrap non-AppError as a critical error
			criticalErr := errors.NewCriticalError(
				errors.ErrorTypeUnknown,
				"command_execution_failed",
				"Command execution failed",
				"An unexpected error occurred while executing the command",
				err,
			).WithOperation("Execute").WithComponent("root-command")
			
			errors.Handle(criticalErr)
		}
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

	// Add error handling and logging flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose error reporting with technical details")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode with detailed logging and stack traces")
	rootCmd.PersistentFlags().String("log-level", "info", "set log level (debug, info, warn, error)")
}

// initializeErrorHandling sets up error handling and logging based on command flags
func initializeErrorHandling(cmd *cobra.Command) error {
	// Get flag values
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return errors.NewConfigurationError(
			"flag_parse_error",
			"Failed to parse verbose flag",
			"Unable to parse command line flags",
			err,
		).WithOperation("ParseFlags").WithComponent("root-command")
	}

	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return errors.NewConfigurationError(
			"flag_parse_error",
			"Failed to parse debug flag",
			"Unable to parse command line flags",
			err,
		).WithOperation("ParseFlags").WithComponent("root-command")
	}

	logLevel, err := cmd.Flags().GetString("log-level")
	if err != nil {
		return errors.NewConfigurationError(
			"flag_parse_error",
			"Failed to parse log-level flag",
			"Unable to parse command line flags",
			err,
		).WithOperation("ParseFlags").WithComponent("root-command")
	}

	// Configure the global error handler
	errorHandler := errors.GetGlobalHandler()
	errorHandler.SetVerbose(verbose)
	errorHandler.SetDebug(debug)

	// Update logging configuration
	if debug {
		logging.EnableDebugMode()
	} else if verbose {
		logging.EnableVerboseMode()
	} else {
		// Always validate and set the log level, even if it's the default
		if err := logging.UpdateLogLevel(logLevel); err != nil {
			return errors.NewConfigurationError(
				"invalid_log_level",
				"Invalid log level specified",
				"Invalid log level. Valid options are: debug, info, warn, error",
				err,
			).WithOperation("UpdateLogLevel").WithComponent("root-command").
				WithMetadata("provided_log_level", logLevel)
		}
	}

	return nil
}

// performEarlyValidation performs early validation of AWS configuration and credentials
func performEarlyValidation(cmd *cobra.Command) error {
	// Get region from flags
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return errors.NewConfigurationError(
			"flag_parse_error",
			"Failed to parse region flag",
			"Unable to parse AWS region from command line flags",
			err,
		).WithOperation("ParseFlags").WithComponent("root-command")
	}

	// Create AWS configuration validator
	awsValidator := validation.NewAWSConfigValidator(region)

	// Perform validation with context
	ctx := context.Background()
	if err := awsValidator.Validate(ctx); err != nil {
		// The validator already returns properly structured AppError instances
		return err
	}

	// Store the validated AWS config for later use by commands
	// This could be stored in a global variable or passed through context
	// For now, we'll just validate and let individual commands load their own config

	return nil
}