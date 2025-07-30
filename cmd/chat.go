/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types" //nolint:goimports // false positive from CI version diff
	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/factory"
	"github.com/chat-cli/chat-cli/repository"
	"github.com/chat-cli/chat-cli/utils"
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

		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if initErr := fm.InitializeViper(); initErr != nil {
			log.Fatal(initErr)
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
			log.Fatalf("unable to get flag: %v", err)
		}

		modelIdFlag, err := flagCmd.PersistentFlags().GetString("model-id")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		customArnFlag, err := flagCmd.PersistentFlags().GetString("custom-arn")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// Get configuration values with precedence order (flag -> config -> default)
		modelId := fm.GetConfigValue("model-id", modelIdFlag, "anthropic.claude-3-5-sonnet-20240620-v1:0").(string)
		customArn := fm.GetConfigValue("custom-arn", customArnFlag, "").(string)

		// Ensure custom-arn takes precedence over model-id when both are set
		// If custom-arn is set (from any source), use it; otherwise use model-id
		var finalModelId string
		if customArn != "" {
			finalModelId = customArn
		} else {
			finalModelId = modelId
		}

		chatId, err := flagCmd.PersistentFlags().GetString("chat-id")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		temperature, err := flagCmd.PersistentFlags().GetFloat32("temperature")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		topP, err := flagCmd.PersistentFlags().GetFloat32("topP")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		maxTokens, err := flagCmd.PersistentFlags().GetInt32("max-tokens")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// set up connection to AWS
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err != nil {
			log.Fatalf("unable to load AWS config: %v", err)
		}

		var modelIdString string

		if customArn == "" {
			// Using model-id, need to validate with Bedrock
			bedrockSvc := bedrock.NewFromConfig(cfg)

			// get foundation model details
			model, modelErr := bedrockSvc.GetFoundationModel(context.TODO(), &bedrock.GetFoundationModelInput{
				ModelIdentifier: &finalModelId,
			})
			if modelErr != nil {
				log.Fatalf("error: %v", modelErr)
			}

			// check if this is a text model
			if !slices.Contains(model.ModelDetails.OutputModalities, "TEXT") {
				log.Fatalf("model %s is not a text model, so it can't be used with the chat function", *model.ModelDetails.ModelId)
			}

			// check if model supports streaming
			if !*model.ModelDetails.ResponseStreamingSupported {
				log.Fatalf("model %s does not support streaming so it can't be used with the chat function", *model.ModelDetails.ModelId)
			}

			modelIdString = *model.ModelDetails.ModelId
		} else {
			// Using custom-arn, skip validation and use directly
			modelIdString = finalModelId
		}

		svc := bedrockruntime.NewFromConfig(cfg)

		conf := types.InferenceConfiguration{
			MaxTokens:   &maxTokens,
			TopP:        &topP,
			Temperature: &temperature,
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
		fmt.Printf("💡 Tip: Type '/models' to switch models or '/quit' to exit.\n")
		fmt.Println()

		config := db.Config{
			Driver: driver,
			Name:   dbPath,
		}

		database, err := factory.CreateDatabase(&config)
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		defer func() {
			if err := database.Close(); err != nil {
				log.Printf("Warning: failed to close database: %v", err)
			}
		}()

		// Run migrations to ensure tables exist
		if err := database.Migrate(); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		// Create repositories
		chatRepo := repository.NewChatRepository(database)

		// load saved conversation
		if chatId != "" {
			if chats, err := chatRepo.GetMessages(chatId); err != nil {
				log.Printf("Failed to load messages: %v", err)
			} else {
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

			// handle /models slash command
			if prompt == "/models\n" {
				fmt.Println("\n🔄 Opening model selector...")
				if err := runInteractiveModelSelectorWithRegion(region); err != nil {
					fmt.Printf("Error with models command: %v\n", err)
					continue
				}

				// Reload configuration after model selection
				modelId = fm.GetConfigValue("model-id", modelIdFlag, "anthropic.claude-3-5-sonnet-20240620-v1:0").(string)
				customArn = fm.GetConfigValue("custom-arn", customArnFlag, "").(string)

				// Update final model ID for this conversation
				if customArn != "" {
					finalModelId = customArn
				} else {
					finalModelId = modelId
				}
				modelIdString = finalModelId
				converseStreamInput.ModelId = aws.String(modelIdString)
				fmt.Printf("✓ Model updated for this conversation\n")
				continue
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
				log.Fatal(err)
			}

			// Use the repository without knowing the underlying database type
			chat := &repository.Chat{
				ChatId:  chatId,
				Persona: "User",
				Message: prompt,
			}

			if createErr := chatRepo.Create(chat); createErr != nil {
				log.Printf("Failed to create chat: %v", createErr)
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
				log.Fatal("streaming output processing error: ", err)
			}

			converseStreamInput.Messages = append(converseStreamInput.Messages, assistantMsg)

			chat = &repository.Chat{
				ChatId:  chatId,
				Persona: "Assistant",
				Message: out,
			}

			if err := chatRepo.Create(chat); err != nil {
				log.Printf("Failed to create chat: %v", err)
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
