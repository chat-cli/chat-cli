package factory

import (
	"fmt"

	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/db/sqlite"
	"github.com/chat-cli/chat-cli/errors"
)

// CreateDatabase is a factory function that returns the appropriate database implementation
func CreateDatabase(config *db.Config) (db.Database, error) {
	if config == nil {
		return nil, errors.NewDatabaseError(
			"config_nil",
			"Database configuration is nil",
			"Database configuration is missing",
			nil,
		).WithOperation("CreateDatabase").
			WithComponent("DatabaseFactory").
			WithSeverity(errors.ErrorSeverityHigh)
	}

	switch config.Driver {
	case "sqlite":
		database := sqlite.NewSQLiteDB(config)
		if err := database.Connect(); err != nil {
			// The error is already wrapped by SQLiteDB.Connect()
			return nil, err
		}
		return database, nil
	// case "postgres":
	// 	database := postgres.NewPostgresDB(config)
	// 	return database, database.Connect()
	default:
		return nil, errors.NewDatabaseError(
			"unsupported_driver",
			fmt.Sprintf("Unsupported database driver: %s", config.Driver),
			fmt.Sprintf("Database driver '%s' is not supported. Please use 'sqlite'.", config.Driver),
			nil,
		).WithOperation("CreateDatabase").
			WithComponent("DatabaseFactory").
			WithMetadata("driver", config.Driver).
			WithSeverity(errors.ErrorSeverityHigh)
	}
}

// CreateDatabaseWithFallback creates a database with graceful degradation
func CreateDatabaseWithFallback(config *db.Config) (db.Database, error) {
	database, err := CreateDatabase(config)
	if err != nil {
		// Check if this is a recoverable error
		if appErr, ok := err.(*errors.AppError); ok {
			// Nil config and unsupported drivers are not recoverable
			if appErr.Code == "config_nil" || appErr.Code == "unsupported_driver" {
				return nil, err
			}
			if appErr.IsRecoverable() {
				// Log the error but allow the application to continue without database
				errors.Handle(appErr)
				return nil, nil // Return nil database to indicate graceful degradation
			}
		}
		return nil, err
	}

	// Test the database with migration
	if err := database.Migrate(); err != nil {
		// Migration failure might be recoverable depending on the error
		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.IsRecoverable() {
				errors.Handle(appErr)
				// Return the database even if migration failed - basic operations might still work
				return database, nil
			}
		}
		// Close the database if migration failed critically
		database.Close()
		return nil, err
	}

	return database, nil
}

// ValidateDatabaseConfig validates the database configuration
func ValidateDatabaseConfig(config *db.Config) error {
	if config == nil {
		return errors.NewValidationError(
			"config_nil",
			"Database configuration is nil",
			"Database configuration is required",
			nil,
		).WithOperation("ValidateDatabaseConfig").
			WithComponent("DatabaseFactory")
	}

	if config.Driver == "" {
		return errors.NewValidationError(
			"driver_empty",
			"Database driver is empty",
			"Database driver must be specified (e.g., 'sqlite')",
			nil,
		).WithOperation("ValidateDatabaseConfig").
			WithComponent("DatabaseFactory")
	}

	if config.Name == "" {
		return errors.NewValidationError(
			"name_empty",
			"Database name is empty",
			"Database name/path must be specified",
			nil,
		).WithOperation("ValidateDatabaseConfig").
			WithComponent("DatabaseFactory")
	}

	return nil
}
