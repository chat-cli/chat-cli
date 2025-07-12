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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/factory"
	"github.com/chat-cli/chat-cli/repository"
	"github.com/chat-cli/chat-cli/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"

	conf "github.com/chat-cli/chat-cli/config"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	Long: `Begin an interactive chat session with an LLM via Amazon Bedrock
	
To quit the chat, just type "quit"	
`,

	Run: func(cmd *cobra.Command, args []string) {

		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if err := fm.InitializeViper(); err != nil {
			log.Fatal(err)
		}

		// Get SQLite database path
		dbPath := fm.GetDBPath()

		// Get DBDriver from config
		driver := fm.GetDBDriver()

		// get options
		region, err := cmd.Parent().PersistentFlags().GetString("region")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		modelIdFlag, err := cmd.PersistentFlags().GetString("model-id")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		customArnFlag, err := cmd.PersistentFlags().GetString("custom-arn")
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

		chatId, err := cmd.PersistentFlags().GetString("chat-id")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		temperature, err := cmd.PersistentFlags().GetFloat32("temperature")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		topP, err := cmd.PersistentFlags().GetFloat32("topP")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		maxTokens, err := cmd.PersistentFlags().GetInt32("max-tokens")
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
			model, err := bedrockSvc.GetFoundationModel(context.TODO(), &bedrock.GetFoundationModelInput{
				ModelIdentifier: &finalModelId,
			})
			if err != nil {
				log.Fatalf("error: %v", err)
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
		fmt.Printf("Hi there. You can ask me stuff!\n")

		config := db.Config{
			Driver: driver,
			Name:   dbPath,
		}

		database, err := factory.CreateDatabase(config)
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		defer database.Close()

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

			// gets user input
			prompt := utils.StringPrompt(">")

			// check for special words

			// quit the program
			if prompt == "quit\n" {
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
				log.Fatal(err)
			}

			// Use the repository without knowing the underlying database type
			chat := &repository.Chat{
				ChatId:  chatId,
				Persona: "User",
				Message: prompt,
			}

			if err := chatRepo.Create(chat); err != nil {
				log.Printf("Failed to create chat: %v", err)
			}

			fmt.Print("[Assistant]: ")

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

			fmt.Println()

		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.PersistentFlags().StringP("model-id", "m", "anthropic.claude-3-5-sonnet-20240620-v1:0", "set the model id")
	chatCmd.PersistentFlags().String("custom-arn", "", "pass a custom arn from bedrock marketplace or cross-region inference")
	chatCmd.PersistentFlags().String("chat-id", "", "pass a valid chat-id to load a previous conversation")

	chatCmd.PersistentFlags().Float32("temperature", 1.0, "temperature setting")
	chatCmd.PersistentFlags().Float32("topP", 0.999, "topP setting")
	chatCmd.PersistentFlags().Int32("max-tokens", 500, "max tokens")
}
