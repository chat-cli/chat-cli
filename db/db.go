package db

import "database/sql"

// Database represents a generic database connection
type Database interface {
	GetDB() *sql.DB
	Connect() error
	Close() error
	Migrate() error
}

// Config holds common database configuration
type Config struct {
	Driver   string
	Host     string
	Port     int
	Name     string
	Username string
	Password string
}
