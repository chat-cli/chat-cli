package db

import (
	"database/sql"
)

// MockDB implements the Database interface for testing
type MockDB struct {
	DB *sql.DB
}

// NewMockDB creates a new mock database
func NewMockDB() *MockDB {
	return &MockDB{}
}

// GetDB returns the mock sql.DB
func (m *MockDB) GetDB() *sql.DB {
	return m.DB
}

// Connect mocks connecting to a database
func (m *MockDB) Connect() error {
	return nil
}

// Close mocks closing a database connection
func (m *MockDB) Close() error {
	return nil
}

// Migrate mocks migrating database schema
func (m *MockDB) Migrate() error {
	return nil
}