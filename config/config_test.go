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
func TestDefaultErrorConfig(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}
	defer cleanup(t, fm)

	config := fm.DefaultErrorConfig()

	if config.VerboseErrors != false {
		t.Errorf("expected VerboseErrors false, got %v", config.VerboseErrors)
	}

	if config.DebugMode != false {
		t.Errorf("expected DebugMode false, got %v", config.DebugMode)
	}

	if config.LogLevel != "info" {
		t.Errorf("expected LogLevel 'info', got %q", config.LogLevel)
	}

	if config.LogFile != "" {
		t.Errorf("expected LogFile empty, got %q", config.LogFile)
	}

	if config.MaxLogSize != 10*1024*1024 {
		t.Errorf("expected MaxLogSize 10MB, got %d", config.MaxLogSize)
	}

	if config.MaxLogFiles != 5 {
		t.Errorf("expected MaxLogFiles 5, got %d", config.MaxLogFiles)
	}

	if config.EnableColor != true {
		t.Errorf("expected EnableColor true, got %v", config.EnableColor)
	}

	if config.RetryAttempts != 3 {
		t.Errorf("expected RetryAttempts 3, got %d", config.RetryAttempts)
	}

	if config.RetryDelay != 1000 {
		t.Errorf("expected RetryDelay 1000ms, got %d", config.RetryDelay)
	}
}

func TestLoadErrorConfig(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}
	defer cleanup(t, fm)

	// Reset viper for clean test
	viper.Reset()
	fm.setErrorConfigDefaults()

	// Test loading default configuration
	config, err := fm.LoadErrorConfig()
	if err != nil {
		t.Fatalf("LoadErrorConfig failed: %v", err)
	}

	if config.LogLevel != "info" {
		t.Errorf("expected default LogLevel 'info', got %q", config.LogLevel)
	}

	// Test loading custom configuration
	viper.Set("error.verbose_errors", true)
	viper.Set("error.debug_mode", true)
	viper.Set("error.log_level", "debug")
	viper.Set("error.retry_attempts", 5)

	config, err = fm.LoadErrorConfig()
	if err != nil {
		t.Fatalf("LoadErrorConfig with custom values failed: %v", err)
	}

	if !config.VerboseErrors {
		t.Errorf("expected VerboseErrors true, got %v", config.VerboseErrors)
	}

	if !config.DebugMode {
		t.Errorf("expected DebugMode true, got %v", config.DebugMode)
	}

	if config.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got %q", config.LogLevel)
	}

	if config.RetryAttempts != 5 {
		t.Errorf("expected RetryAttempts 5, got %d", config.RetryAttempts)
	}
}

func TestValidateErrorConfig(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}
	defer cleanup(t, fm)

	tests := []struct {
		name        string
		config      *ErrorConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default config",
			config:      fm.DefaultErrorConfig(),
			expectError: false,
		},
		{
			name: "invalid log level",
			config: &ErrorConfig{
				LogLevel:      "invalid",
				MaxLogSize:    10 * 1024 * 1024,
				MaxLogFiles:   5,
				RetryAttempts: 3,
				RetryDelay:    1000,
			},
			expectError: true,
			errorMsg:    "invalid log level",
		},
		{
			name: "negative max log size",
			config: &ErrorConfig{
				LogLevel:      "info",
				MaxLogSize:    -1,
				MaxLogFiles:   5,
				RetryAttempts: 3,
				RetryDelay:    1000,
			},
			expectError: true,
			errorMsg:    "max_log_size_mb must be positive",
		},
		{
			name: "zero max log files",
			config: &ErrorConfig{
				LogLevel:      "info",
				MaxLogSize:    10 * 1024 * 1024,
				MaxLogFiles:   0,
				RetryAttempts: 3,
				RetryDelay:    1000,
			},
			expectError: true,
			errorMsg:    "max_log_files must be positive",
		},
		{
			name: "negative retry attempts",
			config: &ErrorConfig{
				LogLevel:      "info",
				MaxLogSize:    10 * 1024 * 1024,
				MaxLogFiles:   5,
				RetryAttempts: -1,
				RetryDelay:    1000,
			},
			expectError: true,
			errorMsg:    "retry_attempts must be non-negative",
		},
		{
			name: "negative retry delay",
			config: &ErrorConfig{
				LogLevel:      "info",
				MaxLogSize:    10 * 1024 * 1024,
				MaxLogFiles:   5,
				RetryAttempts: 3,
				RetryDelay:    -1,
			},
			expectError: true,
			errorMsg:    "retry_delay_ms must be non-negative",
		},
		{
			name: "valid custom config",
			config: &ErrorConfig{
				VerboseErrors: true,
				DebugMode:     true,
				LogLevel:      "debug",
				LogFile:       "app.log",
				MaxLogSize:    50 * 1024 * 1024,
				MaxLogFiles:   10,
				EnableColor:   false,
				RetryAttempts: 5,
				RetryDelay:    2000,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fm.ValidateErrorConfig(tt.config)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestIsValidLogLevel(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}
	defer cleanup(t, fm)

	validLevels := []string{"debug", "info", "warn", "error"}

	tests := []struct {
		level    string
		expected bool
	}{
		{"debug", true},
		{"info", true},
		{"warn", true},
		{"error", true},
		{"DEBUG", true}, // Case insensitive
		{"INFO", true},
		{"invalid", false},
		{"trace", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			result := fm.isValidLogLevel(tt.level, validLevels)
			if result != tt.expected {
				t.Errorf("isValidLogLevel(%q) = %v, expected %v", tt.level, result, tt.expected)
			}
		})
	}
}

func TestGetErrorLogPath(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}
	defer cleanup(t, fm)

	tests := []struct {
		name     string
		config   *ErrorConfig
		expected string
	}{
		{
			name: "empty log file",
			config: &ErrorConfig{
				LogFile: "",
			},
			expected: "",
		},
		{
			name: "relative log file",
			config: &ErrorConfig{
				LogFile: "app.log",
			},
			expected: filepath.Join(fm.DataPath, "app.log"),
		},
		{
			name: "absolute log file",
			config: &ErrorConfig{
				LogFile: "/tmp/app.log",
			},
			expected: "/tmp/app.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.GetErrorLogPath(tt.config)
			if result != tt.expected {
				t.Errorf("GetErrorLogPath() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestUpdateErrorConfig(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}
	defer cleanup(t, fm)

	// Initialize viper and create config file
	viper.Reset()
	if err := fm.InitializeViper(); err != nil {
		t.Fatalf("InitializeViper failed: %v", err)
	}

	// Test valid updates
	updates := map[string]interface{}{
		"verbose_errors": true,
		"debug_mode":     true,
		"log_level":      "debug",
		"retry_attempts": 5,
	}

	err = fm.UpdateErrorConfig(updates)
	if err != nil {
		t.Fatalf("UpdateErrorConfig failed: %v", err)
	}

	// Verify updates were applied
	config, err := fm.LoadErrorConfig()
	if err != nil {
		t.Fatalf("LoadErrorConfig after update failed: %v", err)
	}

	if !config.VerboseErrors {
		t.Errorf("expected VerboseErrors true after update, got %v", config.VerboseErrors)
	}

	if !config.DebugMode {
		t.Errorf("expected DebugMode true after update, got %v", config.DebugMode)
	}

	if config.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug' after update, got %q", config.LogLevel)
	}

	if config.RetryAttempts != 5 {
		t.Errorf("expected RetryAttempts 5 after update, got %d", config.RetryAttempts)
	}

	// Test invalid updates
	invalidUpdates := map[string]interface{}{
		"log_level": "invalid",
	}

	err = fm.UpdateErrorConfig(invalidUpdates)
	if err == nil {
		t.Errorf("expected error for invalid log level update")
	}
}

func TestSetErrorConfigDefaults(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}
	defer cleanup(t, fm)

	// Reset viper
	viper.Reset()

	// Set defaults
	fm.setErrorConfigDefaults()

	// Check that defaults are set
	if !viper.IsSet("error.verbose_errors") {
		t.Error("error.verbose_errors default not set")
	}

	if !viper.IsSet("error.debug_mode") {
		t.Error("error.debug_mode default not set")
	}

	if !viper.IsSet("error.log_level") {
		t.Error("error.log_level default not set")
	}

	if viper.GetString("error.log_level") != "info" {
		t.Errorf("expected default log_level 'info', got %q", viper.GetString("error.log_level"))
	}

	if viper.GetInt("error.retry_attempts") != 3 {
		t.Errorf("expected default retry_attempts 3, got %d", viper.GetInt("error.retry_attempts"))
	}
}

func TestInitializeViperWithErrorDefaults(t *testing.T) {
	fm, err := NewFileManager("test-app")
	if err != nil {
		t.Fatalf("NewFileManager failed: %v", err)
	}
	defer cleanup(t, fm)

	// Reset viper
	viper.Reset()

	err = fm.InitializeViper()
	if err != nil {
		t.Fatalf("InitializeViper failed: %v", err)
	}

	// Check that error defaults are set
	if !viper.IsSet("error.verbose_errors") {
		t.Error("error.verbose_errors default not set after InitializeViper")
	}

	if !viper.IsSet("error.log_level") {
		t.Error("error.log_level default not set after InitializeViper")
	}

	// Load error config to ensure it works
	config, err := fm.LoadErrorConfig()
	if err != nil {
		t.Fatalf("LoadErrorConfig after InitializeViper failed: %v", err)
	}

	if config.LogLevel != "info" {
		t.Errorf("expected default LogLevel 'info', got %q", config.LogLevel)
	}
}

// Helper functions for tests

func cleanup(t *testing.T, fm *FileManager) {
	if err := os.RemoveAll(fm.ConfigPath); err != nil {
		t.Errorf("Failed to remove config path: %v", err)
	}
	if err := os.RemoveAll(fm.DataPath); err != nil {
		t.Errorf("Failed to remove data path: %v", err)
	}
	viper.Reset()
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}