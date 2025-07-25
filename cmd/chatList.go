/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	conf "github.com/chat-cli/chat-cli/config"
	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/factory"
	"github.com/chat-cli/chat-cli/repository"
	"github.com/spf13/cobra"
)

// chatListCmd represents the chatList command
var chatListCmd = &cobra.Command{
	Use:   "list",
	Short: "Prints a list of recent chats and IDs",

	Run: func(cmd *cobra.Command, args []string) {

		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if initErr := fm.InitializeViper(); initErr != nil {
			log.Fatal(initErr)
		}

		// Get SQLite database path
		dbPath := fm.GetDBPath()

		// Get the database driver from the configuration
		driver := fm.GetDBDriver()

		config := db.Config{
			Driver: driver,
			Name:   dbPath,
		}

		database, err := factory.CreateDatabase(&config)
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		defer func() {
			if err := database.Close(); err != nil {
				log.Printf("Warning: failed to close database: %v", err)
			}
		}()

		// Run migrations to ensure tables exist
		if err := database.Migrate(); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		// Create repositories
		chatRepo := repository.NewChatRepository(database)

		if chats, err := chatRepo.List(); err != nil {
			log.Printf("Failed to create chat: %v", err)
		} else {
			fmt.Println("")

			// Create a new tabwriter
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			// Print the header
			if _, err := fmt.Fprintln(w, "Created Date\t Chat ID\t Title"); err != nil {
				log.Printf("Error writing header: %v", err)
			}

			if _, err := fmt.Fprintln(w, "\t\t"); err != nil {
				log.Printf("Error writing separator: %v", err)
			}

			for _, chat := range chats {
				if _, err := fmt.Fprintf(w, "%s\t %s\t %s\n", chat.Created, chat.ChatId, truncate(chat.Message, 40)); err != nil {
					log.Printf("Error writing chat data: %v", err)
				}
			}
		}
	},
}

func init() {
	chatCmd.AddCommand(chatListCmd)
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "\n"
}
