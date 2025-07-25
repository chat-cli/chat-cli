/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/chat-cli/chat-cli/utils"
	"github.com/go-micah/go-bedrock/providers"
	"github.com/spf13/cobra"
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Generate an image with a prompt",
	Long:  `Send a prompt to one of the models on Amazon Bedrock that supports image generation and save the reuslt to disk.`,

	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		prompt := args[0]

		document, err := utils.LoadDocument()
		if err != nil {
			log.Fatalf("unable to load document: %v", err)
		}
		prompt += document

		accept := "*/*"
		contentType := "application/json"

		var bodyString []byte

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

		bedrockSvc := bedrock.NewFromConfig(cfg)

		model, err := bedrockSvc.GetFoundationModel(context.TODO(), &bedrock.GetFoundationModelInput{
			ModelIdentifier: &modelId,
		})
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		// validate model supports image generation
		if !slices.Contains(model.ModelDetails.OutputModalities, "IMAGE") {
			log.Fatalf("model %s does not support image generation. please use a different model", *model.ModelDetails.ModelId)
		}

		// get options
		scale, err := cmd.PersistentFlags().GetFloat64("scale")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		steps, err := cmd.PersistentFlags().GetInt("steps")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		seed, err := cmd.PersistentFlags().GetInt("seed")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		filename, err := cmd.PersistentFlags().GetString("filename")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// serialize body
		switch *model.ModelDetails.ProviderName {
		case "Stability AI":
			body := providers.StabilityAIStableDiffusionInvokeModelInput{
				Prompt: []providers.StabilityAIStableDiffusionTextPrompt{
					{
						Text: prompt,
					},
				},
				Scale: scale,
				Steps: steps,
				Seed:  seed,
			}

			bodyString, err = json.Marshal(body)
			if err != nil {
				log.Fatalf("unable to marshal body: %v", err)
			}
		case "Amazon":
			body := providers.AmazonTitanImageInvokeModelInput{
				TaskType: "TEXT_IMAGE",
				TextToImageParams: providers.AmazonTitanImageInvokeModelInputTextToImageParams{
					Text: prompt,
				},
				ImageGenerationConfig: providers.AmazonTitanImageInvokeModelInputImageGenerationConfig{
					NumberOfImages: 1,
					Scale:          scale,
					Seed:           seed,
				},
			}

			bodyString, err = json.Marshal(body)
			if err != nil {
				log.Fatalf("unable to marshal body: %v", err)
			}
		default:
			log.Fatalf("invalid model: %s", *model.ModelDetails.ModelId)
		}

		svc := bedrockruntime.NewFromConfig(cfg)

		resp, err := svc.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
			Accept:      &accept,
			ModelId:     model.ModelDetails.ModelId,
			ContentType: &contentType,
			Body:        bodyString,
		})
		if err != nil {
			log.Fatalf("error from Bedrock, %v", err)
		}

		// save images to disk
		switch *model.ModelDetails.ProviderName {
		case "Stability AI":
			var out providers.StabilityAIStableDiffusionInvokeModelOutput

			err = json.Unmarshal(resp.Body, &out)
			if err != nil {
				log.Fatalf("unable to unmarshal response from Bedrock: %v", err)
			}

			decoded, decodeErr := utils.DecodeImage(out.Artifacts[0].Base64)
			if decodeErr != nil {
				log.Fatalf("unable to decode image: %v", decodeErr)
			}

			outputFile := fmt.Sprintf("%d.jpg", time.Now().Unix())

			// if we have a filename set, us it instead
			if filename != "" {
				outputFile = filename
			}

			err = os.WriteFile(outputFile, decoded, 0600)
			if err != nil {
				log.Fatalf("error writing to file: %v", err)
			}

			log.Println("image written to file", outputFile)
		case "Amazon":
			var out providers.AmazonTitanImageInvokeModelOutput

			err = json.Unmarshal(resp.Body, &out)
			if err != nil {
				log.Fatalf("unable to unmarshal response from Bedrock: %v", err)
			}

			decoded, err := utils.DecodeImage(out.Images[0])
			if err != nil {
				log.Fatalf("unable to decode image: %v", err)
			}

			outputFile := fmt.Sprintf("%d.jpg", time.Now().Unix())

			// if we have a filename set, us it instead
			if filename != "" {
				outputFile = filename
			}

			err = os.WriteFile(outputFile, decoded, 0600)
			if err != nil {
				log.Fatalf("error writing to file: %v", err)
			}

			log.Println("image written to file", outputFile)
		}
	},
}

func init() {
	rootCmd.AddCommand(imageCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	imageCmd.PersistentFlags().Float64("scale", 10, "Set the scale")
	imageCmd.PersistentFlags().Int("steps", 10, "Set the steps")
	imageCmd.PersistentFlags().Int("seed", 0, "Set the seed")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// imageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	imageCmd.PersistentFlags().StringP("model-id", "m", "amazon.nova-canvas-v1:0", "set the model id")
	imageCmd.PersistentFlags().StringP("filename", "f", "", "provide an output filename")

}
