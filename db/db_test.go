package db

import (
	"testing"
)

func TestNewMockDB(t *testing.T) {
	mockDB := NewMockDB()
	if mockDB == nil {
		t.Error("Expected non-nil MockDB")
	}
}

func TestMockDBImplementsInterface(t *testing.T) {
	mockDB := NewMockDB()
	var _ Database = mockDB // Verify MockDB implements Database interface
}

func TestMockDBMethods(t *testing.T) {
	mockDB := NewMockDB()
	
	// Test Connect
	if err := mockDB.Connect(); err != nil {
		t.Errorf("Connect() returned unexpected error: %v", err)
	}
	
	// Test GetDB
	if db := mockDB.GetDB(); db != nil {
		t.Errorf("Expected nil DB, got: %v", db)
	}
	
	// Test Close
	if err := mockDB.Close(); err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
	
	// Test Migrate
	if err := mockDB.Migrate(); err != nil {
		t.Errorf("Migrate() returned unexpected error: %v", err)
	}
}