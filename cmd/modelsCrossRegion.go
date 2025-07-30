/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/spf13/cobra"
)

// modelsCrossRegionCmd represents the cross-region command
var modelsCrossRegionCmd = &cobra.Command{
	Use:   "cross-region",
	Short: "List models supporting cross-region inference",
	Run: func(cmd *cobra.Command, args []string) {
		listCrossRegionModels()
	},
}

func init() {
	modelsCmd.AddCommand(modelsCrossRegionCmd)
}

func listCrossRegionModels() {
	// Load the default configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	// Create a new Bedrock client
	svc := bedrock.NewFromConfig(cfg)

	// Call the ListInferenceProfiles API to get cross-region inference profiles
	profilesResult, err := svc.ListInferenceProfiles(context.TODO(), &bedrock.ListInferenceProfilesInput{})
	if err != nil {
		fmt.Println("Error listing inference profiles:", err)
		return
	}

	fmt.Println("")

	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print the header
	if _, err := fmt.Fprintln(w, "Name\t Profile ID\t Inference Profile ARN"); err != nil {
		log.Printf("Error writing header: %v", err)
	}

	if _, err := fmt.Fprintln(w, "\t\t"); err != nil {
		log.Printf("Error writing separator: %v", err)
	}

	// Print inference profiles
	for i := range profilesResult.InferenceProfileSummaries {
		profile := &profilesResult.InferenceProfileSummaries[i]

		if _, err := fmt.Fprintf(w, "%s\t %s\t %s\n",
			aws.ToString(profile.InferenceProfileName),
			aws.ToString(profile.InferenceProfileId),
			aws.ToString(profile.InferenceProfileArn)); err != nil {
			log.Printf("Error writing profile data: %v", err)
		}
	}

	// Flush the writer
	if err := w.Flush(); err != nil {
		log.Printf("Error flushing writer: %v", err)
	}
}
