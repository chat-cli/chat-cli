// db/sqlite/migrations.go
package sqlite

import (
	"database/sql"
	"fmt"
)

type SQLiteMigration struct {
	db *sql.DB
}

func NewSQLiteMigration(db *sql.DB) *SQLiteMigration {
	return &SQLiteMigration{db: db}
}

func (m *SQLiteMigration) MigrateUp() error {

	var err error

	// Create users table if it doesn't exist
	usersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    
    -- Create trigger to update the updated_at timestamp
    CREATE TRIGGER IF NOT EXISTS users_updated_at 
    AFTER UPDATE ON users
    BEGIN
        UPDATE users SET updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.id;
    END;`

	_, err = m.db.Exec(usersTable)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	// Create users table if it doesn't exist
	chatsTable := `
	CREATE TABLE IF NOT EXISTS chats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id TEXT NOT NULL,
		persona TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Create trigger to update the updated_at timestamp
	CREATE TRIGGER IF NOT EXISTS chats_updated_at 
	AFTER UPDATE ON chats
	BEGIN
		UPDATE chats SET updated_at = CURRENT_TIMESTAMP
		WHERE id = NEW.id;
	END;`

	_, err = m.db.Exec(chatsTable)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	return nil
}

func (m *SQLiteMigration) MigrateDown() error {
	// Drop the users table and its trigger
	dropTables := `
    DROP TRIGGER IF EXISTS users_updated_at;
    DROP TABLE IF EXISTS users;
	DROP TRIGGER IF EXISTS chats_updated_at;
    DROP TABLE IF EXISTS chats;`

	_, err := m.db.Exec(dropTables)
	if err != nil {
		return fmt.Errorf("error dropping tables: %v", err)
	}

	return nil
}
