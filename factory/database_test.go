package factory

import (
	"testing"

	"github.com/chat-cli/chat-cli/db"
	"github.com/stretchr/testify/assert"
)

func TestCreateDatabase(t *testing.T) {
	// Test creating database with valid config
	config := db.Config{
		Driver: "sqlite", // Using sqlite as driver for testing
		Name:   "test.db",
	}

	// Use the actual CreateDatabase function
	database, err := CreateDatabase(config)
	
	assert.NoError(t, err)
	assert.NotNil(t, database)
	
	// Test error handling with invalid driver
	invalidConfig := db.Config{
		Driver: "invalid",
		Name:   "test.db",
	}
	
	database, err = CreateDatabase(invalidConfig)
	assert.Error(t, err)
	assert.Nil(t, database)
}