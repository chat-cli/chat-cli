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
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/spf13/cobra"
)

// modelsListCmd represents the list command
var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available models",

	Run: func(cmd *cobra.Command, args []string) {
		listModels()
	},
}

func init() {
	modelsCmd.AddCommand(modelsListCmd)
}

func listModels() {
	// Load the default configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	// Create a new Bedrock client
	svc := bedrock.NewFromConfig(cfg)

	// Call the ListModels API
	result, err := svc.ListFoundationModels(context.TODO(), &bedrock.ListFoundationModelsInput{})
	if err != nil {
		fmt.Println("Error listing models:", err)
		return
	}

	fmt.Println("")

	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print the header
	if _, err := fmt.Fprintln(w, "Provider\t Name\t Model ID\t Status\t Cross-Region"); err != nil {
		log.Printf("Error writing header: %v", err)
	}

	if _, err := fmt.Fprintln(w, "\t\t\t\t"); err != nil {
		log.Printf("Error writing separator: %v", err)
	}

	// Print the models
	for i := range result.ModelSummaries {
		model := &result.ModelSummaries[i]

		// Determine status
		status := "UNKNOWN"
		if model.ModelLifecycle != nil {
			status = string(model.ModelLifecycle.Status)
		}

		// Check for cross-region inference capability
		// Cross-region models typically have ARNs containing 'inference-profile' or 'us.'
		crossRegion := "No"
		modelArn := aws.ToString(model.ModelArn)
		if modelArn != "" && (strings.Contains(modelArn, "inference-profile") || strings.Contains(modelArn, "us.")) {
			crossRegion = "Yes"
		}

		if _, err := fmt.Fprintf(w, "%s\t %s\t %s\t %s\t %s\n",
			aws.ToString(model.ProviderName),
			aws.ToString(model.ModelName),
			aws.ToString(model.ModelId),
			status,
			crossRegion); err != nil {
			log.Printf("Error writing model data: %v", err)
		}
	}

	// Flush the writer
	if err := w.Flush(); err != nil {
		log.Printf("Error flushing writer: %v", err)
	}
}
