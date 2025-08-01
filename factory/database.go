package factory

import (
	"fmt"

	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/db/sqlite"
)

// CreateDatabase is a factory function that returns the appropriate database implementation
func CreateDatabase(config *db.Config) (db.Database, error) {
	switch config.Driver {
	case "sqlite":
		database := sqlite.NewSQLiteDB(config)
		err := database.Connect()
		return database, err
	// case "postgres":
	// 	database := postgres.NewPostgresDB(config)
	// 	return database, database.Connect()
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
}
