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
	"github.com/chat-cli/chat-cli/tools"
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

		systemFlag, err := flagCmd.PersistentFlags().GetString("system")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		toolsEnabled, err := flagCmd.PersistentFlags().GetBool("tools")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		thinkingEnabled, err := flagCmd.PersistentFlags().GetBool("thinking")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		thinkingBudget, err := flagCmd.PersistentFlags().GetInt32("thinking-budget")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		thinkingEffort, err := flagCmd.PersistentFlags().GetString("thinking-effort")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}
		thinkingEffort, err = normalizeThinkingEffort(thinkingEffort)
		if err != nil {
			log.Fatal(err)
		}

		noContextFile, err := flagCmd.PersistentFlags().GetBool("no-context-file")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// Get configuration values with precedence order (flag -> config -> default)
		modelId := fm.GetConfigValue("model-id", modelIdFlag, DefaultModelID).(string)
		customArn := fm.GetConfigValue("custom-arn", customArnFlag, "").(string)
		systemPrompt := fm.GetConfigValue("system-prompt", systemFlag, "").(string)

		// #88: when no explicit system prompt was supplied (flag or config),
		// automatically discover a project-context file (AGENTS.md/CLAUDE.md/
		// .github/copilot-instructions.md, or a configured override) and use
		// it as the system prompt, unless disabled via --no-context-file or
		// an empty context-files config value.
		if systemPrompt == "" && !noContextFile {
			contextFilesConfig := fm.GetConfigValue("context-files", "", "").(string)
			candidates := resolveContextFilenames(contextFilesConfig)

			if len(candidates) > 0 {
				if cwd, cwdErr := os.Getwd(); cwdErr == nil {
					content, sourcePath, truncated, found := resolveAndLoadProjectContext(cwd, candidates)
					if found {
						if truncated {
							fmt.Fprintf(os.Stderr, "warning: project context file %s exceeds 32KB and was truncated\n", sourcePath)
						}
						systemPrompt = content
						fmt.Printf("\033[90mUsing project context: %s\033[0m\n", sourcePath)
					}
				}
			}
		}

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

		temperature, err := optionalFloat32Flag(flagCmd.PersistentFlags(), "temperature")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		topP, err := optionalFloat32Flag(flagCmd.PersistentFlags(), "topP")
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

		if customArn == "" && !isInferenceProfileID(finalModelId) {
			// Using a foundation model-id, validate with Bedrock
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
			// Inference profile or custom ARN — pass through to Converse directly
			modelIdString = finalModelId
		}

		svc := bedrockruntime.NewFromConfig(cfg)

		conf := buildInferenceConfiguration(maxTokens, temperature, topP)

		if chatId == "" {
			chatSessionId := uuid.NewV4()
			chatId = chatSessionId.String()
		}

		metadata := map[string]string{
			"chat-session-id": chatId,
		}

		converseStreamInput := &bedrockruntime.ConverseStreamInput{
			ModelId:                      aws.String(modelIdString),
			InferenceConfig:              &conf,
			RequestMetadata:              metadata,
			System:                       withSystemCachePoint(buildSystemContentBlocks(systemPrompt)),
			AdditionalModelRequestFields: buildReasoningConfig(modelIdString, thinkingEnabled, thinkingBudget, thinkingEffort),
		}

		// Tool registry is empty (and therefore inert - ToolConfiguration()
		// returns nil, request shape unchanged) unless --tools is set, since
		// Bedrock has no way to report whether a given model supports tool
		// use and we don't want to break chat for models that don't.
		registry := tools.NewRegistry()
		if toolsEnabled {
			registry.Register(tools.NewReadFileTool())
		}

		sendFn := func(ctx context.Context, in *bedrockruntime.ConverseStreamInput) (<-chan types.ConverseStreamOutput, error) {
			out, streamErr := converseStreamWithFallbacks(ctx, svc, in)
			if streamErr != nil {
				return nil, streamErr
			}
			return out.GetStream().Events(), nil
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

			userMsg := types.Message{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: prompt,
					},
				},
			}

			converseStreamInput.Messages = append(converseStreamInput.Messages, userMsg)

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

			reasoningActive := false
			onText := func(ctx context.Context, part string) error {
				if reasoningActive {
					fmt.Print("\033[0m\n\n")
					reasoningActive = false
				}
				fmt.Print(part)
				return nil
			}

			onReasoning := func(ctx context.Context, part string) error {
				if !reasoningActive {
					fmt.Print("\033[90m[thinking] ")
					reasoningActive = true
				}
				fmt.Print(part)
				return nil
			}

			out, err := runChatTurnWithTools(context.Background(), sendFn, converseStreamInput, registry, onText, onReasoning)
			if err != nil && hasSystemCachePoint(converseStreamInput.System) {
				log.Printf("prompt caching not supported for this request, retrying without it: %v", err)
				converseStreamInput.System = stripSystemCachePoints(converseStreamInput.System)
				out, err = runChatTurnWithTools(context.Background(), sendFn, converseStreamInput, registry, onText, onReasoning)
			}

			if err != nil {
				log.Fatal("streaming output processing error: ", err)
			}

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
