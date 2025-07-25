package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

// FileManager handles OS-specific paths for configuration and data storage
type FileManager struct {
	AppName     string
	ConfigFile  string
	DBFile      string
	ConfigPath  string
	DataPath    string
	Environment string
}

// NewFileManager creates a new instance of FileManager with OS-specific paths
func NewFileManager(appName string) (*FileManager, error) {
	fm := &FileManager{
		AppName:     appName,
		ConfigFile:  "config.yaml",
		DBFile:      "data.db",
		Environment: os.Getenv("APP_ENV"),
	}

	// Set default environment if not specified
	if fm.Environment == "" {
		fm.Environment = "development"
	}

	// Initialize paths based on OS
	if err := fm.initializePaths(); err != nil {
		return nil, err
	}

	return fm, nil
}

// initializePaths sets up OS-specific paths for config and data storage
func (fm *FileManager) initializePaths() error {
	var configBase string
	var dataBase string

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		configBase = appData
		dataBase = appData

	case "darwin":
		home := os.Getenv("HOME")
		configBase = filepath.Join(home, "Library", "Application Support")
		dataBase = configBase

	default: // Linux and other Unix-like systems
		// Follow XDG Base Directory Specification
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			xdgConfig = filepath.Join(os.Getenv("HOME"), ".config")
		}
		xdgData := os.Getenv("XDG_DATA_HOME")
		if xdgData == "" {
			xdgData = filepath.Join(os.Getenv("HOME"), ".local", "share")
		}
		configBase = xdgConfig
		dataBase = xdgData
	}

	// Set final paths
	fm.ConfigPath = filepath.Join(configBase, fm.AppName)
	fm.DataPath = filepath.Join(dataBase, fm.AppName)

	// Create directories if they don't exist
	if err := os.MkdirAll(fm.ConfigPath, 0750); err != nil {
		return err
	}
	if err := os.MkdirAll(fm.DataPath, 0750); err != nil {
		return err
	}

	return nil
}

// InitializeViper sets up Viper with the correct config file path
func (fm *FileManager) InitializeViper() error {
	viper.SetConfigName(fm.ConfigFile[:len(fm.ConfigFile)-len(filepath.Ext(fm.ConfigFile))])
	viper.SetConfigType("yaml")
	viper.AddConfigPath(fm.ConfigPath)

	// Set some default configurations
	viper.SetDefault("environment", fm.Environment)
	viper.SetDefault("db_path", fm.GetDBPath())
	viper.SetDefault("db_driver", "sqlite3")

	// Create config file if it doesn't exist
	if err := fm.createDefaultConfig(); err != nil {
		return err
	}

	return viper.ReadInConfig()
}

// GetDBPath returns the full path to the SQLite database file
func (fm *FileManager) GetDBPath() string {
	return filepath.Join(fm.DataPath, fm.DBFile)
}

// GetDBDriver returns the database type from the config
func (fm *FileManager) GetDBDriver() string {
	return viper.GetString("db_driver")
}

// createDefaultConfig creates a default config file if it doesn't exist
func (fm *FileManager) createDefaultConfig() error {
	configPath := filepath.Join(fm.ConfigPath, fm.ConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return viper.SafeWriteConfig()
	}
	return nil
}

// GetConfigValue returns a configuration value with precedence order:
// 1. Feature flag (command line argument)
// 2. Configuration file
// 3. Default value
func (fm *FileManager) GetConfigValue(key string, flagValue interface{}, defaultValue interface{}) interface{} {
	// Check if flag value is provided and not empty/zero value
	switch v := flagValue.(type) {
	case string:
		if v != "" && v != defaultValue {
			return v
		}
	case int32:
		if v != 0 && v != defaultValue {
			return v
		}
	case float32:
		if v != 0.0 && v != defaultValue {
			return v
		}
	}

	// Check configuration file
	if viper.IsSet(key) {
		return viper.Get(key)
	}

	// Return default value
	return defaultValue
}
