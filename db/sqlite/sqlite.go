// db/sqlite/sqlite.go
package sqlite

import (
	"database/sql"
	"time"

	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/errors"

	_ "modernc.org/sqlite"
)

type SQLiteDB struct {
	db     *sql.DB
	config db.Config
}

func (s *SQLiteDB) Migrate() error {
	if s.db == nil {
		return s.wrapConnectionError("db_not_connected", "Database connection not established", nil)
	}

	migration := NewSQLiteMigration(s.db)
	if err := migration.MigrateUp(); err != nil {
		userMessage := errors.GetUserMessage(errors.ErrorTypeDatabase, "migration_failed")
		if userMessage == "" {
			userMessage = "Database migration failed. Some features may not work properly."
		}

		return errors.NewDatabaseError("migration_failed", "Database migration failed", userMessage, err).
			WithOperation("Migrate").
			WithComponent("SQLiteDB").
			WithMetadata("database_path", s.config.Name).
			WithSeverity(errors.ErrorSeverityHigh)
	}
	return nil
}

func NewSQLiteDB(config *db.Config) *SQLiteDB {
	return &SQLiteDB{config: *config}
}

func (s *SQLiteDB) Connect() error {
	return s.connectWithRetry()
}

// connectWithRetry attempts to connect to the database with retry logic
func (s *SQLiteDB) connectWithRetry() error {
	const maxRetries = 3
	const baseDelay = 500 * time.Millisecond

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		db, err := sql.Open("sqlite", s.config.Name)
		if err != nil {
			lastErr = err
			if attempt < maxRetries-1 {
				time.Sleep(baseDelay * time.Duration(1<<attempt))
				continue
			}
			break
		}

		// Test the connection
		if err := db.Ping(); err != nil {
			db.Close()
			lastErr = err
			if attempt < maxRetries-1 {
				time.Sleep(baseDelay * time.Duration(1<<attempt))
				continue
			}
			break
		}

		// Configure connection settings for better reliability
		db.SetMaxOpenConns(1) // SQLite works best with single connection
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(time.Hour)

		s.db = db
		return nil
	}

	// All attempts failed
	return s.wrapConnectionError("connection_failed", "Failed to connect to SQLite database", lastErr)
}

// wrapConnectionError creates a structured database connection error
func (s *SQLiteDB) wrapConnectionError(code, message string, cause error) error {
	userMessage := errors.GetUserMessage(errors.ErrorTypeDatabase, code)
	if userMessage == "" {
		userMessage = "Failed to connect to the database. Chat history may not be available."
	}

	return errors.NewDatabaseError(code, message, userMessage, cause).
		WithOperation("Connect").
		WithComponent("SQLiteDB").
		WithMetadata("database_path", s.config.Name).
		WithSeverity(errors.ErrorSeverityHigh)
}

func (s *SQLiteDB) GetDB() *sql.DB {
	return s.db
}

func (s *SQLiteDB) Close() error {
	if s.db == nil {
		return nil // Already closed or never opened
	}

	if err := s.db.Close(); err != nil {
		userMessage := "Failed to properly close database connection"
		return errors.NewDatabaseError("close_failed", "Database close failed", userMessage, err).
			WithOperation("Close").
			WithComponent("SQLiteDB").
			WithSeverity(errors.ErrorSeverityLow)
	}
	
	s.db = nil
	return nil
}
// IsHealthy checks if the database connection is healthy
func (s *SQLiteDB) IsHealthy() error {
	if s.db == nil {
		return s.wrapConnectionError("db_not_connected", "Database connection not established", nil)
	}

	// Test the connection with a simple query
	if err := s.db.Ping(); err != nil {
		return s.wrapConnectionError("connection_unhealthy", "Database connection is not healthy", err)
	}

	return nil
}

// ValidateConnection ensures the database connection is ready for operations
func (s *SQLiteDB) ValidateConnection() error {
	if err := s.IsHealthy(); err != nil {
		// Try to reconnect if the connection is unhealthy
		if reconnectErr := s.connectWithRetry(); reconnectErr != nil {
			return reconnectErr
		}
	}
	return nil
}

// GetConnectionInfo returns information about the current database connection
func (s *SQLiteDB) GetConnectionInfo() map[string]interface{} {
	info := map[string]interface{}{
		"driver":       "sqlite",
		"database_path": s.config.Name,
		"connected":    s.db != nil,
	}

	if s.db != nil {
		stats := s.db.Stats()
		info["open_connections"] = stats.OpenConnections
		info["in_use"] = stats.InUse
		info["idle"] = stats.Idle
	}

	return info
}