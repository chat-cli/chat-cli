package factory

import (
	"testing"

	"github.com/chat-cli/chat-cli/db"
	appErrors "github.com/chat-cli/chat-cli/errors"
)

func TestCreateDatabase_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		config      *db.Config
		expectError bool
		errorType   appErrors.ErrorType
		errorCode   string
	}{
		{
			name:        "Nil config",
			config:      nil,
			expectError: true,
			errorType:   appErrors.ErrorTypeDatabase,
			errorCode:   "config_nil",
		},
		{
			name: "Unsupported driver",
			config: &db.Config{
				Driver: "mysql",
				Name:   "test.db",
			},
			expectError: true,
			errorType:   appErrors.ErrorTypeDatabase,
			errorCode:   "unsupported_driver",
		},
		{
			name: "Valid SQLite config",
			config: &db.Config{
				Driver: "sqlite",
				Name:   ":memory:",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, err := CreateDatabase(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if appErr, ok := err.(*appErrors.AppError); ok {
					if appErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, appErr.Type)
					}
					if appErr.Code != tt.errorCode {
						t.Errorf("Expected error code %s, got %s", tt.errorCode, appErr.Code)
					}
				} else {
					t.Errorf("Expected AppError, got %T", err)
				}

				if database != nil {
					t.Errorf("Expected nil database on error, got %v", database)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if database == nil {
					t.Errorf("Expected database instance but got nil")
				} else {
					// Clean up
					database.Close()
				}
			}
		})
	}
}

func TestCreateDatabaseWithFallback(t *testing.T) {
	tests := []struct {
		name           string
		config         *db.Config
		expectDatabase bool
		expectError    bool
	}{
		{
			name:           "Nil config - non-recoverable error",
			config:         nil,
			expectDatabase: false,
			expectError:    true,
		},
		{
			name: "Valid config",
			config: &db.Config{
				Driver: "sqlite",
				Name:   ":memory:",
			},
			expectDatabase: true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, err := CreateDatabaseWithFallback(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if tt.expectDatabase {
				if database == nil {
					t.Errorf("Expected database instance but got nil")
				} else {
					database.Close()
				}
			} else {
				if database != nil {
					t.Errorf("Expected nil database but got %v", database)
					database.Close()
				}
			}
		})
	}
}

func TestValidateDatabaseConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *db.Config
		expectErr bool
		errorCode string
	}{
		{
			name:      "Nil config",
			config:    nil,
			expectErr: true,
			errorCode: "config_nil",
		},
		{
			name: "Empty driver",
			config: &db.Config{
				Driver: "",
				Name:   "test.db",
			},
			expectErr: true,
			errorCode: "driver_empty",
		},
		{
			name: "Empty name",
			config: &db.Config{
				Driver: "sqlite",
				Name:   "",
			},
			expectErr: true,
			errorCode: "name_empty",
		},
		{
			name: "Valid config",
			config: &db.Config{
				Driver: "sqlite",
				Name:   "test.db",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDatabaseConfig(tt.config)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if appErr, ok := err.(*appErrors.AppError); ok {
					if appErr.Type != appErrors.ErrorTypeValidation {
						t.Errorf("Expected validation error type, got %v", appErr.Type)
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