// repository/chat.go
package repository

import (
	"database/sql"
	"time"

	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/errors"
)

type Chat struct { //nolint:govet // fieldalignment is a minor optimization
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
	return r.executeWithRetry("Create", func() error {
		// Check if database connection is available
		if r.db.GetDB() == nil {
			return r.wrapDatabaseError("db_not_connected", "Database connection not available", nil).
				WithOperation("Create").
				WithComponent("ChatRepository").
				WithChatID(chat.ChatId)
		}

		query := `
            INSERT INTO chats (chat_id, persona, message)
            VALUES ($1, $2, $3)
            RETURNING id`

		err := r.db.GetDB().QueryRow(query, chat.ChatId, chat.Persona, chat.Message).Scan(&chat.ID)
		if err != nil {
			return r.wrapDatabaseError("chat_create_failed", "Failed to create chat message", err).
				WithOperation("Create").
				WithComponent("ChatRepository").
				WithChatID(chat.ChatId).
				WithMetadata("persona", chat.Persona)
		}
		return nil
	})
}

// Function to list 10 most recent chats
func (r *ChatRepository) List() ([]Chat, error) {
	var chats []Chat
	err := r.executeWithRetry("List", func() error {
		// Check if database connection is available
		if r.db.GetDB() == nil {
			return r.wrapDatabaseError("db_not_connected", "Database connection not available", nil).
				WithOperation("List").
				WithComponent("ChatRepository")
		}

		query := `
            SELECT id, chat_id, persona, message, created_at
            FROM chats
            GROUP BY chat_id
            ORDER BY id DESC
            LIMIT 10`

		rows, err := r.db.GetDB().Query(query)
		if err != nil {
			return r.wrapDatabaseError("chat_list_failed", "Failed to retrieve chat list", err).
				WithOperation("List").
				WithComponent("ChatRepository")
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				// Log the close error but don't fail the operation
				appErr := r.wrapDatabaseError("rows_close_failed", "Failed to close database rows", closeErr).
					WithOperation("List").
					WithComponent("ChatRepository").
					WithSeverity(errors.ErrorSeverityLow)
				errors.Handle(appErr)
			}
		}()

		chats = nil // Reset slice for retry attempts
		for rows.Next() {
			var chat Chat
			err := rows.Scan(&chat.ID, &chat.ChatId, &chat.Persona, &chat.Message, &chat.Created)
			if err != nil {
				return r.wrapDatabaseError("chat_scan_failed", "Failed to scan chat data", err).
					WithOperation("List").
					WithComponent("ChatRepository")
			}
			chats = append(chats, chat)
		}

		if err := rows.Err(); err != nil {
			return r.wrapDatabaseError("rows_iteration_failed", "Error occurred while reading chat data", err).
				WithOperation("List").
				WithComponent("ChatRepository")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return chats, nil
}

// function to retrieve all messages for a given chat_id
func (r *ChatRepository) GetMessages(chatId string) ([]Chat, error) {
	if chatId == "" {
		return nil, r.wrapDatabaseError("invalid_chat_id", "Chat ID cannot be empty", nil).
			WithOperation("GetMessages").
			WithComponent("ChatRepository").
			WithSeverity(errors.ErrorSeverityMedium)
	}

	var chats []Chat
	err := r.executeWithRetry("GetMessages", func() error {
		// Check if database connection is available
		if r.db.GetDB() == nil {
			return r.wrapDatabaseError("db_not_connected", "Database connection not available", nil).
				WithOperation("GetMessages").
				WithComponent("ChatRepository").
				WithChatID(chatId)
		}

		query := `
            SELECT id, chat_id, persona, message
            FROM chats
            WHERE chat_id = $1
            ORDER BY id ASC`

		rows, err := r.db.GetDB().Query(query, chatId)
		if err != nil {
			return r.wrapDatabaseError("chat_messages_query_failed", "Failed to retrieve chat messages", err).
				WithOperation("GetMessages").
				WithComponent("ChatRepository").
				WithChatID(chatId)
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				// Log the close error but don't fail the operation
				appErr := r.wrapDatabaseError("rows_close_failed", "Failed to close database rows", closeErr).
					WithOperation("GetMessages").
					WithComponent("ChatRepository").
					WithChatID(chatId).
					WithSeverity(errors.ErrorSeverityLow)
				errors.Handle(appErr)
			}
		}()

		chats = nil // Reset slice for retry attempts
		for rows.Next() {
			var chat Chat
			err := rows.Scan(&chat.ID, &chat.ChatId, &chat.Persona, &chat.Message)
			if err != nil {
				return r.wrapDatabaseError("chat_message_scan_failed", "Failed to scan chat message data", err).
					WithOperation("GetMessages").
					WithComponent("ChatRepository").
					WithChatID(chatId)
			}
			chats = append(chats, chat)
		}

		if err := rows.Err(); err != nil {
			return r.wrapDatabaseError("rows_iteration_failed", "Error occurred while reading chat messages", err).
				WithOperation("GetMessages").
				WithComponent("ChatRepository").
				WithChatID(chatId)
		}

		return nil
	})

	if err != nil {
		// Check if this is a "chat not found" scenario (no error but empty results)
		if len(chats) == 0 && err == nil {
			return nil, r.wrapDatabaseError("chat_not_found", "Chat not found", nil).
				WithOperation("GetMessages").
				WithComponent("ChatRepository").
				WithChatID(chatId).
				WithSeverity(errors.ErrorSeverityMedium).
				WithRecoverable(true)
		}
		return nil, err
	}
	return chats, nil
}

// wrapDatabaseError creates a structured database error with appropriate user messages
func (r *ChatRepository) wrapDatabaseError(code, message string, cause error) *errors.AppError {
	userMessage := errors.GetUserMessage(errors.ErrorTypeDatabase, code)
	if userMessage == "" {
		// Fallback to a generic user message if no template exists
		userMessage = "A database error occurred. Please try again."
	}
	
	return errors.NewDatabaseError(code, message, userMessage, cause)
}

// executeWithRetry executes a database operation with retry logic for transient errors
func (r *ChatRepository) executeWithRetry(operation string, fn func() error) error {
	const maxRetries = 3
	const baseDelay = 100 * time.Millisecond

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		// Check if this is an AppError and if it's retryable
		if appErr, ok := err.(*errors.AppError); ok {
			if !errors.IsRetryableError(appErr) {
				return err // Don't retry non-retryable errors
			}
			lastErr = err
		} else {
			// For non-AppError types, check if it's a retryable database error
			if !r.isRetryableDBError(err) {
				return r.wrapDatabaseError("non_retryable_error", "Database operation failed", err).
					WithOperation(operation).
					WithComponent("ChatRepository")
			}
			lastErr = err
		}

		// Don't sleep on the last attempt
		if attempt < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
			time.Sleep(delay)
		}
	}

	// If we get here, all retries failed
	if appErr, ok := lastErr.(*errors.AppError); ok {
		return appErr.WithMetadata("retry_attempts", maxRetries)
	}
	
	return r.wrapDatabaseError("max_retries_exceeded", "Database operation failed after multiple attempts", lastErr).
		WithOperation(operation).
		WithComponent("ChatRepository").
		WithMetadata("retry_attempts", maxRetries)
}

// isRetryableDBError determines if a database error is worth retrying
func (r *ChatRepository) isRetryableDBError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common retryable database errors
	errStr := err.Error()
	
	// SQLite specific retryable errors
	retryablePatterns := []string{
		"database is locked",
		"database is busy",
		"disk I/O error",
		"temporary failure",
		"connection reset",
		"broken pipe",
	}

	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	// Check for specific SQL error types
	if err == sql.ErrConnDone {
		return true
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ValidateChatID validates that a chat ID is in the correct format
func (r *ChatRepository) ValidateChatID(chatID string) error {
	if chatID == "" {
		return r.wrapDatabaseError("chat_id_empty", "Chat ID cannot be empty", nil).
			WithOperation("ValidateChatID").
			WithComponent("ChatRepository").
			WithSeverity(errors.ErrorSeverityMedium)
	}

	// Basic UUID format validation (simplified)
	if len(chatID) < 8 {
		return r.wrapDatabaseError("chat_id_invalid", "Chat ID format is invalid", nil).
			WithOperation("ValidateChatID").
			WithComponent("ChatRepository").
			WithChatID(chatID).
			WithSeverity(errors.ErrorSeverityMedium)
	}

	return nil
}

// GetMessagesWithFallback retrieves messages with graceful degradation
func (r *ChatRepository) GetMessagesWithFallback(chatID string) ([]Chat, error) {
	// Validate chat ID first
	if err := r.ValidateChatID(chatID); err != nil {
		return nil, err
	}

	messages, err := r.GetMessages(chatID)
	if err != nil {
		// Check if this is a recoverable error
		if appErr, ok := err.(*errors.AppError); ok && appErr.IsRecoverable() {
			// Log the error but return empty slice to allow graceful degradation
			errors.Handle(appErr)
			return []Chat{}, nil
		}
		return nil, err
	}

	return messages, nil
}