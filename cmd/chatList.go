/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"fmt"
	"log"

	conf "github.com/chat-cli/chat-cli/config"
	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/factory"
	"github.com/chat-cli/chat-cli/repository"
	"github.com/spf13/cobra"
)

// chatListCmd represents the chatList command
var chatListCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if err := fm.InitializeViper(); err != nil {
			log.Fatal(err)
		}

		// Get SQLite database path
		dbPath := fm.GetDBPath()

		config := db.Config{
			Driver: "sqlite3",
			Name:   dbPath,
		}

		database, err := factory.CreateDatabase(config)
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		defer database.Close()

		// Run migrations to ensure tables exist
		if err := database.Migrate(); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		// Create repositories
		chatRepo := repository.NewChatRepository(database)

		if chats, err := chatRepo.List(); err != nil {
			log.Printf("Failed to create chat: %v", err)
		} else {
			for _, chat := range chats {
				fmt.Printf("%s | %s\n", chat.ChatId, truncate(chat.Message, 30))
			}
		}
	},
}

func init() {
	chatCmd.AddCommand(chatListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chatListCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chatListCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}
