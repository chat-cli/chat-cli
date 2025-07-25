/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	conf "github.com/chat-cli/chat-cli/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  `Manage configuration settings for chat-cli. You can set, unset, and list configuration values.`,
}

// configSetCmd represents the config set command
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  `Set a configuration value. Supported keys: custom-arn, model-id`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize configuration
		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if initErr := fm.InitializeViper(); initErr != nil {
			log.Fatal(initErr)
		}

		key := args[0]
		value := args[1]

		// Validate supported keys
		supportedKeys := map[string]bool{
			"custom-arn": true,
			"model-id":   true,
		}

		if !supportedKeys[key] {
			fmt.Printf("Error: unsupported configuration key '%s'\n", key)
			fmt.Println("Supported keys: custom-arn, model-id")
			os.Exit(1)
		}

		// Set the configuration value
		viper.Set(key, value)

		// Write the configuration to file
		if err := viper.WriteConfig(); err != nil {
			fmt.Printf("Error writing config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Configuration set: %s = %s\n", key, value)
	},
}

// configUnsetCmd represents the config unset command
var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Unset a configuration value",
	Long:  `Unset (remove) a configuration value. Supported keys: custom-arn, model-id`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize configuration
		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if initErr := fm.InitializeViper(); initErr != nil {
			log.Fatal(initErr)
		}

		key := args[0]

		// Validate supported keys
		supportedKeys := map[string]bool{
			"custom-arn": true,
			"model-id":   true,
		}

		if !supportedKeys[key] {
			fmt.Printf("Error: unsupported configuration key '%s'\n", key)
			fmt.Println("Supported keys: custom-arn, model-id")
			os.Exit(1)
		}

		// Check if the key exists
		if !viper.IsSet(key) {
			fmt.Printf("Configuration key '%s' is not set\n", key)
			return
		}

		// Get config file path
		configPath := filepath.Join(fm.ConfigPath, fm.ConfigFile)

		// Read current config
		var configData map[string]interface{}
		if configFile, readErr := os.ReadFile(configPath); readErr == nil { // nolint:gosec // configPath is from user config directory
			if err := yaml.Unmarshal(configFile, &configData); err != nil {
				log.Printf("Warning: failed to parse config file: %v", err)
			}
		}

		if configData == nil {
			configData = make(map[string]interface{})
		}

		// Remove the key
		delete(configData, key)

		// Write back to file
		yamlData, err := yaml.Marshal(configData)
		if err != nil {
			fmt.Printf("Error marshaling config: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(configPath, yamlData, 0600); err != nil {
			fmt.Printf("Error writing config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Configuration unset: %s\n", key)
	},
}

// configListCmd represents the config list command
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `List all current configuration values.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize configuration
		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if initErr := fm.InitializeViper(); initErr != nil {
			log.Fatal(initErr)
		}

		fmt.Println("Current configuration:")

		// Define the keys we care about
		configKeys := []string{"custom-arn", "model-id"}

		hasConfig := false
		for _, key := range configKeys {
			if viper.IsSet(key) {
				fmt.Printf("  %s = %s\n", key, viper.GetString(key))
				hasConfig = true
			}
		}

		if !hasConfig {
			fmt.Println("  No configuration values set")
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configListCmd)
}
