// repository/chat.go
package repository

import (
	"fmt"

	"github.com/chat-cli/chat-cli/db"
)

type Chat struct {
	ID      int
	ChatId  string
	Persona string
	Message string
}

// ChatRepository implements Repository interface for Chat
type ChatRepository struct {
	BaseRepository
}

func NewChatRepository(db db.Database) *ChatRepository {
	return &ChatRepository{
		BaseRepository: BaseRepository{db: db},
	}
}

func (r *ChatRepository) Create(chat *Chat) error {
	query := `
        INSERT INTO chats (chat_id, persona, message)
        VALUES ($1, $2, $3)
        RETURNING id`

	err := r.db.GetDB().QueryRow(query, chat.ChatId, chat.Persona, chat.Message).Scan(&chat.ID)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	return nil
}
