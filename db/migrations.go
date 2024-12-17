// db/migrations.go
package db

// Migration defines what any database migration must implement
type Migration interface {
	// MigrateUp creates or updates database schema
	MigrateUp() error
	// MigrateDown rolls back database changes (useful for testing)
	MigrateDown() error
}
