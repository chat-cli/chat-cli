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
	"sort"
	"strconv"
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
				selectedModel, err := handleModelsSlashCommand(region)
				if err != nil {
					fmt.Printf("Error with models command: %v\n", err)
					continue
				}
				if selectedModel != "" {
					// Update the model for this conversation
					finalModelId = selectedModel
					modelIdString = selectedModel
					converseStreamInput.ModelId = aws.String(modelIdString)
					fmt.Printf("✓ Switched to model: %s\n", selectedModel)
				}
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

// handleModelsSlashCommand provides an inline model selection interface within chat
func handleModelsSlashCommand(region string) (string, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("error loading AWS configuration: %w", err)
	}

	// Create Bedrock client
	svc := bedrock.NewFromConfig(cfg)

	// Fetch models
	result, err := svc.ListFoundationModels(context.TODO(), &bedrock.ListFoundationModelsInput{})
	if err != nil {
		return "", fmt.Errorf("error listing models: %w", err)
	}

	// Convert models to a simplified list for chat interface
	var models []ModelOption
	seenModels := make(map[string]bool)

	for i := range result.ModelSummaries {
		model := &result.ModelSummaries[i]

		// Only include active models
		if model.ModelLifecycle != nil && string(model.ModelLifecycle.Status) == "ACTIVE" {
			modelID := aws.ToString(model.ModelId)

			// Filter out model variants with capacity/context size suffixes (same logic as interactive)
			if strings.Contains(modelID, ":") {
				parts := strings.Split(modelID, ":")
				if len(parts) >= 3 {
					lastPart := parts[len(parts)-1]
					if isCapacitySuffix(lastPart) {
						continue
					}
				} else if len(parts) == 2 {
					suffix := parts[1]
					if isCapacitySuffix(suffix) {
						continue
					}
				}
			}

			// Avoid duplicates
			if seenModels[modelID] {
				continue
			}
			seenModels[modelID] = true

			modelArn := aws.ToString(model.ModelArn)
			crossRegion := requiresCrossRegionProfile(modelID, modelArn)

			models = append(models, ModelOption{
				ID:          modelID,
				Name:        aws.ToString(model.ModelName),
				Provider:    aws.ToString(model.ProviderName),
				CrossRegion: crossRegion,
				Arn:         modelArn,
			})
		}
	}

	// Sort models by provider, then by name
	sort.Slice(models, func(i, j int) bool {
		if models[i].Provider != models[j].Provider {
			return models[i].Provider < models[j].Provider
		}
		return models[i].Name < models[j].Name
	})

	// Display models in a simple numbered list
	fmt.Println("\n📋 Available Models:")
	fmt.Println("==================")

	for i, model := range models {
		crossRegionIndicator := ""
		if model.CrossRegion {
			crossRegionIndicator = " 🌐"
		}
		fmt.Printf("%2d) %s - %s (%s)%s\n", i+1, model.Provider, model.Name, model.ID, crossRegionIndicator)
	}

	fmt.Println("\n🌐 = Cross-region model")
	fmt.Printf("\nSelect a model (1-%d) or press Enter to cancel: ", len(models))

	// Get user selection
	selection := utils.StringPrompt("")
	selection = strings.TrimSpace(selection)

	if selection == "" {
		fmt.Println("Model selection cancelled.")
		return "", nil
	}

	// Parse selection
	index, err := strconv.Atoi(selection)
	if err != nil || index < 1 || index > len(models) {
		return "", fmt.Errorf("invalid selection: %s", selection)
	}

	selectedModel := models[index-1]

	// For cross-region models, use inference profile ARN
	if selectedModel.CrossRegion {
		inferenceProfileArn := generateInferenceProfileArn(selectedModel.ID)
		if err := setCustomArnInConfig(inferenceProfileArn); err != nil {
			return "", fmt.Errorf("error setting custom ARN: %w", err)
		}
		return inferenceProfileArn, nil
	} else {
		// Regular models use model-id
		if err := setModelInConfig(selectedModel.ID); err != nil {
			return "", fmt.Errorf("error setting model: %w", err)
		}
		return selectedModel.ID, nil
	}
}

// ModelOption represents a model choice for the slash command
type ModelOption struct {
	ID          string
	Name        string
	Provider    string
	CrossRegion bool
	Arn         string
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
