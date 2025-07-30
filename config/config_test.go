package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/viper"
)

func TestNewFileManager(t *testing.T) {
	appName := "test-app"
	fm, err := NewFileManager(appName)
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}

	if fm.AppName != appName {
		t.Errorf("expected AppName %q, got %q", appName, fm.AppName)
	}

	if fm.ConfigFile != "config.yaml" {
		t.Errorf("expected ConfigFile %q, got %q", "config.yaml", fm.ConfigFile)
	}

	if fm.DBFile != "data.db" {
		t.Errorf("expected DBFile %q, got %q", "data.db", fm.DBFile)
	}

	// Test that paths are set correctly
	if fm.ConfigPath == "" {
		t.Error("ConfigPath should not be empty")
	}

	if fm.DataPath == "" {
		t.Error("DataPath should not be empty")
	}

	// Cleanup
	if err := os.RemoveAll(fm.ConfigPath); err != nil {
		t.Errorf("Failed to remove config path: %v", err)
	}
	if err := os.RemoveAll(fm.DataPath); err != nil {
		t.Errorf("Failed to remove data path: %v", err)
	}
}

func TestFileManagerWithEnvironment(t *testing.T) {
	originalEnv := os.Getenv("APP_ENV")
	defer func() {
		if err := os.Setenv("APP_ENV", originalEnv); err != nil {
			t.Errorf("Failed to restore APP_ENV: %v", err)
		}
	}()

	if err := os.Setenv("APP_ENV", "testing"); err != nil {
		t.Fatalf("Failed to set APP_ENV: %v", err)
	}

	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}

	if fm.Environment != "testing" {
		t.Errorf("expected Environment %q, got %q", "testing", fm.Environment)
	}

	// Cleanup
	if err := os.RemoveAll(fm.ConfigPath); err != nil {
		t.Errorf("Failed to remove config path: %v", err)
	}
	if err := os.RemoveAll(fm.DataPath); err != nil {
		t.Errorf("Failed to remove data path: %v", err)
	}
}

func TestFileManagerDefaultEnvironment(t *testing.T) {
	originalEnv := os.Getenv("APP_ENV")
	defer func() {
		if err := os.Setenv("APP_ENV", originalEnv); err != nil {
			t.Errorf("Failed to restore APP_ENV: %v", err)
		}
	}()

	if err := os.Unsetenv("APP_ENV"); err != nil {
		t.Fatalf("Failed to unset APP_ENV: %v", err)
	}

	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}

	if fm.Environment != "development" {
		t.Errorf("expected default Environment %q, got %q", "development", fm.Environment)
	}

	// Cleanup
	if err := os.RemoveAll(fm.ConfigPath); err != nil {
		t.Errorf("Failed to remove config path: %v", err)
	}
	if err := os.RemoveAll(fm.DataPath); err != nil {
		t.Errorf("Failed to remove data path: %v", err)
	}
}

func TestInitializePaths(t *testing.T) {
	fm := &FileManager{
		AppName:     "test-app",
		ConfigFile:  "config.yaml",
		DBFile:      "data.db",
		Environment: "testing",
	}

	err := fm.initializePaths()
	if err != nil {
		t.Fatalf("initializePaths failed: %v", err)
	}

	// Check that directories were created
	if _, err := os.Stat(fm.ConfigPath); os.IsNotExist(err) {
		t.Errorf("ConfigPath directory was not created: %s", fm.ConfigPath)
	}

	if _, err := os.Stat(fm.DataPath); os.IsNotExist(err) {
		t.Errorf("DataPath directory was not created: %s", fm.DataPath)
	}

	// Test OS-specific paths
	switch runtime.GOOS {
	case "windows":
		if !filepath.IsAbs(fm.ConfigPath) {
			t.Error("ConfigPath should be absolute on Windows")
		}
	case "darwin":
		if !filepath.IsAbs(fm.ConfigPath) {
			t.Error("ConfigPath should be absolute on macOS")
		}
	default: // Linux
		if !filepath.IsAbs(fm.ConfigPath) {
			t.Error("ConfigPath should be absolute on Linux")
		}
	}

	// Cleanup
	if err := os.RemoveAll(fm.ConfigPath); err != nil {
		t.Errorf("Failed to remove config path: %v", err)
	}
	if err := os.RemoveAll(fm.DataPath); err != nil {
		t.Errorf("Failed to remove data path: %v", err)
	}
}

func TestGetDBPath(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}

	dbPath := fm.GetDBPath()
	expectedPath := filepath.Join(fm.DataPath, fm.DBFile)

	if dbPath != expectedPath {
		t.Errorf("expected DB path %q, got %q", expectedPath, dbPath)
	}

	// Cleanup
	if err := os.RemoveAll(fm.ConfigPath); err != nil {
		t.Errorf("Failed to remove config path: %v", err)
	}
	if err := os.RemoveAll(fm.DataPath); err != nil {
		t.Errorf("Failed to remove data path: %v", err)
	}
}

func TestGetDBDriver(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}

	// Initialize viper with defaults
	viper.Reset()
	viper.SetDefault("db_driver", "sqlite")

	driver := fm.GetDBDriver()
	if driver != "sqlite" {
		t.Errorf("expected db_driver %q, got %q", "sqlite", driver)
	}

	// Cleanup
	if err := os.RemoveAll(fm.ConfigPath); err != nil {
		t.Errorf("Failed to remove config path: %v", err)
	}
	if err := os.RemoveAll(fm.DataPath); err != nil {
		t.Errorf("Failed to remove data path: %v", err)
	}
	viper.Reset()
}

func TestGetConfigValue(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}

	// Reset viper for clean test
	viper.Reset()

	tests := []struct { //nolint:govet // fieldalignment is a minor test optimization
		name         string
		key          string
		flagValue    interface{}
		defaultValue interface{}
		configValue  interface{}
		expected     interface{}
	}{
		{
			name:         "flag takes precedence",
			key:          "test-key",
			flagValue:    "flag-value",
			defaultValue: "default-value",
			configValue:  "config-value",
			expected:     "flag-value",
		},
		{
			name:         "config value when flag is default",
			key:          "test-key2",
			flagValue:    "default-value",
			defaultValue: "default-value",
			configValue:  "config-value",
			expected:     "config-value",
		},
		{
			name:         "default value when no config or flag",
			key:          "test-key3",
			flagValue:    "",
			defaultValue: "default-value",
			configValue:  nil,
			expected:     "default-value",
		},
		{
			name:         "integer flag value",
			key:          "test-int",
			flagValue:    int32(42),
			defaultValue: int32(0),
			configValue:  int32(100),
			expected:     int32(42),
		},
		{
			name:         "float flag value",
			key:          "test-float",
			flagValue:    float32(3.14),
			defaultValue: float32(0.0),
			configValue:  float32(2.71),
			expected:     float32(3.14),
		},
		{
			name:         "empty string config falls back to default",
			key:          "test-empty-string",
			flagValue:    "",
			defaultValue: "default-value",
			configValue:  "",
			expected:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set config value if provided
			if tt.configValue != nil {
				viper.Set(tt.key, tt.configValue)
			}

			result := fm.GetConfigValue(tt.key, tt.flagValue, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}

			// Clean up for next test
			viper.Set(tt.key, nil)
		})
	}

	// Cleanup
	if err := os.RemoveAll(fm.ConfigPath); err != nil {
		t.Errorf("Failed to remove config path: %v", err)
	}
	if err := os.RemoveAll(fm.DataPath); err != nil {
		t.Errorf("Failed to remove data path: %v", err)
	}
	viper.Reset()
}

func TestInitializeViper(t *testing.T) {
	// Create a temporary directory for this test
	tempDir := t.TempDir()

	fm := &FileManager{
		AppName:     "test-app",
		ConfigFile:  "config.yaml",
		DBFile:      "data.db",
		Environment: "testing",
		ConfigPath:  tempDir,
		DataPath:    tempDir,
	}

	// Reset viper
	viper.Reset()

	err := fm.InitializeViper()
	if err != nil {
		t.Fatalf("InitializeViper failed: %v", err)
	}

	// Check that defaults are set
	if viper.GetString("environment") != "testing" {
		t.Errorf("expected environment %q, got %q", "testing", viper.GetString("environment"))
	}

	expectedDBPath := fm.GetDBPath()
	if viper.GetString("db_path") != expectedDBPath {
		t.Errorf("expected db_path %q, got %q", expectedDBPath, viper.GetString("db_path"))
	}

	if viper.GetString("db_driver") != "sqlite" {
		t.Errorf("expected db_driver %q, got %q", "sqlite", viper.GetString("db_driver"))
	}

	// Check that config file was created
	configPath := filepath.Join(fm.ConfigPath, fm.ConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file was not created: %s", configPath)
	}

	viper.Reset()
}
