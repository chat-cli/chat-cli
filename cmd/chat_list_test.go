package cmd

import (
	"testing"
	"time"

	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/repository"
	"github.com/stretchr/testify/assert"
)

func TestChatListCommand(t *testing.T) {
	// Test that the chatList command has the expected properties
	assert.Equal(t, "list", chatListCmd.Use)
	assert.Equal(t, "List all chats", chatListCmd.Short)
	assert.Contains(t, chatListCmd.Long, "List all chats stored in the database")
}

type mockChatRepository struct {
	mockList        func() ([]repository.Chat, error)
	mockGetMessages func(chatId string) ([]repository.Chat, error)
}

func (m *mockChatRepository) List() ([]repository.Chat, error) {
	if m.mockList != nil {
		return m.mockList()
	}
	return []repository.Chat{}, nil
}

func (m *mockChatRepository) GetMessages(chatId string) ([]repository.Chat, error) {
	if m.mockGetMessages != nil {
		return m.mockGetMessages(chatId)
	}
	return []repository.Chat{}, nil
}

func (m *mockChatRepository) Create(chat *repository.Chat) error {
	return nil
}

func TestListAllChats(t *testing.T) {
	// Create mock chat repository
	now := time.Now()
	mockRepo := &mockChatRepository{
		mockList: func() ([]repository.Chat, error) {
			return []repository.Chat{
				{
					ID:        "chat-1",
					Message:   "Hello",
					Response:  "Hi there",
					CreatedAt: now,
				},
				{
					ID:        "chat-2",
					Message:   "How are you?",
					Response:  "I'm good",
					CreatedAt: now.Add(time.Hour),
				},
			}, nil
		},
	}

	// Test listing chats
	chats, err := listAllChats(mockRepo)
	assert.NoError(t, err)
	assert.Len(t, chats, 2)
	assert.Equal(t, "chat-1", chats[0].ID)
	assert.Equal(t, "Hello", chats[0].Message)
	assert.Equal(t, "Hi there", chats[0].Response)
	assert.Equal(t, "chat-2", chats[1].ID)
}