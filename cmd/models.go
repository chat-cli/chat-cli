/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// modelsCmd represents the models command
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Configure and list available models",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("models called")
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)
}
