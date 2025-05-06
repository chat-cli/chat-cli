package factory

import (
	"testing"

	"github.com/chat-cli/chat-cli/db"
	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	// Test creating a database
	config := db.Config{
		Driver: "sqlite",
		Name:   "test.db",
	}

	// Create a database
	database, err := CreateDatabase(config)
	assert.NoError(t, err)
	assert.NotNil(t, database)

	// Test with mock database
	mockDB := &db.MockDatabase{}
	assert.Implements(t, (*db.Database)(nil), mockDB)

	// Test creating with unsupported driver
	invalidConfig := db.Config{
		Driver: "unsupported",
		Name:   "test.db",
	}
	
	database, err = CreateDatabase(invalidConfig)
	assert.Error(t, err)
	assert.Nil(t, database)
}