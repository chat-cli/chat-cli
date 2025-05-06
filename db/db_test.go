package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabase(t *testing.T) {
	// Test with SQLite driver
	db, err := NewDatabase("sqlite", ":memory:")
	assert.NoError(t, err)
	assert.NotNil(t, db)
	
	// Test with unsupported driver
	db, err = NewDatabase("unsupported", "connection_string")
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "unsupported database driver")
	
	// Test with invalid connection string
	db, err = NewDatabase("sqlite", "invalid://connection")
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestMigrate(t *testing.T) {
	// Create an in-memory SQLite database for testing
	testDB, err := NewDatabase("sqlite", ":memory:")
	assert.NoError(t, err)
	
	// Run migrations
	err = Migrate(testDB)
	assert.NoError(t, err)
	
	// Verify the chats table was created by querying it
	_, err = testDB.Query("SELECT id, message, response, created_at FROM chats LIMIT 1")
	assert.NoError(t, err)
}