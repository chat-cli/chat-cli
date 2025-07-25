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
type Config struct { //nolint:govet // fieldalignment is a minor optimization
	Port     int
	Driver   string
	Host     string
	Name     string
	Username string
	Password string
}
