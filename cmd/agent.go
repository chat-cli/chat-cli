package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/chat-cli/chat-cli/agents"
	"github.com/spf13/cobra"

	conf "github.com/chat-cli/chat-cli/config"
)

var (
	agentRegistry agents.Registry
)

func init() {
	agentRegistry = agents.NewRegistry()
}

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Interact with autonomous agents",
	Long: `The agent command allows you to interact with autonomous agents that can perform tasks.

Available subcommands:
  list    - List all available agents
  run     - Run an agent with a specific task
  info    - Get information about a specific agent`,
}

// agentListCmd lists all available agents
var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available agents",
	Run: func(cmd *cobra.Command, args []string) {
		agents := agentRegistry.ListAgents()

		if len(agents) == 0 {
			fmt.Println("No agents registered")
			return
		}

		fmt.Printf("Available agents (%d):\n\n", len(agents))
		for _, agent := range agents {
			fmt.Printf("Name: %s\n", agent.Name())
			fmt.Printf("Description: %s\n", agent.Description())

			tools := agent.Tools()
			if len(tools) > 0 {
				fmt.Printf("Tools: ")
				for i, tool := range tools {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", tool.Name())
				}
				fmt.Printf("\n")
			}
			fmt.Printf("\n")
		}
	},
}

// agentRunCmd runs an agent with a specific task
var agentRunCmd = &cobra.Command{
	Use:   "run [agent_name] [task]",
	Short: "Run an agent with a specific task",
	Long: `Run an agent with a specific task. If no agent name is provided, 
the system will automatically select the best agent for the task.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var agent agents.Agent
		var task string
		var err error

		if len(args) == 1 {
			// Auto-select agent based on task
			task = args[0]
			agent, err = agentRegistry.FindAgentForTask(task)
			if err != nil {
				log.Fatalf("No suitable agent found: %v", err)
			}
			fmt.Printf("Auto-selected agent: %s\n\n", agent.Name())
		} else {
			// Use specified agent
			agentName := args[0]
			task = args[1]
			agent, err = agentRegistry.GetAgent(agentName)
			if err != nil {
				log.Fatalf("Agent not found: %v", err)
			}
		}

		// Initialize configuration for AWS region and model
		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if err := fm.InitializeViper(); err != nil {
			log.Fatal(err)
		}

		region, err := cmd.Parent().PersistentFlags().GetString("region")
		if err != nil {
			log.Fatalf("unable to get region flag: %v", err)
		}
		_ = region // region is passed via config to agent initialization

		// Get context from flags if provided
		contextStr, _ := cmd.Flags().GetString("context")
		var taskContext map[string]interface{}
		if contextStr != "" {
			if err := json.Unmarshal([]byte(contextStr), &taskContext); err != nil {
				log.Fatalf("Invalid context JSON: %v", err)
			}
		}

		fmt.Printf("Running agent '%s' with task: %s\n", agent.Name(), task)
		fmt.Println("Working...")

		// Execute the agent
		result, err := agent.Execute(context.Background(), task, taskContext)
		if err != nil {
			log.Fatalf("Agent execution failed: %v", err)
		}

		// Display results
		fmt.Printf("\nAgent Result:\n")
		fmt.Printf("Success: %t\n", result.Success)

		if result.Message != "" {
			fmt.Printf("Message: %s\n", result.Message)
		}

		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}

		if len(result.ToolResults) > 0 {
			fmt.Printf("\nTool Results:\n")
			for _, toolResult := range result.ToolResults {
				fmt.Printf("- %s: %t", toolResult.ToolName, toolResult.Success)
				if toolResult.Error != "" {
					fmt.Printf(" (Error: %s)", toolResult.Error)
				}
				fmt.Printf("\n")
			}
		}

		if result.Data != nil {
			fmt.Printf("\nData: %v\n", result.Data)
		}
	},
}

// agentInfoCmd shows information about a specific agent
var agentInfoCmd = &cobra.Command{
	Use:   "info [agent_name]",
	Short: "Get information about a specific agent",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentName := args[0]
		agent, err := agentRegistry.GetAgent(agentName)
		if err != nil {
			log.Fatalf("Agent not found: %v", err)
		}

		fmt.Printf("Agent: %s\n", agent.Name())
		fmt.Printf("Description: %s\n\n", agent.Description())

		tools := agent.Tools()
		if len(tools) > 0 {
			fmt.Printf("Available Tools (%d):\n", len(tools))
			for _, tool := range tools {
				fmt.Printf("\nTool: %s\n", tool.Name())
				fmt.Printf("Description: %s\n", tool.Description())

				schema := tool.Schema()
				if schema != nil {
					schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
					fmt.Printf("Schema: %s\n", string(schemaJSON))
				}
			}
		} else {
			fmt.Println("No tools available")
		}
	},
}

// initializeAgents registers all available agents
func initializeAgents() {
	// Initialize configuration
	fm, err := conf.NewFileManager("chat-cli")
	if err != nil {
		log.Printf("Warning: Could not initialize config manager: %v", err)
		return
	}

	if err := fm.InitializeViper(); err != nil {
		log.Printf("Warning: Could not initialize viper: %v", err)
		return
	}

	// Get default values
	region := fm.GetConfigValue("region", "", "us-east-1").(string)
	modelId := fm.GetConfigValue("model-id", "", "anthropic.claude-3-5-sonnet-20240620-v1:0").(string)

	// Register file edit agent
	fileAgent, err := agents.NewFileEditAgent(region, modelId)
	if err != nil {
		log.Printf("Warning: Could not create file edit agent: %v", err)
		return
	}

	if err := agentRegistry.RegisterAgent(fileAgent); err != nil {
		log.Printf("Warning: Could not register file edit agent: %v", err)
	}
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentRunCmd)
	agentCmd.AddCommand(agentInfoCmd)

	// Add flags
	agentRunCmd.Flags().StringP("context", "c", "", "JSON context to pass to the agent")

	// Initialize agents when the package loads
	initializeAgents()
}
