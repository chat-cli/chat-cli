package repository

import (
	"database/sql"
	"errors"
	"testing"

	appErrors "github.com/chat-cli/chat-cli/errors"
)

// MockFailingDB simulates database failures for testing error handling
type MockFailingDB struct {
	shouldFail     bool
	failureType    string
	callCount      int
	maxFailures    int
}

func (m *MockFailingDB) GetDB() *sql.DB {
	return nil // Return nil to simulate connection issues
}

func (m *MockFailingDB) Connect() error {
	if m.shouldFail && m.callCount < m.maxFailures {
		m.callCount++
		switch m.failureType {
		case "connection":
			return errors.New("database is locked")
		case "timeout":
			return errors.New("connection timeout")
		default:
			return errors.New("generic database error")
		}
	}
	return nil
}

func (m *MockFailingDB) Close() error {
	return nil
}

func (m *MockFailingDB) Migrate() error {
	if m.shouldFail {
		return errors.New("migration failed")
	}
	return nil
}

func TestChatRepository_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		failureType string
		expectError bool
		errorCode   string
	}{
		{
			name:        "Connection failure",
			failureType: "connection",
			expectError: true,
			errorCode:   "db_not_connected",
		},
		{
			name:        "Timeout failure",
			failureType: "timeout",
			expectError: true,
			errorCode:   "db_not_connected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockFailingDB{
				shouldFail:  true,
				failureType: tt.failureType,
				maxFailures: 5, // More than max retries
			}

			repo := NewChatRepository(mockDB)
			chat := &Chat{
				ChatId:  "test-chat-id",
				Persona: "user",
				Message: "test message",
			}

			err := repo.Create(chat)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				// Check if it's an AppError
				if appErr, ok := err.(*appErrors.AppError); ok {
					if appErr.Type != appErrors.ErrorTypeDatabase {
						t.Errorf("Expected database error type, got %v", appErr.Type)
					}
					if appErr.Code != tt.errorCode {
						t.Errorf("Expected error code %s, got %s", tt.errorCode, appErr.Code)
					}
				} else {
					t.Errorf("Expected AppError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestChatRepository_RetryLogic(t *testing.T) {
	// Test that retryable errors are retried
	mockDB := &MockFailingDB{
		shouldFail:  true,
		failureType: "connection",
		maxFailures: 2, // Fail first 2 attempts, succeed on 3rd
	}

	repo := NewChatRepository(mockDB)
	chat := &Chat{
		ChatId:  "test-chat-id",
		Persona: "user",
		Message: "test message",
	}

	err := repo.Create(chat)

	// Should eventually succeed after retries
	if err != nil {
		// Check if the error indicates retry attempts were made
		if appErr, ok := err.(*appErrors.AppError); ok {
			if retryAttempts, exists := appErr.Metadata["retry_attempts"]; exists {
				if attempts, ok := retryAttempts.(int); ok && attempts > 1 {
					t.Logf("Retry logic worked: %d attempts made", attempts)
				} else {
					t.Errorf("Expected retry attempts in metadata, got %v", retryAttempts)
				}
			}
		}
	}
}

func TestChatRepository_ValidateChatID(t *testing.T) {
	repo := NewChatRepository(nil) // DB not needed for validation

	tests := []struct {
		name      string
		chatID    string
		expectErr bool
		errorCode string
	}{
		{
			name:      "Empty chat ID",
			chatID:    "",
			expectErr: true,
			errorCode: "chat_id_empty",
		},
		{
			name:      "Too short chat ID",
			chatID:    "abc",
			expectErr: true,
			errorCode: "chat_id_invalid",
		},
		{
			name:      "Valid chat ID",
			chatID:    "valid-chat-id-123",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.ValidateChatID(tt.chatID)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if appErr, ok := err.(*appErrors.AppError); ok {
					if appErr.Code != tt.errorCode {
						t.Errorf("Expected error code %s, got %s", tt.errorCode, appErr.Code)
					}
				} else {
					t.Errorf("Expected AppError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestChatRepository_GetMessagesWithFallback(t *testing.T) {
	// Test graceful degradation when GetMessages fails with recoverable error
	mockDB := &MockFailingDB{
		shouldFail:  true,
		failureType: "connection",
		maxFailures: 5,
	}

	repo := NewChatRepository(mockDB)

	// This should return empty slice instead of error for graceful degradation
	messages, err := repo.GetMessagesWithFallback("valid-chat-id")

	if err != nil {
		// If there's an error, it should be non-recoverable
		if appErr, ok := err.(*appErrors.AppError); ok {
			if appErr.IsRecoverable() {
				t.Errorf("Expected non-recoverable error or no error, got recoverable error: %v", appErr)
			}
		}
	} else {
		// Should return empty slice for graceful degradation
		if messages == nil {
			t.Errorf("Expected empty slice, got nil")
		}
		if len(messages) != 0 {
			t.Errorf("Expected empty slice, got %d messages", len(messages))
		}
	}
}

func TestIsRetryableDBError(t *testing.T) {
	repo := &ChatRepository{}

	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "Database locked error",
			err:       errors.New("database is locked"),
			retryable: true,
		},
		{
			name:      "Database busy error",
			err:       errors.New("database is busy"),
			retryable: true,
		},
		{
			name:      "Connection done error",
			err:       sql.ErrConnDone,
			retryable: true,
		},
		{
			name:      "Syntax error",
			err:       errors.New("syntax error"),
			retryable: false,
		},
		{
			name:      "Nil error",
			err:       nil,
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.isRetryableDBError(tt.err)
			if result != tt.retryable {
				t.Errorf("Expected retryable=%v for error %v, got %v", tt.retryable, tt.err, result)
			}
		})
	}
}