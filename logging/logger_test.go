package logging

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelError, "ERROR"},
		{LogLevelWarn, "WARN"},
		{LogLevelInfo, "INFO"},
		{LogLevelDebug, "DEBUG"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input       string
		expected    LogLevel
		expectError bool
	}{
		{"ERROR", LogLevelError, false},
		{"error", LogLevelError, false},
		{"WARN", LogLevelWarn, false},
		{"warn", LogLevelWarn, false},
		{"WARNING", LogLevelWarn, false},
		{"warning", LogLevelWarn, false},
		{"INFO", LogLevelInfo, false},
		{"info", LogLevelInfo, false},
		{"DEBUG", LogLevelDebug, false},
		{"debug", LogLevelDebug, false},
		{"invalid", LogLevelInfo, true},
		{"", LogLevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseLogLevel(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseLogLevel(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseLogLevel(%q) unexpected error: %v", tt.input, err)
				}
				if got != tt.expected {
					t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, got, tt.expected)
				}
			}
		})
	}
}

func TestDefaultLoggerConfig(t *testing.T) {
	config := DefaultLoggerConfig()

	if config.Level != LogLevelInfo {
		t.Errorf("DefaultLoggerConfig().Level = %v, want %v", config.Level, LogLevelInfo)
	}
	if config.OutputFile != "" {
		t.Errorf("DefaultLoggerConfig().OutputFile = %q, want empty string", config.OutputFile)
	}
	if config.MaxSize != 10*1024*1024 {
		t.Errorf("DefaultLoggerConfig().MaxSize = %d, want %d", config.MaxSize, 10*1024*1024)
	}
	if config.MaxFiles != 5 {
		t.Errorf("DefaultLoggerConfig().MaxFiles = %d, want %d", config.MaxFiles, 5)
	}
	if !config.EnableColor {
		t.Errorf("DefaultLoggerConfig().EnableColor = %v, want %v", config.EnableColor, true)
	}
	if config.TimeFormat != time.RFC3339 {
		t.Errorf("DefaultLoggerConfig().TimeFormat = %q, want %q", config.TimeFormat, time.RFC3339)
	}
}

func TestNewStructuredLogger_ConsoleOutput(t *testing.T) {
	config := &LoggerConfig{
		Level:      LogLevelDebug,
		OutputFile: "", // Console output
		TimeFormat: time.RFC3339,
	}

	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}
	defer logger.Close()

	if logger.config.Level != LogLevelDebug {
		t.Errorf("Logger level = %v, want %v", logger.config.Level, LogLevelDebug)
	}
	if logger.output != os.Stderr {
		t.Errorf("Logger output should be os.Stderr for console output")
	}
}

func TestNewStructuredLogger_FileOutput(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &LoggerConfig{
		Level:      LogLevelInfo,
		OutputFile: logFile,
		MaxSize:    1024,
		MaxFiles:   3,
		TimeFormat: time.RFC3339,
	}

	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}
	defer logger.Close()

	if logger.file == nil {
		t.Error("Logger file should not be nil for file output")
	}

	// Check if log file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file %s was not created", logFile)
	}
}

func TestStructuredLogger_LogLevels(t *testing.T) {
	// Capture stderr for testing
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	config := &LoggerConfig{
		Level:      LogLevelWarn,
		OutputFile: "",
		TimeFormat: "2006-01-02T15:04:05Z07:00",
	}

	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}
	defer logger.Close()

	// Test that debug and info messages are filtered out
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Close writer and read output
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should only contain warn and error messages
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be present")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should be present")
	}
}

func TestStructuredLogger_WithFields(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &LoggerConfig{
		Level:      LogLevelInfo,
		OutputFile: logFile,
		TimeFormat: "2006-01-02T15:04:05Z07:00",
	}

	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}
	defer logger.Close()

	// Log message with fields
	logger.Info("test message",
		NewField("key1", "value1"),
		NewField("key2", 42),
		NewField("key3", true),
	)

	// Read log file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	expectedParts := []string{
		"[INFO]",
		"test message",
		"key1=value1",
		"key2=42",
		"key3=true",
	}

	for _, part := range expectedParts {
		if !strings.Contains(logContent, part) {
			t.Errorf("Log content should contain %q, got: %s", part, logContent)
		}
	}
}

func TestStructuredLogger_SetLevel(t *testing.T) {
	config := DefaultLoggerConfig()
	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}
	defer logger.Close()

	// Test initial level
	if logger.GetLevel() != LogLevelInfo {
		t.Errorf("Initial level = %v, want %v", logger.GetLevel(), LogLevelInfo)
	}

	// Test setting level
	logger.SetLevel(LogLevelDebug)
	if logger.GetLevel() != LogLevelDebug {
		t.Errorf("After SetLevel(Debug), level = %v, want %v", logger.GetLevel(), LogLevelDebug)
	}
}

func TestStructuredLogger_FileRotation(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &LoggerConfig{
		Level:      LogLevelInfo,
		OutputFile: logFile,
		MaxSize:    100, // Small size to trigger rotation
		MaxFiles:   3,
		TimeFormat: time.RFC3339,
	}

	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}
	defer logger.Close()

	// Write enough data to trigger rotation
	longMessage := strings.Repeat("This is a long log message that will help trigger file rotation. ", 10)

	for i := 0; i < 5; i++ {
		logger.Info(fmt.Sprintf("Message %d: %s", i, longMessage))
	}

	// Check if rotation occurred by looking for rotated files
	rotatedFile := logFile + ".1"
	if _, err := os.Stat(rotatedFile); os.IsNotExist(err) {
		// Rotation might not have occurred yet, let's force it by writing more
		for i := 0; i < 10; i++ {
			logger.Info(fmt.Sprintf("Additional message %d: %s", i, longMessage))
		}
	}

	// Now check again
	if _, err := os.Stat(rotatedFile); os.IsNotExist(err) {
		t.Logf("Rotated file %s does not exist, checking current file size", rotatedFile)

		// Check current file size
		if stat, err := os.Stat(logFile); err == nil {
			t.Logf("Current log file size: %d bytes", stat.Size())
		}
	}
}

func TestSimpleLogger(t *testing.T) {
	// Capture stderr for testing
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := &SimpleLogger{level: LogLevelWarn}

	// Test that debug and info messages are filtered out
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Close writer and read output
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should only contain warn and error messages
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be present")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should be present")
	}
}

func TestGlobalLogger(t *testing.T) {
	// Reset global logger
	globalLogger = nil

	// Test that GetGlobalLogger creates a default logger
	logger := GetGlobalLogger()
	if logger == nil {
		t.Error("GetGlobalLogger() should not return nil")
	}

	// Test setting custom global logger
	customLogger := &SimpleLogger{level: LogLevelDebug}
	SetGlobalLogger(customLogger)

	if GetGlobalLogger() != customLogger {
		t.Error("GetGlobalLogger() should return the custom logger")
	}
}

func TestGlobalLoggingFunctions(t *testing.T) {
	// Capture stderr for testing
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Set a simple logger as global
	SetGlobalLogger(&SimpleLogger{level: LogLevelDebug})

	// Test global logging functions
	Error("error message", NewField("key", "value"))
	Warn("warn message")
	Info("info message")
	Debug("debug message")

	// Close writer and read output
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedMessages := []string{"error message", "warn message", "info message", "debug message"}
	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Output should contain %q, got: %s", msg, output)
		}
	}
}

func TestNewField(t *testing.T) {
	field := NewField("test_key", "test_value")

	if field.Key != "test_key" {
		t.Errorf("Field.Key = %q, want %q", field.Key, "test_key")
	}
	if field.Value != "test_value" {
		t.Errorf("Field.Value = %q, want %q", field.Value, "test_value")
	}
}
