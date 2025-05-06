package db

import (
	"fmt"
)

// MockDatabase is a mock implementation of the Database interface for testing
type MockDatabase struct {
	MockExec            func(query string, args ...interface{}) error
	MockQuery           func(query string, args ...interface{}) ([]map[string]interface{}, error)
	MockQueryRow        func(query string, args ...interface{}) (map[string]interface{}, error)
	MockPrepareNamedExec func(query string, arg interface{}) error
}

// Exec mocks the Exec method of the Database interface
func (m *MockDatabase) Exec(query string, args ...interface{}) error {
	if m.MockExec != nil {
		return m.MockExec(query, args...)
	}
	return nil
}

// Query mocks the Query method of the Database interface
func (m *MockDatabase) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	if m.MockQuery != nil {
		return m.MockQuery(query, args...)
	}
	return []map[string]interface{}{}, nil
}

// QueryRow mocks the QueryRow method of the Database interface
func (m *MockDatabase) QueryRow(query string, args ...interface{}) (map[string]interface{}, error) {
	if m.MockQueryRow != nil {
		return m.MockQueryRow(query, args...)
	}
	return map[string]interface{}{}, nil
}

// PrepareNamedExec mocks the PrepareNamedExec method of the Database interface
func (m *MockDatabase) PrepareNamedExec(query string, arg interface{}) error {
	if m.MockPrepareNamedExec != nil {
		return m.MockPrepareNamedExec(query, arg)
	}
	return nil
}