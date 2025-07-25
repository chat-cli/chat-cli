package repository

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// MockDatabase implements the db.Database interface for testing
type MockDatabase struct {
	db *sql.DB
}

func (m *MockDatabase) GetDB() *sql.DB {
	return m.db
}

func (m *MockDatabase) Connect() error {
	return nil // Already connected in setupTestDB
}

func (m *MockDatabase) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
func (m *MockDatabase) Close() error {
	if m.db != nil {
		err := m.db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MockDatabase) Migrate() error {
	return nil // Migration handled in setupTestDB
}

func setupTestDB(t *testing.T) *MockDatabase {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the chats table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS chats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chat_id TEXT NOT NULL,
			persona TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return &MockDatabase{db: db}
}

func TestNewChatRepository(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	repo := NewChatRepository(mockDB)
	if repo == nil {
		t.Error("NewChatRepository returned nil")
	}

	if repo.db == nil {
		t.Error("ChatRepository db field not set correctly")
	}
}

func TestChatRepository_Create(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	repo := NewChatRepository(mockDB)

	chat := &Chat{
		ChatId:  "test-chat-id",
		Persona: "user",
		Message: "Hello, world!",
	}

	err := repo.Create(chat)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}

	if chat.ID == 0 {
		t.Error("Chat ID was not set after creation")
	}

	// Verify the chat was actually inserted
	var count int
	err = mockDB.db.QueryRow("SELECT COUNT(*) FROM chats WHERE chat_id = ?", chat.ChatId).Scan(&count)
	if err != nil {
		t.Errorf("Failed to verify chat insertion: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 chat in database, got %d", count)
	}
}

func TestChatRepository_List(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	repo := NewChatRepository(mockDB)

	// Insert test data
	testChats := []Chat{
		{ChatId: "chat-1", Persona: "user", Message: "Message 1"},
		{ChatId: "chat-2", Persona: "assistant", Message: "Message 2"},
		{ChatId: "chat-3", Persona: "user", Message: "Message 3"},
	}

	for i := range testChats {
		err := repo.Create(&testChats[i])
		if err != nil {
			t.Fatalf("Failed to create test chat %d: %v", i, err)
		}
	}

	// Test List function
	chats, err := repo.List()
	if err != nil {
		t.Errorf("List failed: %v", err)
	}

	if len(chats) != 3 {
		t.Errorf("Expected 3 chats, got %d", len(chats))
	}

	// Check that chats are ordered by ID DESC (most recent first)
	if len(chats) >= 2 && chats[0].ID < chats[1].ID {
		t.Error("Chats are not ordered by ID DESC")
	}
}

func TestChatRepository_GetMessages(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	repo := NewChatRepository(mockDB)

	// Insert test data for multiple chats
	testChats := []Chat{
		{ChatId: "chat-1", Persona: "user", Message: "First message"},
		{ChatId: "chat-1", Persona: "assistant", Message: "Assistant response"},
		{ChatId: "chat-1", Persona: "user", Message: "Second user message"},
		{ChatId: "chat-2", Persona: "user", Message: "Different chat message"},
	}

	for i := range testChats {
		err := repo.Create(&testChats[i])
		if err != nil {
			t.Fatalf("Failed to create test chat %d: %v", i, err)
		}
	}

	// Test GetMessages for chat-1
	messages, err := repo.GetMessages("chat-1")
	if err != nil {
		t.Errorf("GetMessages failed: %v", err)
	}

	if len(messages) != 3 {
		t.Errorf("Expected 3 messages for chat-1, got %d", len(messages))
	}

	// Check that messages are ordered by ID ASC (chronological order)
	if len(messages) >= 2 && messages[0].ID > messages[1].ID {
		t.Error("Messages are not ordered by ID ASC")
	}

	// Verify all messages belong to the correct chat
	for _, msg := range messages {
		if msg.ChatId != "chat-1" {
			t.Errorf("Expected ChatId 'chat-1', got '%s'", msg.ChatId)
		}
	}

	// Test GetMessages for non-existent chat
	emptyMessages, err := repo.GetMessages("non-existent-chat")
	if err != nil {
		t.Errorf("GetMessages failed for non-existent chat: %v", err)
	}

	if len(emptyMessages) != 0 {
		t.Errorf("Expected 0 messages for non-existent chat, got %d", len(emptyMessages))
	}
}

func TestChatRepository_ListLimit(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.Close()

	repo := NewChatRepository(mockDB)

	// Insert more than 10 chats to test the limit
	for i := 1; i <= 15; i++ {
		chat := &Chat{
			ChatId:  "chat-" + string(rune(i)),
			Persona: "user",
			Message: "Message " + string(rune(i)),
		}
		err := repo.Create(chat)
		if err != nil {
			t.Fatalf("Failed to create test chat %d: %v", i, err)
		}
repo := NewChatRepository(mockDB)

	// Insert more than 10 chats to test the limit
	stmt, err := mockDB.db.Prepare("INSERT INTO chats (chat_id, persona, message) VALUES (?, ?, ?)")
	if err != nil {
		t.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for i := 1; i <= 15; i++ {
		chatID := fmt.Sprintf("chat-%d", i)
		message := fmt.Sprintf("Message %d", i)
		_, err := stmt.Exec(chatID, "user", message)
		if err != nil {
			t.Fatalf("Failed to insert test chat %d: %v", i, err)
		}
	}

	// Test that List returns only 10 chats
	chats, err := repo.List()
	if err != nil {
		t.Errorf("List failed: %v", err)
	}

	if len(chats) != 10 {
		t.Errorf("Expected 10 chats (limit), got %d", len(chats))
	}
}

func TestChatStruct(t *testing.T) {
	chat := Chat{
		ID:      1,
		ChatId:  "test-chat",
		Persona: "user",
		Message: "test message",
		Created: "2023-01-01T00:00:00Z",
	}

	if chat.ID != 1 {
		t.Errorf("Expected ID 1, got %d", chat.ID)
	}
	if chat.ChatId != "test-chat" {
		t.Errorf("Expected ChatId 'test-chat', got '%s'", chat.ChatId)
	}
	if chat.Persona != "user" {
		t.Errorf("Expected Persona 'user', got '%s'", chat.Persona)
	}
	if chat.Message != "test message" {
		t.Errorf("Expected Message 'test message', got '%s'", chat.Message)
	}
	if chat.Created != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected Created '2023-01-01T00:00:00Z', got '%s'", chat.Created)
	}
}
