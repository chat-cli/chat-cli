/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types" //nolint:goimports // false positive from CI version diff
	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/errors"
	"github.com/chat-cli/chat-cli/factory"
	"github.com/chat-cli/chat-cli/repository"
	"github.com/chat-cli/chat-cli/utils"
	"github.com/chat-cli/chat-cli/validation"
	uuid "github.com/satori/go.uuid" //nolint:goimports // false positive from CI version diff
	"github.com/spf13/cobra"

	conf "github.com/chat-cli/chat-cli/config"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Chat session management",
	Long: `Manage your chat sessions and history.

Available subcommands:
  - list: View your recent chat conversations

To start a new interactive chat session, run 'chat-cli' (without the 'chat' subcommand).
To resume an existing conversation, use: chat-cli --chat-id <id>`,

	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			configErr := errors.NewConfigurationError(
				"file_manager_init_failed",
				fmt.Sprintf("Failed to initialize file manager: %v", err),
				"Unable to initialize configuration. Please check your system permissions.",
				err,
			).WithOperation("NewFileManager").WithComponent("chat-command")
			errors.Handle(configErr)
			return
		}

		if initErr := fm.InitializeViper(); initErr != nil {
			configErr := errors.NewConfigurationError(
				"viper_init_failed",
				fmt.Sprintf("Failed to initialize configuration: %v", initErr),
				"Unable to load configuration. Please check your config file permissions.",
				initErr,
			).WithOperation("InitializeViper").WithComponent("chat-command")
			errors.Handle(configErr)
			return
		}

		// Get SQLite database path
		dbPath := fm.GetDBPath()

		// Get DBDriver from config
		driver := fm.GetDBDriver()

		// get options - check if we have a parent (called as subcommand) or not (called from root)
		var flagCmd *cobra.Command
		if cmd.Parent() != nil {
			flagCmd = cmd.Parent()
		} else {
			flagCmd = cmd
		}

		region, err := flagCmd.PersistentFlags().GetString("region")
		if err != nil {
			flagErr := errors.NewValidationError(
				"flag_parse_failed",
				fmt.Sprintf("Unable to get region flag: %v", err),
				"Unable to parse command line flags. Please check your command syntax.",
				err,
			).WithOperation("GetRegionFlag").WithComponent("chat-command")
			errors.Handle(flagErr)
			return
		}

		modelIdFlag, err := flagCmd.PersistentFlags().GetString("model-id")
		if err != nil {
			flagErr := errors.NewValidationError(
				"flag_parse_failed",
				fmt.Sprintf("Unable to get model-id flag: %v", err),
				"Unable to parse command line flags. Please check your command syntax.",
				err,
			).WithOperation("GetModelIdFlag").WithComponent("chat-command")
			errors.Handle(flagErr)
			return
		}

		customArnFlag, err := flagCmd.PersistentFlags().GetString("custom-arn")
		if err != nil {
			flagErr := errors.NewValidationError(
				"flag_parse_failed",
				fmt.Sprintf("Unable to get custom-arn flag: %v", err),
				"Unable to parse command line flags. Please check your command syntax.",
				err,
			).WithOperation("GetCustomArnFlag").WithComponent("chat-command")
			errors.Handle(flagErr)
			return
		}

		// Get configuration values with precedence order (flag -> config -> default)
		modelId := fm.GetConfigValue("model-id", modelIdFlag, "anthropic.claude-3-5-sonnet-20240620-v1:0").(string)
		customArn := fm.GetConfigValue("custom-arn", customArnFlag, "").(string)

		chatId, err := flagCmd.PersistentFlags().GetString("chat-id")
		if err != nil {
			flagErr := errors.NewValidationError(
				"flag_parse_failed",
				fmt.Sprintf("Unable to get chat-id flag: %v", err),
				"Unable to parse command line flags. Please check your command syntax.",
				err,
			).WithOperation("GetChatIdFlag").WithComponent("chat-command")
			errors.Handle(flagErr)
			return
		}

		temperature, err := flagCmd.PersistentFlags().GetFloat32("temperature")
		if err != nil {
			flagErr := errors.NewValidationError(
				"flag_parse_failed",
				fmt.Sprintf("Unable to get temperature flag: %v", err),
				"Unable to parse command line flags. Please check your command syntax.",
				err,
			).WithOperation("GetTemperatureFlag").WithComponent("chat-command")
			errors.Handle(flagErr)
			return
		}

		topP, err := flagCmd.PersistentFlags().GetFloat32("topP")
		if err != nil {
			flagErr := errors.NewValidationError(
				"flag_parse_failed",
				fmt.Sprintf("Unable to get topP flag: %v", err),
				"Unable to parse command line flags. Please check your command syntax.",
				err,
			).WithOperation("GetTopPFlag").WithComponent("chat-command")
			errors.Handle(flagErr)
			return
		}

		maxTokens, err := flagCmd.PersistentFlags().GetInt32("max-tokens")
		if err != nil {
			flagErr := errors.NewValidationError(
				"flag_parse_failed",
				fmt.Sprintf("Unable to get max-tokens flag: %v", err),
				"Unable to parse command line flags. Please check your command syntax.",
				err,
			).WithOperation("GetMaxTokensFlag").WithComponent("chat-command")
			errors.Handle(flagErr)
			return
		}

		// Validate AWS configuration early
		awsValidator := validation.NewAWSConfigValidator(region)
		if err := awsValidator.Validate(ctx); err != nil {
			errors.Handle(err.(*errors.AppError))
			return
		}
		cfg := awsValidator.GetConfig()

		// Validate model before starting chat session
		modelValidator := validation.NewModelValidator(modelId, customArn, region, cfg)
		if err := modelValidator.Validate(ctx); err != nil {
			errors.Handle(err.(*errors.AppError))
			return
		}

		// Use the final model ID (custom ARN takes precedence)
		var modelIdString string
		if customArn != "" {
			modelIdString = customArn
		} else {
			modelIdString = modelId
		}

		svc := bedrockruntime.NewFromConfig(cfg)

		// Build inference configuration
		// Note: Some models (like Claude Sonnet 4.5) don't allow both temperature and topP
		// Only include parameters that were explicitly set or use temperature as default
		conf := types.InferenceConfiguration{
			MaxTokens: &maxTokens,
		}

		// Check if flags were explicitly changed from defaults
		tempChanged := flagCmd.PersistentFlags().Changed("temperature")
		topPChanged := flagCmd.PersistentFlags().Changed("topP")

		if tempChanged && topPChanged {
			// Both explicitly set - use temperature and warn user
			fmt.Println("Warning: Both temperature and topP were set. Some models don't support both parameters simultaneously. Using temperature only.")
			conf.Temperature = &temperature
		} else if topPChanged {
			// Only topP was explicitly set
			conf.TopP = &topP
		} else {
			// Use temperature (either explicitly set or default)
			conf.Temperature = &temperature
		}

		if chatId == "" {
			chatSessionId := uuid.NewV4()
			chatId = chatSessionId.String()
		}

		metadata := map[string]string{
			"chat-session-id": chatId,
		}

		converseStreamInput := &bedrockruntime.ConverseStreamInput{
			ModelId:         aws.String(modelIdString),
			InferenceConfig: &conf,
			RequestMetadata: metadata,
		}

		// initial prompt
		fmt.Println()
		fmt.Printf("Hi there. You can ask me stuff!\n")
		fmt.Println()

		config := db.Config{
			Driver: driver,
			Name:   dbPath,
		}

		database, err := factory.CreateDatabase(&config)
		if err != nil {
			dbErr := errors.NewDatabaseError(
				"database_creation_failed",
				fmt.Sprintf("Failed to create database: %v", err),
				"Unable to initialize chat history database. Chat will work but history won't be saved.",
				err,
			).WithOperation("CreateDatabase").WithComponent("chat-command").
				WithChatID(chatId)
			
			// This is recoverable - we can continue without database
			errors.Handle(dbErr)
			database = nil
		} else {
			defer func() {
				if err := database.Close(); err != nil {
					log.Printf("Warning: failed to close database: %v", err)
				}
			}()

			// Run migrations to ensure tables exist
			if err := database.Migrate(); err != nil {
				dbErr := errors.NewDatabaseError(
					"database_migration_failed",
					fmt.Sprintf("Failed to migrate database: %v", err),
					"Unable to set up chat history database. Chat will work but history won't be saved.",
					err,
				).WithOperation("MigrateDatabase").WithComponent("chat-command").
					WithChatID(chatId)
				
				// This is recoverable - we can continue without database
				errors.Handle(dbErr)
				database = nil
			}
		}

		// Create repositories (only if database is available)
		var chatRepo *repository.ChatRepository
		if database != nil {
			chatRepo = repository.NewChatRepository(database)
		}

		// Load saved conversation with graceful degradation
		if chatId != "" && chatRepo != nil {
			if chats, err := chatRepo.GetMessages(chatId); err != nil {
				historyErr := errors.NewDatabaseError(
					"chat_history_load_failed",
					fmt.Sprintf("Failed to load chat history: %v", err),
					fmt.Sprintf("Unable to load chat history for session '%s'. Starting fresh conversation.", chatId),
					err,
				).WithOperation("LoadChatHistory").WithComponent("chat-command").
					WithChatID(chatId)
				
				// This is recoverable - continue with fresh conversation
				errors.Handle(historyErr)
			} else {
				// Successfully loaded chat history
				if len(chats) > 0 {
					fmt.Printf("Loaded %d previous messages from chat session %s\n", len(chats), chatId)
				}
				
				for _, chat := range chats {
					if chat.Persona == "User" {
						fmt.Printf("[User]: %s\n", chat.Message)
						userMsg := types.Message{
							Role: types.ConversationRoleUser,
							Content: []types.ContentBlock{
								&types.ContentBlockMemberText{
									Value: chat.Message,
								},
							},
						}
						converseStreamInput.Messages = append(converseStreamInput.Messages, userMsg)
					} else {
						fmt.Printf("[Assistant]: %s\n", chat.Message)
						assistantMsg := types.Message{
							Role: types.ConversationRoleAssistant,
							Content: []types.ContentBlock{
								&types.ContentBlockMemberText{
									Value: chat.Message,
								},
							},
						}
						converseStreamInput.Messages = append(converseStreamInput.Messages, assistantMsg)
					}
				}
			}
		} else if chatId != "" && chatRepo == nil {
			// Database not available but chat ID was specified
			fmt.Printf("Note: Chat history is not available due to database issues. Starting fresh conversation.\n")
		}

		// tty-loop
		for {
			// Add a single newline for spacing
			fmt.Println()

			// gets user input with fancy bubble input
			prompt := utils.StringPrompt("")

			// Print the user's input as plain text with gray color
			fmt.Printf("\033[90m> %s\033[0m", strings.TrimSpace(prompt))

			// check for special words

			// quit the program
			if prompt == "quit\n" || prompt == "/quit\n" {
				os.Exit(0)
			}

			userMsg := types.Message{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: prompt,
					},
				},
			}

			converseStreamInput.Messages = append(converseStreamInput.Messages, userMsg)

			output, err := svc.ConverseStream(context.Background(), converseStreamInput)

			if err != nil {
				bedrockErr := errors.NewAWSError(
					"bedrock_converse_failed",
					fmt.Sprintf("Bedrock conversation failed: %v", err),
					"Unable to get response from AI model. Please check your AWS configuration and try again.",
					err,
				).WithOperation("ConverseStream").WithComponent("chat-command").
					WithChatID(chatId).WithMetadata("model_id", modelIdString)
				
				errors.Handle(bedrockErr)
				continue // Continue the chat loop instead of terminating
			}

			// Save user message to database (if available)
			if chatRepo != nil {
				chat := &repository.Chat{
					ChatId:  chatId,
					Persona: "User",
					Message: prompt,
				}

				if createErr := chatRepo.Create(chat); createErr != nil {
					dbErr := errors.NewDatabaseError(
						"chat_save_failed",
						fmt.Sprintf("Failed to save user message: %v", createErr),
						"Unable to save message to chat history. Conversation will continue but won't be saved.",
						createErr,
					).WithOperation("SaveUserMessage").WithComponent("chat-command").
						WithChatID(chatId)
					
					// This is recoverable - continue conversation
					errors.Handle(dbErr)
				}
			}

			// Add an extra line between user message and assistant response
			fmt.Print("\n\n* ")

			var out string

			assistantMsg, err := utils.ProcessStreamingOutput(output, func(ctx context.Context, part string) error {
				fmt.Print(part)
				out += part
				return nil
			})

			if err != nil {
				streamErr := errors.NewAWSError(
					"streaming_output_failed",
					fmt.Sprintf("Streaming output processing error: %v", err),
					"Error processing AI response. Please try your question again.",
					err,
				).WithOperation("ProcessStreamingOutput").WithComponent("chat-command").
					WithChatID(chatId).WithMetadata("model_id", modelIdString)
				
				errors.Handle(streamErr)
				continue // Continue the chat loop instead of terminating
			}

			converseStreamInput.Messages = append(converseStreamInput.Messages, assistantMsg)

			// Save assistant message to database (if available)
			if chatRepo != nil {
				chat := &repository.Chat{
					ChatId:  chatId,
					Persona: "Assistant",
					Message: out,
				}

				if err := chatRepo.Create(chat); err != nil {
					dbErr := errors.NewDatabaseError(
						"chat_save_failed",
						fmt.Sprintf("Failed to save assistant message: %v", err),
						"Unable to save response to chat history. Conversation will continue but won't be saved.",
						err,
					).WithOperation("SaveAssistantMessage").WithComponent("chat-command").
						WithChatID(chatId)
					
					// This is recoverable - continue conversation
					errors.Handle(dbErr)
				}
			}

			// Add extra lines after response for better conversation readability
			fmt.Println()
			fmt.Println()

		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
