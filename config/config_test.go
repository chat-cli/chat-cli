package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileManager(t *testing.T) {
	// Test creation of a new FileManager
	fm, err := NewFileManager("test-app")
	assert.NoError(t, err)
	assert.NotNil(t, fm)
	assert.Equal(t, "test-app", fm.AppName)
}

func TestInitializeViper(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "config-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up environment for testing
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tempDir)

	// Create a new FileManager
	fm, err := NewFileManager("test-app")
	assert.NoError(t, err)

	// Initialize Viper
	err = fm.InitializeViper()
	assert.NoError(t, err)

	// Verify that the config directory was created
	cfgDir := filepath.Join(tempDir, ".test-app")
	_, err = os.Stat(cfgDir)
	assert.NoError(t, err)
}

func TestGetDBPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "config-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up environment for testing
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tempDir)

	// Create a new FileManager
	fm, err := NewFileManager("test-app")
	assert.NoError(t, err)

	// Test getting the DB path
	dbPath := fm.GetDBPath()
	expectedPath := filepath.Join(tempDir, ".test-app", "test-app.db")
	assert.Equal(t, expectedPath, dbPath)
}

func TestGetDBDriver(t *testing.T) {
	// Create a new FileManager
	fm, err := NewFileManager("test-app")
	assert.NoError(t, err)

	// Test getting the DB driver
	driver := fm.GetDBDriver()
	assert.Equal(t, "sqlite", driver) // Assuming sqlite is the default
}