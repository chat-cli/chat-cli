/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
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

		// read a document from stdin
		var document string

		if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
			// do nothing
		} else {
			stdin, err := io.ReadAll(os.Stdin)

			if err != nil {
				panic(err)
			}
			document = string(stdin)
		}

		if document != "" {
			document = "<document>\n\n" + document + "\n\n</document>\n\n"
			prompt = document + prompt
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

		modelId, err := cmd.PersistentFlags().GetString("model-id")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// get feature floag for image attachment
		image, err := cmd.PersistentFlags().GetString("image")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// check if --no-stream is set
		noStream, err := cmd.PersistentFlags().GetBool("no-stream")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		customArn, err := cmd.PersistentFlags().GetString("custom-arn")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		var modelIdString string

		bedrockSvc := bedrock.NewFromConfig(cfg)

		if customArn == "" {
			model, err := bedrockSvc.GetFoundationModel(context.TODO(), &bedrock.GetFoundationModelInput{
				ModelIdentifier: &modelId,
			})
			if err != nil {
				log.Fatalf("error: %v", err)
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
			modelIdString = customArn
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
			imageBytes, imageType, err := readImage(image)
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

			reponse, _ := output.Output.(*types.ConverseOutputMemberMessage)
			responseContentBlock := reponse.Value.Content[0]
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

			_, err = processStreamingOutput(output, func(ctx context.Context, part string) error {
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

type StreamingOutputHandler func(ctx context.Context, part string) error

func processStreamingOutput(output *bedrockruntime.ConverseStreamOutput, handler StreamingOutputHandler) (types.Message, error) {

	var combinedResult string

	msg := types.Message{}

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ConverseStreamOutputMemberMessageStart:

			msg.Role = v.Value.Role

		case *types.ConverseStreamOutputMemberContentBlockDelta:

			textResponse := v.Value.Delta.(*types.ContentBlockDeltaMemberText)
			handler(context.Background(), textResponse.Value)
			combinedResult = combinedResult + textResponse.Value

		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)
		}
	}

	msg.Content = append(msg.Content,
		&types.ContentBlockMemberText{
			Value: combinedResult,
		},
	)

	return msg, nil
}

func readImage(filename string) ([]byte, string, error) {

	// Define a base directory for allowed images
	baseDir, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("unable to get working directory: %w", err)
	}

	// Clean the filename and create the full path
	cleanFilename := filepath.Clean(filename)
	fullPath := filepath.Join(baseDir, cleanFilename)

	// Ensure the full path is within the base directory
	relPath, err := filepath.Rel(baseDir, fullPath)
	if err != nil || strings.HasPrefix(relPath, "..") || strings.HasPrefix(relPath, string(filepath.Separator)) {
		return nil, "", fmt.Errorf("access denied: %s is outside of the allowed directory", filename)
	}

	// Check if the file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("file does not exist: %s", filename)
	}

	// Read the file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", fmt.Errorf("unable to read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		ext = ext[1:] // Remove the leading dot
	}

	var imageType string

	switch ext {
	case "jpg":
		imageType = "jpeg"
	case "jpeg":
		imageType = "jpeg"
	case "png":
		imageType = "png"
	case "gif":
		imageType = "gif"
	case "webp":
		imageType = "webp"
	default:
		return nil, "", fmt.Errorf("unsupported file type")

	}

	return data, imageType, nil
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
