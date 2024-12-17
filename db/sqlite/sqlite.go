// db/sqlite/sqlite.go
package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/chat-cli/chat-cli/db"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	db     *sql.DB
	config db.Config
}

func (s *SQLiteDB) Migrate() error {
	migration := NewSQLiteMigration(s.db)
	return migration.MigrateUp()
}

func NewSQLiteDB(config db.Config) *SQLiteDB {
	return &SQLiteDB{config: config}
}

func (s *SQLiteDB) Connect() error {
	db, err := sql.Open("sqlite3", s.config.Name)
	if err != nil {
		return fmt.Errorf("sqlite connection error: %v", err)
	}
	s.db = db
	return nil
}

func (s *SQLiteDB) GetDB() *sql.DB {
	return s.db
}

func (s *SQLiteDB) Close() error {
	return s.db.Close()
}
