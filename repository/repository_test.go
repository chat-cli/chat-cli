package repository

import (
	"testing"

	"github.com/chat-cli/chat-cli/db"
	"github.com/stretchr/testify/assert"
)

// Test BaseRepository functionality
func TestBaseRepository(t *testing.T) {
	// Create a mock database
	mockDB := &db.MockDatabase{}
	
	// Create a base repository
	baseRepo := BaseRepository{
		db: mockDB,
	}
	
	// Ensure the database is correctly set
	assert.Equal(t, mockDB, baseRepo.db)
	assert.Implements(t, (*db.Database)(nil), baseRepo.db)
}