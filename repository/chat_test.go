package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/chat-cli/chat-cli/db"
	"github.com/stretchr/testify/assert"
)

func TestNewChatRepository(t *testing.T) {
	mockDB := &db.MockDatabase{}
	repo := NewChatRepository(mockDB)
	
	assert.NotNil(t, repo)
	assert.Equal(t, mockDB, repo.db)
}

func TestChatRepository_Create(t *testing.T) {
	// Test successful creation
	mockDB := &db.MockDatabase{
		MockPrepareNamedExec: func(query string, arg interface{}) error {
			return nil
		},
	}
	
	repo := NewChatRepository(mockDB)
	
	chat := &Chat{
		ID:        "test-id",
		Message:   "Hello, world!",
		Response:  "Hi there!",
		CreatedAt: time.Now(),
	}
	
	err := repo.Create(chat)
	assert.NoError(t, err)
	
	// Test creation failure
	mockDB.MockPrepareNamedExec = func(query string, arg interface{}) error {
		return errors.New("database error")
	}
	
	err = repo.Create(chat)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestChatRepository_List(t *testing.T) {
	// Test successful list
	mockDB := &db.MockDatabase{
		MockQuery: func(query string, args ...interface{}) ([]map[string]interface{}, error) {
			now := time.Now()
			return []map[string]interface{}{
				{
					"id":         "chat-1",
					"message":    "Hello",
					"response":   "Hi",
					"created_at": now,
				},
				{
					"id":         "chat-2",
					"message":    "How are you?",
					"response":   "I'm good",
					"created_at": now,
				},
			}, nil
		},
	}
	
	repo := NewChatRepository(mockDB)
	
	chats, err := repo.List()
	assert.NoError(t, err)
	assert.Len(t, chats, 2)
	assert.Equal(t, "chat-1", chats[0].ID)
	assert.Equal(t, "Hello", chats[0].Message)
	assert.Equal(t, "Hi", chats[0].Response)
	assert.Equal(t, "chat-2", chats[1].ID)
	
	// Test list with database error
	mockDB.MockQuery = func(query string, args ...interface{}) ([]map[string]interface{}, error) {
		return nil, errors.New("database error")
	}
	
	chats, err = repo.List()
	assert.Error(t, err)
	assert.Nil(t, chats)
	assert.Contains(t, err.Error(), "database error")
}

func TestChatRepository_GetMessages(t *testing.T) {
	// Test successful get messages
	mockDB := &db.MockDatabase{
		MockQuery: func(query string, args ...interface{}) ([]map[string]interface{}, error) {
			now := time.Now()
			return []map[string]interface{}{
				{
					"id":         "msg-1",
					"message":    "Message 1",
					"response":   "Response 1",
					"created_at": now,
				},
				{
					"id":         "msg-2",
					"message":    "Message 2",
					"response":   "Response 2",
					"created_at": now,
				},
			}, nil
		},
	}
	
	repo := NewChatRepository(mockDB)
	
	messages, err := repo.GetMessages("chat-1")
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "msg-1", messages[0].ID)
	assert.Equal(t, "Message 1", messages[0].Message)
	assert.Equal(t, "Response 1", messages[0].Response)
	assert.Equal(t, "msg-2", messages[1].ID)
	
	// Test get messages with database error
	mockDB.MockQuery = func(query string, args ...interface{}) ([]map[string]interface{}, error) {
		return nil, errors.New("database error")
	}
	
	messages, err = repo.GetMessages("chat-1")
	assert.Error(t, err)
	assert.Nil(t, messages)
	assert.Contains(t, err.Error(), "database error")
}