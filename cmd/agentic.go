package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/chat-cli/chat-cli/agents"
	"github.com/spf13/cobra"

	conf "github.com/chat-cli/chat-cli/config"
)

// agenticCmd represents the agentic command for quick file operations
var agenticCmd = &cobra.Command{
	Use:   "agentic [task]",
	Short: "Perform agentic file operations",
	Long: `Perform autonomous file operations using AI agents. 

Examples:
  chat-cli agentic "read the README.md file"
  chat-cli agentic "create a hello.txt file with 'Hello World'"
  chat-cli agentic "list all .go files in the current directory"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		task := args[0]

		// Initialize configuration
		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if err := fm.InitializeViper(); err != nil {
			log.Fatal(err)
		}

		// Get configuration values
		region, err := cmd.Parent().PersistentFlags().GetString("region")
		if err != nil {
			log.Fatalf("unable to get region flag: %v", err)
		}

		modelIdFlag, err := cmd.Parent().PersistentFlags().GetString("model-id")
		if err != nil {
			// This flag might not exist on parent, use default
			modelIdFlag = ""
		}

		// Get configuration values with precedence order (flag -> config -> default)
		modelId := fm.GetConfigValue("model-id", modelIdFlag, "anthropic.claude-3-sonnet-20240229-v1:0").(string)

		// Create file edit agent
		fileAgent, err := agents.NewFileEditAgent(region, modelId)
		if err != nil {
			log.Fatalf("Failed to create file agent: %v", err)
		}

		fmt.Printf("🤖 Running agentic task: %s\n\n", task)

		// Execute the agent
		result, err := fileAgent.Execute(context.Background(), task, nil)
		if err != nil {
			log.Fatalf("Agent execution failed: %v", err)
		}

		// Display results
		if result.Success {
			fmt.Printf("✅ Task completed successfully!\n")
			if result.Message != "" {
				fmt.Printf("📄 Result: %s\n", result.Message)
			}
		} else {
			fmt.Printf("❌ Task failed\n")
			if result.Error != "" {
				fmt.Printf("🚨 Error: %s\n", result.Error)
			}
		}

		if len(result.ToolResults) > 0 {
			fmt.Printf("\n🔧 Tool Results:\n")
			for _, toolResult := range result.ToolResults {
				status := "✅"
				if !toolResult.Success {
					status = "❌"
				}
				fmt.Printf("  %s %s", status, toolResult.ToolName)
				if toolResult.Error != "" {
					fmt.Printf(" - %s", toolResult.Error)
				} else if toolResult.Result != nil {
					// Display tool result content for list_files specifically
					if toolResult.ToolName == "list_files" {
						if resultMap, ok := toolResult.Result.(map[string]interface{}); ok {
							if files, ok := resultMap["files"].([]interface{}); ok {
								fmt.Printf("\n")
								for _, file := range files {
									if fileMap, ok := file.(map[string]interface{}); ok {
										name, _ := fileMap["name"].(string)
										isDir, _ := fileMap["is_directory"].(bool)
										if isDir {
											fmt.Printf("    📁 %s/\n", name)
										} else {
											// Check if it's a markdown file
											if strings.HasSuffix(strings.ToLower(name), ".md") {
												fmt.Printf("    📄 %s (markdown)\n", name)
											} else {
												fmt.Printf("    📄 %s\n", name)
											}
										}
									}
								}
							}
						}
					}
				}
				fmt.Printf("\n")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(agenticCmd)
}
