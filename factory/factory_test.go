package factory

import (
	"testing"
)

// Test database factory functions
func TestGetDatabase(t *testing.T) {
	t.Run("GetDatabase should not panic", func(t *testing.T) {
		// This is a simplified test
		// A proper test would require mocking environment variables
		// or using a test database configuration
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetDatabase panicked: %v", r)
			}
		}()
		
		// This might fail without proper environment setup, but shouldn't panic
		_ = GetDatabase()
	})
}

func TestNewSQLiteDatabase(t *testing.T) {
	t.Run("NewSQLiteDatabase should create a database instance", func(t *testing.T) {
		config := db.Config{
			Driver: "sqlite3",
			Name:   ":memory:", // Use in-memory SQLite for testing
		}
		
		database := NewSQLiteDatabase(config)
		if database == nil {
			t.Error("Expected non-nil database instance")
		}
	})
}