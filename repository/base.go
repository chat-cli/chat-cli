// repository/base.go
package repository

import "github.com/chat-cli/chat-cli/db"

// Repository defines the standard operations to be implemented by all repositories
type Repository[T any] interface {
	Create(entity *T) error
	GetByID(id int) (*T, error)
	Update(entity *T) error
	Delete(id int) error
	List() ([]T, error)
}

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
	db db.Database
}
