package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileManager(t *testing.T) {
	// Test creation with valid app name
	fm, err := NewFileManager("test-app")
	assert.NoError(t, err)
	assert.NotNil(t, fm)
	assert.Equal(t, "test-app", fm.appName)

	// Test creation with empty app name
	_, err = NewFileManager("")
	assert.Error(t, err)
}

func TestGetDBPath(t *testing.T) {
	// Create test file manager
	fm, err := NewFileManager("test-app")
	assert.NoError(t, err)
	
	// Mock the home directory path
	tempHomeDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHomeDir)
	
	// Initialize paths
	err = fm.initializePaths()
	assert.NoError(t, err)
	
	// Test GetDBPath
	dbPath := fm.GetDBPath()
	expectedPath := filepath.Join(tempHomeDir, ".test-app", "test-app.db")
	assert.Equal(t, expectedPath, dbPath)
}

func TestGetDBDriver(t *testing.T) {
	fm, err := NewFileManager("test-app")
	assert.NoError(t, err)
	
	// Test default driver
	driver := fm.GetDBDriver()
	assert.Equal(t, "sqlite", driver)
}

func TestInitializeViper(t *testing.T) {
	// Create test file manager
	fm, err := NewFileManager("test-app")
	assert.NoError(t, err)
	
	// Mock the home directory path
	tempHomeDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHomeDir)
	
	// Initialize paths
	err = fm.initializePaths()
	assert.NoError(t, err)
	
	// Test InitializeViper
	err = fm.InitializeViper()
	assert.NoError(t, err)
	
	// Verify config file was created
	configPath := filepath.Join(tempHomeDir, ".test-app", "config.yaml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}