/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/go-micah/chat-cli/models"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	Long: `Begin an interactive chat session with an LLM via Amazon Bedrock
	
To quit the chat, just type "quit"	
`,

	Run: func(cmd *cobra.Command, args []string) {
		var err error

		modelId, err := cmd.PersistentFlags().GetString("model-id")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// validate model is supported
		m, err := models.GetModel(modelId)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		// check if model supports streaming
		if !m.SupportsStreaming {
			log.Fatalf("model %s does not support streaming so it can't be used with the chat function", m.ModelID)
		}

		// get options
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
		region, err := cmd.Parent().PersistentFlags().GetString("region")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err != nil {
			log.Fatalf("unable to load AWS config: %v", err)
		}

		svc := bedrockruntime.NewFromConfig(cfg)

		conf := types.InferenceConfiguration{
			MaxTokens:   &maxTokens,
			TopP:        &topP,
			Temperature: &temperature,
		}

		converseStreamInput := &bedrockruntime.ConverseStreamInput{
			ModelId:         aws.String(m.ModelID),
			InferenceConfig: &conf,
		}

		// initial prompt
		fmt.Printf("Hi there. You can ask me stuff!\n")

		// tty-loop
		for {

			// gets user input
			prompt := stringPrompt(">")

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

			fmt.Print("[Assistant]: ")

			assistantMsg, err := processStreamingOutput(output, func(ctx context.Context, part string) error {
				fmt.Print(part)
				return nil
			})

			if err != nil {
				log.Fatal("streaming output processing error: ", err)
			}

			converseStreamInput.Messages = append(converseStreamInput.Messages, assistantMsg)

			fmt.Println()

		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.PersistentFlags().StringP("model-id", "m", "anthropic.claude-3-haiku-20240307-v1:0", "set the model id")

	chatCmd.PersistentFlags().Float32("temperature", 1.0, "temperature setting")
	chatCmd.PersistentFlags().Float32("topP", 0.999, "topP setting")
	chatCmd.PersistentFlags().Int32("max-tokens", 500, "max tokens")
}

func stringPrompt(label string) string {

	var s string
	bufferSize := 8192

	r := bufio.NewReaderSize(os.Stdin, bufferSize)

	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}

	return s
}
