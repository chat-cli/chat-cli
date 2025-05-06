package db

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	MigrateCalled bool
	CloseCalled   bool
	ExecuteCalled bool
	QueryCalled   bool
}

func (m *MockDatabase) Migrate() error {
	m.MigrateCalled = true
	return nil
}

func (m *MockDatabase) Close() error {
	m.CloseCalled = true
	return nil
}

func (m *MockDatabase) Execute(query string, args ...interface{}) error {
	m.ExecuteCalled = true
	return nil
}

func (m *MockDatabase) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	m.QueryCalled = true
	return []map[string]interface{}{}, nil
}