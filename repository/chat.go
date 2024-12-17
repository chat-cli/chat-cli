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
	Created string
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

// Function to list 10 most recent chats
func (r *ChatRepository) List() ([]Chat, error) {
	query := `
        SELECT id, chat_id, persona, message, created_at
        FROM chats
		GROUP BY chat_id
        ORDER BY id DESC
        LIMIT 10`

	rows, err := r.db.GetDB().Query(query)
	if err != nil {
		return nil, fmt.Errorf("error listing chats: %v", err)
	}
	defer rows.Close()

	var chats []Chat
	for rows.Next() {
		var chat Chat
		err := rows.Scan(&chat.ID, &chat.ChatId, &chat.Persona, &chat.Message, &chat.Created)
		if err != nil {
			return nil, fmt.Errorf("error scanning chat: %v", err)
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over chats: %v", err)
	}

	return chats, nil
}

// function to retrieve all messages for a given chat_id
func (r *ChatRepository) GetMessages(chatId string) ([]Chat, error) {
	query := `
        SELECT id, chat_id, persona, message
        FROM chats
        WHERE chat_id = $1
        ORDER BY id ASC`

	rows, err := r.db.GetDB().Query(query, chatId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving messages: %v", err)
	}
	defer rows.Close()

	var chats []Chat
	for rows.Next() {
		var chat Chat
		err := rows.Scan(&chat.ID, &chat.ChatId, &chat.Persona, &chat.Message)
		if err != nil {
			return nil, fmt.Errorf("error scanning chat: %v", err)
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over chats: %v", err)
	}

	return chats, nil
}
