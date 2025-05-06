package repository

import (
	"testing"

	"github.com/chat-cli/chat-cli/db"
	"github.com/stretchr/testify/assert"
)

func TestNewChatRepository(t *testing.T) {
	// Create a mock database
	mockDB := &db.MockDatabase{}
	
	// Create a chat repository
	chatRepo := NewChatRepository(mockDB)
	
	// Ensure the repository is created correctly
	assert.NotNil(t, chatRepo)
	assert.Equal(t, mockDB, chatRepo.db)
}

func TestChatRepositoryList(t *testing.T) {
	// Create a mock database
	mockDB := &db.MockDatabase{}
	
	// Create a chat repository
	chatRepo := NewChatRepository(mockDB)
	
	// Test the List method
	// Since we're using a mock, we can't actually test the query execution
	// but we can verify that the Query method is called
	_, err := chatRepo.List()
	
	// There will be an error since our mock doesn't return actual data
	assert.Error(t, err)
	assert.True(t, mockDB.QueryCalled)
}

func TestChatRepositoryGetMessages(t *testing.T) {
	// Create a mock database
	mockDB := &db.MockDatabase{}
	
	// Create a chat repository
	chatRepo := NewChatRepository(mockDB)
	
	// Test the GetMessages method
	_, err := chatRepo.GetMessages("test-chat-id")
	
	// There will be an error since our mock doesn't return actual data
	assert.Error(t, err)
	assert.True(t, mockDB.QueryCalled)
}

func TestChatRepositoryCreate(t *testing.T) {
	// Create a mock database
	mockDB := &db.MockDatabase{}
	
	// Create a chat repository
	chatRepo := NewChatRepository(mockDB)
	
	// Create a test chat
	chat := &Chat{
		ChatId:  "test-chat-id",
		Persona: "default",
		Message: "Test message",
		Created: "2023-01-01",
	}
	
	// Test the Create method
	err := chatRepo.Create(chat)
	
	// The mock will return nil error
	assert.NoError(t, err)
	assert.True(t, mockDB.ExecuteCalled)
}