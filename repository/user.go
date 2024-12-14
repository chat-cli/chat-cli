// repository/user.go
package repository

import (
	"fmt"

	"github.com/chat-cli/chat-cli/db"
)

type User struct {
	ID       int
	Username string
	Email    string
}

// UserRepository implements Repository interface for User
type UserRepository struct {
	BaseRepository
}

func NewUserRepository(db db.Database) *UserRepository {
	return &UserRepository{
		BaseRepository: BaseRepository{db: db},
	}
}

func (r *UserRepository) Create(user *User) error {
	// The query syntax is the same for both SQLite and Postgres in this case
	query := `
        INSERT INTO users (username, email)
        VALUES ($1, $2)
        RETURNING id`

	err := r.db.GetDB().QueryRow(query, user.Username, user.Email).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	return nil
}
