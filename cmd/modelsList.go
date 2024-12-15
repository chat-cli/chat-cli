/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"fmt"
	"os"
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
	fmt.Fprintln(w, "Provider\t Name\t Model ID")

	fmt.Fprintln(w, "\t\t")

	// Print the models
	for _, model := range result.ModelSummaries {
		fmt.Fprintf(w, "%s\t %s\t %s\n", aws.ToString(model.ProviderName), aws.ToString(model.ModelName), aws.ToString(model.ModelId))
	}

	// Flush the writer
	w.Flush()
}
