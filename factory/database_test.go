package factory

import (
	"errors"
	"testing"

	"github.com/chat-cli/chat-cli/db"
	"github.com/stretchr/testify/assert"
)

// Mock db.NewDatabase function for testing
type mockDatabaseFactory struct {
	database db.Database
	err      error
}

func (m *mockDatabaseFactory) NewDatabase(driver string, connectionString string) (db.Database, error) {
	return m.database, m.err
}

func TestCreateDatabase(t *testing.T) {
	// Test successful database creation
	mockDB := &db.MockDatabase{}
	
	// Save original function and restore it after the test
	original := dbFactory
	defer func() { dbFactory = original }()
	
	// Set up the mock factory
	dbFactory = &mockDatabaseFactory{
		database: mockDB,
		err:      nil,
	}
	
	// Test with valid config
	config := db.Config{
		Driver:           "sqlite",
		ConnectionString: ":memory:",
	}
	
	database, err := CreateDatabase(config)
	assert.NoError(t, err)
	assert.Equal(t, mockDB, database)
	
	// Test with error from database factory
	dbFactory = &mockDatabaseFactory{
		database: nil,
		err:      errors.New("database creation error"),
	}
	
	database, err = CreateDatabase(config)
	assert.Error(t, err)
	assert.Nil(t, database)
	assert.Contains(t, err.Error(), "database creation error")
}