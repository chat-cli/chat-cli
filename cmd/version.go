/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags (GoReleaser / Makefile).
// Local builds without ldflags report "dev".
var version = "dev"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the current version",
	Long:  `Prints the current version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("chat-cli %s, %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
