package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// ErrorConfig represents error handling configuration options
type ErrorConfig struct {
	VerboseErrors bool   `yaml:"verbose_errors" mapstructure:"verbose_errors"`
	DebugMode     bool   `yaml:"debug_mode" mapstructure:"debug_mode"`
	LogLevel      string `yaml:"log_level" mapstructure:"log_level"`
	LogFile       string `yaml:"log_file" mapstructure:"log_file"`
	MaxLogSize    int64  `yaml:"max_log_size_mb" mapstructure:"max_log_size_mb"`
	MaxLogFiles   int    `yaml:"max_log_files" mapstructure:"max_log_files"`
	EnableColor   bool   `yaml:"enable_color" mapstructure:"enable_color"`
	TimeFormat    string `yaml:"time_format" mapstructure:"time_format"`
	RetryAttempts int    `yaml:"retry_attempts" mapstructure:"retry_attempts"`
	RetryDelay    int    `yaml:"retry_delay_ms" mapstructure:"retry_delay_ms"`
}

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
	viper.SetDefault("db_driver", "sqlite")

	// Set error handling defaults
	fm.setErrorConfigDefaults()

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
func (fm *FileManager) GetConfigValue(key string, flagValue, defaultValue interface{}) interface{} {
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

// DefaultErrorConfig returns default error handling configuration
func (fm *FileManager) DefaultErrorConfig() *ErrorConfig {
	return &ErrorConfig{
		VerboseErrors: false,
		DebugMode:     false,
		LogLevel:      "info",
		LogFile:       "",               // Empty means console output
		MaxLogSize:    10 * 1024 * 1024, // 10MB
		MaxLogFiles:   5,
		EnableColor:   true,
		TimeFormat:    "2006-01-02T15:04:05Z07:00", // RFC3339
		RetryAttempts: 3,
		RetryDelay:    1000, // 1 second in milliseconds
	}
}

// setErrorConfigDefaults sets default values in Viper for error configuration
func (fm *FileManager) setErrorConfigDefaults() {
	defaults := fm.DefaultErrorConfig()

	viper.SetDefault("error.verbose_errors", defaults.VerboseErrors)
	viper.SetDefault("error.debug_mode", defaults.DebugMode)
	viper.SetDefault("error.log_level", defaults.LogLevel)
	viper.SetDefault("error.log_file", defaults.LogFile)
	viper.SetDefault("error.max_log_size_mb", defaults.MaxLogSize)
	viper.SetDefault("error.max_log_files", defaults.MaxLogFiles)
	viper.SetDefault("error.enable_color", defaults.EnableColor)
	viper.SetDefault("error.time_format", defaults.TimeFormat)
	viper.SetDefault("error.retry_attempts", defaults.RetryAttempts)
	viper.SetDefault("error.retry_delay_ms", defaults.RetryDelay)
}

// LoadErrorConfig loads error handling configuration from Viper
func (fm *FileManager) LoadErrorConfig() (*ErrorConfig, error) {
	config := fm.DefaultErrorConfig()

	// Load from Viper configuration
	if viper.IsSet("error") {
		if err := viper.UnmarshalKey("error", config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error config: %w", err)
		}
	}

	// Validate configuration
	if err := fm.ValidateErrorConfig(config); err != nil {
		return nil, fmt.Errorf("invalid error configuration: %w", err)
	}

	return config, nil
}

// ValidateErrorConfig validates the error handling configuration
func (fm *FileManager) ValidateErrorConfig(config *ErrorConfig) error {
	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !fm.isValidLogLevel(config.LogLevel, validLevels) {
		return fmt.Errorf("invalid log level %q, must be one of: %s", 
			config.LogLevel, strings.Join(validLevels, ", "))
	}

	// Validate max log size
	if config.MaxLogSize <= 0 {
		return fmt.Errorf("max_log_size_mb must be positive, got %d", config.MaxLogSize)
	}

	// Validate max log files
	if config.MaxLogFiles <= 0 {
		return fmt.Errorf("max_log_files must be positive, got %d", config.MaxLogFiles)
	}

	// Validate retry attempts
	if config.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative, got %d", config.RetryAttempts)
	}

	// Validate retry delay
	if config.RetryDelay < 0 {
		return fmt.Errorf("retry_delay_ms must be non-negative, got %d", config.RetryDelay)
	}

	// Validate log file path if specified
	if config.LogFile != "" {
		logPath := config.LogFile
		if !filepath.IsAbs(logPath) {
			logPath = filepath.Join(fm.DataPath, logPath)
		}
		
		// Check if directory exists or can be created
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, 0750); err != nil {
			return fmt.Errorf("cannot create log directory %q: %w", logDir, err)
		}
	}

	return nil
}

// isValidLogLevel checks if the provided log level is valid
func (fm *FileManager) isValidLogLevel(level string, validLevels []string) bool {
	level = strings.ToLower(level)
	for _, valid := range validLevels {
		if level == valid {
			return true
		}
	}
	return false
}

// GetErrorLogPath returns the full path to the error log file
func (fm *FileManager) GetErrorLogPath(config *ErrorConfig) string {
	if config.LogFile == "" {
		return ""
	}
	
	if filepath.IsAbs(config.LogFile) {
		return config.LogFile
	}
	
	return filepath.Join(fm.DataPath, config.LogFile)
}

// UpdateErrorConfig updates error configuration values in Viper
func (fm *FileManager) UpdateErrorConfig(updates map[string]interface{}) error {
	// Apply updates to Viper
	for key, value := range updates {
		viper.Set(fmt.Sprintf("error.%s", key), value)
	}

	// Validate the updated configuration
	config, err := fm.LoadErrorConfig()
	if err != nil {
		return fmt.Errorf("invalid configuration after update: %w", err)
	}

	// If validation passes, save the configuration
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Store the validated config back (this ensures any normalization is applied)
	viper.Set("error.verbose_errors", config.VerboseErrors)
	viper.Set("error.debug_mode", config.DebugMode)
	viper.Set("error.log_level", config.LogLevel)
	viper.Set("error.log_file", config.LogFile)
	viper.Set("error.max_log_size_mb", config.MaxLogSize)
	viper.Set("error.max_log_files", config.MaxLogFiles)
	viper.Set("error.enable_color", config.EnableColor)
	viper.Set("error.time_format", config.TimeFormat)
	viper.Set("error.retry_attempts", config.RetryAttempts)
	viper.Set("error.retry_delay_ms", config.RetryDelay)

	return nil
}
