package logging

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestDefaultLoggingConfig(t *testing.T) {
	config := DefaultLoggingConfig()

	if config.VerboseErrors {
		t.Error("DefaultLoggingConfig().VerboseErrors should be false")
	}
	if config.DebugMode {
		t.Error("DefaultLoggingConfig().DebugMode should be false")
	}
	if config.LogLevel != "info" {
		t.Errorf("DefaultLoggingConfig().LogLevel = %q, want %q", config.LogLevel, "info")
	}
	if config.LogFile != "" {
		t.Errorf("DefaultLoggingConfig().LogFile = %q, want empty string", config.LogFile)
	}
	if config.MaxLogSize != 10*1024*1024 {
		t.Errorf("DefaultLoggingConfig().MaxLogSize = %d, want %d", config.MaxLogSize, 10*1024*1024)
	}
	if config.MaxLogFiles != 5 {
		t.Errorf("DefaultLoggingConfig().MaxLogFiles = %d, want %d", config.MaxLogFiles, 5)
	}
	if !config.EnableColor {
		t.Error("DefaultLoggingConfig().EnableColor should be true")
	}
	if config.TimeFormat != "2006-01-02T15:04:05Z07:00" {
		t.Errorf("DefaultLoggingConfig().TimeFormat = %q, want RFC3339", config.TimeFormat)
	}
}

func TestLoggingConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *LoggingConfig
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultLoggingConfig(),
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: &LoggingConfig{
				LogLevel:    "invalid",
				MaxLogSize:  1024,
				MaxLogFiles: 5,
			},
			wantErr: true,
		},
		{
			name: "invalid max log size",
			config: &LoggingConfig{
				LogLevel:    "info",
				MaxLogSize:  0,
				MaxLogFiles: 5,
			},
			wantErr: true,
		},
		{
			name: "invalid max log files",
			config: &LoggingConfig{
				LogLevel:    "info",
				MaxLogSize:  1024,
				MaxLogFiles: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoggingConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoggingConfig_ToLoggerConfig(t *testing.T) {
	dataPath := "/tmp/test"

	tests := []struct {
		name          string
		config        *LoggingConfig
		expectedLevel LogLevel
		expectedFile  string
		wantErr       bool
	}{
		{
			name:          "default config",
			config:        DefaultLoggingConfig(),
			expectedLevel: LogLevelInfo,
			expectedFile:  "",
			wantErr:       false,
		},
		{
			name: "debug level with relative file",
			config: &LoggingConfig{
				LogLevel:    "debug",
				LogFile:     "app.log",
				MaxLogSize:  1024,
				MaxLogFiles: 3,
				EnableColor: false,
				TimeFormat:  "2006-01-02 15:04:05",
			},
			expectedLevel: LogLevelDebug,
			expectedFile:  filepath.Join(dataPath, "app.log"),
			wantErr:       false,
		},
		{
			name: "absolute file path",
			config: &LoggingConfig{
				LogLevel:    "warn",
				LogFile:     "/var/log/app.log",
				MaxLogSize:  2048,
				MaxLogFiles: 10,
				EnableColor: true,
				TimeFormat:  "2006-01-02T15:04:05Z07:00",
			},
			expectedLevel: LogLevelWarn,
			expectedFile:  "/var/log/app.log",
			wantErr:       false,
		},
		{
			name: "invalid log level",
			config: &LoggingConfig{
				LogLevel:    "invalid",
				MaxLogSize:  1024,
				MaxLogFiles: 5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerConfig, err := tt.config.ToLoggerConfig(dataPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoggingConfig.ToLoggerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if loggerConfig.Level != tt.expectedLevel {
					t.Errorf("LoggerConfig.Level = %v, want %v", loggerConfig.Level, tt.expectedLevel)
				}
				if loggerConfig.OutputFile != tt.expectedFile {
					t.Errorf("LoggerConfig.OutputFile = %q, want %q", loggerConfig.OutputFile, tt.expectedFile)
				}
				if loggerConfig.MaxSize != tt.config.MaxLogSize {
					t.Errorf("LoggerConfig.MaxSize = %d, want %d", loggerConfig.MaxSize, tt.config.MaxLogSize)
				}
				if loggerConfig.MaxFiles != tt.config.MaxLogFiles {
					t.Errorf("LoggerConfig.MaxFiles = %d, want %d", loggerConfig.MaxFiles, tt.config.MaxLogFiles)
				}
			}
		})
	}
}

func TestSetViperDefaults(t *testing.T) {
	// Clear any existing configuration
	viper.Reset()

	// Set defaults
	SetViperDefaults()

	// Test that defaults are set correctly
	if !viper.IsSet("logging.verbose_errors") {
		t.Error("logging.verbose_errors default not set")
	}
	if viper.GetBool("logging.verbose_errors") {
		t.Error("logging.verbose_errors should default to false")
	}

	if !viper.IsSet("logging.debug_mode") {
		t.Error("logging.debug_mode default not set")
	}
	if viper.GetBool("logging.debug_mode") {
		t.Error("logging.debug_mode should default to false")
	}

	if !viper.IsSet("logging.log_level") {
		t.Error("logging.log_level default not set")
	}
	if viper.GetString("logging.log_level") != "info" {
		t.Errorf("logging.log_level should default to 'info', got %q", viper.GetString("logging.log_level"))
	}

	if !viper.IsSet("logging.max_log_size_mb") {
		t.Error("logging.max_log_size_mb default not set")
	}
	if viper.GetInt64("logging.max_log_size_mb") != 10*1024*1024 {
		t.Errorf("logging.max_log_size_mb should default to %d, got %d", 10*1024*1024, viper.GetInt64("logging.max_log_size_mb"))
	}
}

func TestLoadLoggingConfig(t *testing.T) {
	// Test with default configuration
	viper.Reset()
	SetViperDefaults()

	config, err := LoadLoggingConfig()
	if err != nil {
		t.Fatalf("LoadLoggingConfig() error = %v", err)
	}

	expected := DefaultLoggingConfig()
	if config.LogLevel != expected.LogLevel {
		t.Errorf("LoadLoggingConfig().LogLevel = %q, want %q", config.LogLevel, expected.LogLevel)
	}
	if config.MaxLogSize != expected.MaxLogSize {
		t.Errorf("LoadLoggingConfig().MaxLogSize = %d, want %d", config.MaxLogSize, expected.MaxLogSize)
	}
}

func TestUpdateLogLevel(t *testing.T) {
	// Save original global logger
	originalLogger := GetGlobalLogger()
	defer SetGlobalLogger(originalLogger)

	// Set up a structured logger
	config := &LoggerConfig{Level: LogLevelInfo}
	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	SetGlobalLogger(logger)

	tests := []struct {
		name     string
		level    string
		expected LogLevel
		wantErr  bool
	}{
		{"valid debug", "debug", LogLevelDebug, false},
		{"valid info", "info", LogLevelInfo, false},
		{"valid warn", "warn", LogLevelWarn, false},
		{"valid error", "error", LogLevelError, false},
		{"invalid level", "invalid", LogLevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateLogLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateLogLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				logger := GetGlobalLogger()
				if logger.GetLevel() != tt.expected {
					t.Errorf("After UpdateLogLevel(%q), level = %v, want %v", tt.level, logger.GetLevel(), tt.expected)
				}
			}
		})
	}
}

func TestEnableVerboseMode(t *testing.T) {
	// Save original global logger
	originalLogger := GetGlobalLogger()
	defer SetGlobalLogger(originalLogger)

	// Test with logger at debug level (should not change)
	config1 := &LoggerConfig{Level: LogLevelDebug}
	logger1, err := NewStructuredLogger(config1)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger1.Close()

	SetGlobalLogger(logger1)
	EnableVerboseMode()

	if GetGlobalLogger().GetLevel() != LogLevelDebug {
		t.Errorf("EnableVerboseMode() should not change debug level, got %v", GetGlobalLogger().GetLevel())
	}

	// Test with logger at error level (should change to info)
	config2 := &LoggerConfig{Level: LogLevelError}
	logger2, err := NewStructuredLogger(config2)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger2.Close()

	SetGlobalLogger(logger2)
	EnableVerboseMode()

	if GetGlobalLogger().GetLevel() != LogLevelInfo {
		t.Errorf("EnableVerboseMode() should set level to info when above info, got %v", GetGlobalLogger().GetLevel())
	}
}

func TestEnableDebugMode(t *testing.T) {
	// Save original global logger
	originalLogger := GetGlobalLogger()
	defer SetGlobalLogger(originalLogger)

	config := &LoggerConfig{Level: LogLevelError}
	logger1, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger1.Close()

	SetGlobalLogger(logger1)
	EnableDebugMode()

	if GetGlobalLogger().GetLevel() != LogLevelDebug {
		t.Errorf("EnableDebugMode() should set level to debug, got %v", GetGlobalLogger().GetLevel())
	}
}
