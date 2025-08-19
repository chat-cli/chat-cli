package logging

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

// LoggingConfig represents logging configuration options
type LoggingConfig struct {
	VerboseErrors bool   `yaml:"verbose_errors" mapstructure:"verbose_errors"`
	DebugMode     bool   `yaml:"debug_mode" mapstructure:"debug_mode"`
	LogLevel      string `yaml:"log_level" mapstructure:"log_level"`
	LogFile       string `yaml:"log_file" mapstructure:"log_file"`
	MaxLogSize    int64  `yaml:"max_log_size_mb" mapstructure:"max_log_size_mb"`
	MaxLogFiles   int    `yaml:"max_log_files" mapstructure:"max_log_files"`
	EnableColor   bool   `yaml:"enable_color" mapstructure:"enable_color"`
	TimeFormat    string `yaml:"time_format" mapstructure:"time_format"`
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		VerboseErrors: false,
		DebugMode:     false,
		LogLevel:      "info",
		LogFile:       "",               // Empty means console output
		MaxLogSize:    10 * 1024 * 1024, // 10MB
		MaxLogFiles:   5,
		EnableColor:   true,
		TimeFormat:    "2006-01-02T15:04:05Z07:00", // RFC3339
	}
}

// LoadLoggingConfig loads logging configuration from Viper
func LoadLoggingConfig() (*LoggingConfig, error) {
	config := DefaultLoggingConfig()

	// Load from Viper configuration
	if viper.IsSet("logging") {
		if err := viper.UnmarshalKey("logging", config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal logging config: %w", err)
		}
	}

	// Validate and normalize configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logging configuration: %w", err)
	}

	return config, nil
}

// Validate validates the logging configuration
func (c *LoggingConfig) Validate() error {
	// Validate log level
	if _, err := ParseLogLevel(c.LogLevel); err != nil {
		return fmt.Errorf("invalid log level %q: %w", c.LogLevel, err)
	}

	// Validate max log size
	if c.MaxLogSize <= 0 {
		return fmt.Errorf("max_log_size_mb must be positive, got %d", c.MaxLogSize)
	}

	// Validate max log files
	if c.MaxLogFiles <= 0 {
		return fmt.Errorf("max_log_files must be positive, got %d", c.MaxLogFiles)
	}

	return nil
}

// ToLoggerConfig converts LoggingConfig to LoggerConfig
func (c *LoggingConfig) ToLoggerConfig(dataPath string) (*LoggerConfig, error) {
	level, err := ParseLogLevel(c.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	config := &LoggerConfig{
		Level:       level,
		OutputFile:  c.LogFile,
		MaxSize:     c.MaxLogSize,
		MaxFiles:    c.MaxLogFiles,
		EnableColor: c.EnableColor,
		TimeFormat:  c.TimeFormat,
	}

	// If log file is specified but not absolute, make it relative to data path
	if c.LogFile != "" && !filepath.IsAbs(c.LogFile) {
		config.OutputFile = filepath.Join(dataPath, c.LogFile)
	}

	return config, nil
}

// SetViperDefaults sets default values in Viper for logging configuration
func SetViperDefaults() {
	defaults := DefaultLoggingConfig()

	viper.SetDefault("logging.verbose_errors", defaults.VerboseErrors)
	viper.SetDefault("logging.debug_mode", defaults.DebugMode)
	viper.SetDefault("logging.log_level", defaults.LogLevel)
	viper.SetDefault("logging.log_file", defaults.LogFile)
	viper.SetDefault("logging.max_log_size_mb", defaults.MaxLogSize)
	viper.SetDefault("logging.max_log_files", defaults.MaxLogFiles)
	viper.SetDefault("logging.enable_color", defaults.EnableColor)
	viper.SetDefault("logging.time_format", defaults.TimeFormat)
}

// InitializeLogging initializes the global logger based on configuration
func InitializeLogging(dataPath string) error {
	// Load logging configuration
	loggingConfig, err := LoadLoggingConfig()
	if err != nil {
		return fmt.Errorf("failed to load logging config: %w", err)
	}

	// Convert to logger configuration
	loggerConfig, err := loggingConfig.ToLoggerConfig(dataPath)
	if err != nil {
		return fmt.Errorf("failed to create logger config: %w", err)
	}

	// Create and set global logger
	logger, err := NewStructuredLogger(loggerConfig)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	SetGlobalLogger(logger)
	return nil
}

// UpdateLogLevel updates the log level of the global logger
func UpdateLogLevel(level string) error {
	logLevel, err := ParseLogLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}

	logger := GetGlobalLogger()
	logger.SetLevel(logLevel)
	return nil
}

// EnableVerboseMode enables verbose error reporting
func EnableVerboseMode() {
	logger := GetGlobalLogger()
	if logger.GetLevel() < LogLevelInfo {
		logger.SetLevel(LogLevelInfo)
	}
}

// EnableDebugMode enables debug logging
func EnableDebugMode() {
	logger := GetGlobalLogger()
	logger.SetLevel(LogLevelDebug)
}
