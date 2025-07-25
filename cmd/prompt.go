/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/chat-cli/chat-cli/utils" //nolint:goimports // false positive from CI version diff
	"github.com/spf13/cobra"

	conf "github.com/chat-cli/chat-cli/config"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Send a prompt to a LLM",
	Long: `Allows you to send a one-line prompt to Amazon Bedrock like so:

> chat-cli prompt "What is your name?"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		prompt := args[0]

		document, err := utils.LoadDocument()
		if err != nil {
			log.Fatalf("unable to load document: %v", err)
		}
		prompt += document

		// Initialize configuration
		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if initErr := fm.InitializeViper(); initErr != nil {
			log.Fatal(initErr)
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

		modelIdFlag, err := cmd.PersistentFlags().GetString("model-id")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// get feature flag for image attachment
		image, err := cmd.PersistentFlags().GetString("image")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// check if --no-stream is set
		noStream, err := cmd.PersistentFlags().GetBool("no-stream")
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

		var modelIdString string

		bedrockSvc := bedrock.NewFromConfig(cfg)

		if customArn == "" {
			// Using model-id, need to validate with Bedrock
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

			// check if model supports image/vision capabilities
			if (image != "") && (!slices.Contains(model.ModelDetails.InputModalities, "IMAGE")) {
				log.Fatalf("model %s does not support images as input. please use a different model", *model.ModelDetails.ModelId)
			}

			// check if model supports streaming and --no-stream is not set
			if (!noStream) && (!*model.ModelDetails.ResponseStreamingSupported) {
				log.Fatalf("model %s does not support streaming. please use the --no-stream flag", *model.ModelDetails.ModelId)
			}

			modelIdString = *model.ModelDetails.ModelId
		} else {
			// Using custom-arn, skip validation and use directly
			modelIdString = finalModelId
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

		svc := bedrockruntime.NewFromConfig(cfg)

		// craft prompt
		userMsg := types.Message{
			Role: types.ConversationRoleUser,
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: prompt,
				},
			},
		}

		// attach image if we have one
		if image != "" {
			imageBytes, imageType, err := utils.ReadImage(image)
			if err != nil {
				log.Fatalf("unable to read image: %v", err)
			}

			userMsg.Content = append(userMsg.Content, &types.ContentBlockMemberImage{
				Value: types.ImageBlock{
					Format: types.ImageFormat(imageType),
					Source: &types.ImageSourceMemberBytes{
						Value: imageBytes,
					},
				},
			})

		}

		conf := types.InferenceConfiguration{
			MaxTokens:   &maxTokens,
			TopP:        &topP,
			Temperature: &temperature,
		}

		if noStream {
			// set up ConverseInput with model and prompt
			converseInput := &bedrockruntime.ConverseInput{
				ModelId:         &modelIdString,
				InferenceConfig: &conf,
			}
			converseInput.Messages = append(converseInput.Messages, userMsg)

			// invoke and wait for full response
			output, err := svc.Converse(context.TODO(), converseInput)
			if err != nil {
				log.Fatalf("error from Bedrock, %v", err)
			}

			response, _ := output.Output.(*types.ConverseOutputMemberMessage)
			responseContentBlock := response.Value.Content[0]
			text, _ := responseContentBlock.(*types.ContentBlockMemberText)

			fmt.Println(text.Value)

		} else {
			converseStreamInput := &bedrockruntime.ConverseStreamInput{
				ModelId:         &modelIdString,
				InferenceConfig: &conf,
			}
			converseStreamInput.Messages = append(converseStreamInput.Messages, userMsg)

			// invoke with streaming response
			output, err := svc.ConverseStream(context.Background(), converseStreamInput)
			if err != nil {
				log.Fatalf("error from Bedrock, %v", err)
			}

			_, err = utils.ProcessStreamingOutput(output, func(ctx context.Context, part string) error {
				fmt.Print(part)
				return nil
			})
			if err != nil {
				log.Fatal("streaming output processing error: ", err)
			}

			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)
	promptCmd.PersistentFlags().StringP("model-id", "m", "anthropic.claude-3-5-sonnet-20240620-v1:0", "set the model id")
	promptCmd.PersistentFlags().String("custom-arn", "", "pass a custom arn from bedrock marketplace or cross-region inference")

	promptCmd.PersistentFlags().StringP("image", "i", "", "path to image")
	promptCmd.PersistentFlags().Bool("no-stream", false, "return the full response once it has completed")

	promptCmd.PersistentFlags().Float32("temperature", 1.0, "temperature setting")
	promptCmd.PersistentFlags().Float32("topP", 0.999, "topP setting")
	promptCmd.PersistentFlags().Int32("max-tokens", 500, "max tokens")
}
