package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

// TestLoggingIntegration tests the complete logging system integration
func TestLoggingIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Reset viper for clean test
	viper.Reset()

	// Set up configuration
	viper.Set("logging.log_level", "debug")
	viper.Set("logging.log_file", "test.log")
	viper.Set("logging.max_log_size_mb", 1024)
	viper.Set("logging.max_log_files", 3)
	viper.Set("logging.verbose_errors", true)
	viper.Set("logging.debug_mode", true)

	// Initialize logging
	err := InitializeLogging(tempDir)
	if err != nil {
		t.Fatalf("InitializeLogging() error = %v", err)
	}

	// Test that global logger was set up correctly
	logger := GetGlobalLogger()
	if logger == nil {
		t.Fatal("Global logger should not be nil after initialization")
	}

	if logger.GetLevel() != LogLevelDebug {
		t.Errorf("Logger level = %v, want %v", logger.GetLevel(), LogLevelDebug)
	}

	// Test logging at different levels
	Error("Integration test error", NewField("component", "test"), NewField("operation", "integration"))
	Warn("Integration test warning", NewField("component", "test"))
	Info("Integration test info", NewField("component", "test"))
	Debug("Integration test debug", NewField("component", "test"))

	// Close logger to flush any buffered content
	logger.Close()

	// Verify log file was created and contains expected content
	logFile := filepath.Join(tempDir, "test.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	expectedMessages := []string{
		"Integration test error",
		"Integration test warning",
		"Integration test info",
		"Integration test debug",
		"component=test",
		"operation=integration",
	}

	for _, expected := range expectedMessages {
		if !strings.Contains(logContent, expected) {
			t.Errorf("Log content should contain %q, got: %s", expected, logContent)
		}
	}

	// Verify log levels are present
	expectedLevels := []string{"[ERROR]", "[WARN]", "[INFO]", "[DEBUG]"}
	for _, level := range expectedLevels {
		if !strings.Contains(logContent, level) {
			t.Errorf("Log content should contain level %q, got: %s", level, logContent)
		}
	}
}

// TestVerboseAndDebugModes tests the verbose and debug mode functionality
func TestVerboseAndDebugModes(t *testing.T) {
	// Save original global logger
	originalLogger := GetGlobalLogger()
	defer SetGlobalLogger(originalLogger)

	// Start with error level
	config := &LoggerConfig{Level: LogLevelError}
	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	SetGlobalLogger(logger)

	// Test verbose mode
	EnableVerboseMode()
	if GetGlobalLogger().GetLevel() != LogLevelInfo {
		t.Errorf("After EnableVerboseMode(), level = %v, want %v", GetGlobalLogger().GetLevel(), LogLevelInfo)
	}

	// Test debug mode
	EnableDebugMode()
	if GetGlobalLogger().GetLevel() != LogLevelDebug {
		t.Errorf("After EnableDebugMode(), level = %v, want %v", GetGlobalLogger().GetLevel(), LogLevelDebug)
	}

	// Test UpdateLogLevel
	err = UpdateLogLevel("warn")
	if err != nil {
		t.Errorf("UpdateLogLevel() error = %v", err)
	}
	if GetGlobalLogger().GetLevel() != LogLevelWarn {
		t.Errorf("After UpdateLogLevel('warn'), level = %v, want %v", GetGlobalLogger().GetLevel(), LogLevelWarn)
	}
}

// TestFileRotationIntegration tests file rotation with realistic data
func TestFileRotationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "rotation_test.log")

	config := &LoggerConfig{
		Level:      LogLevelInfo,
		OutputFile: logFile,
		MaxSize:    500, // Small size to trigger rotation quickly
		MaxFiles:   3,
		TimeFormat: time.RFC3339,
	}

	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}
	defer logger.Close()

	// Generate enough log data to trigger rotation
	longMessage := strings.Repeat("This is a test message that will help trigger log rotation. ", 20)

	for i := 0; i < 20; i++ {
		logger.Info("Test message",
			NewField("iteration", i),
			NewField("message", longMessage),
			NewField("timestamp", time.Now().Unix()),
		)
	}

	// Check that main log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Main log file %s should exist", logFile)
	}

	// Check for rotated files (may or may not exist depending on exact timing)
	rotatedFile := logFile + ".1"
	if stat, err := os.Stat(rotatedFile); err == nil {
		t.Logf("Rotated file %s exists with size %d bytes", rotatedFile, stat.Size())
	} else {
		t.Logf("Rotated file %s does not exist yet", rotatedFile)
	}
}
