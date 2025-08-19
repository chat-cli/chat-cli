package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case LogLevelError:
		return "ERROR"
	case LogLevelWarn:
		return "WARN"
	case LogLevelInfo:
		return "INFO"
	case LogLevelDebug:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) (LogLevel, error) {
	switch level {
	case "ERROR", "error":
		return LogLevelError, nil
	case "WARN", "warn", "WARNING", "warning":
		return LogLevelWarn, nil
	case "INFO", "info":
		return LogLevelInfo, nil
	case "DEBUG", "debug":
		return LogLevelDebug, nil
	default:
		return LogLevelInfo, fmt.Errorf("invalid log level: %s", level)
	}
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// Logger interface for structured logging
type Logger interface {
	Error(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	SetLevel(level LogLevel)
	GetLevel() LogLevel
	Close() error
}

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level       LogLevel
	OutputFile  string
	MaxSize     int64  // Maximum size in bytes before rotation
	MaxFiles    int    // Maximum number of log files to keep
	EnableColor bool   // Enable colored output for console
	TimeFormat  string // Time format for log entries
}

// DefaultLoggerConfig returns a default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:       LogLevelInfo,
		OutputFile:  "",
		MaxSize:     10 * 1024 * 1024, // 10MB
		MaxFiles:    5,
		EnableColor: true,
		TimeFormat:  time.RFC3339,
	}
}

// StructuredLogger is a structured logger implementation with file rotation
type StructuredLogger struct {
	config      *LoggerConfig
	output      io.Writer
	file        *os.File
	mutex       sync.RWMutex
	currentSize int64
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(config *LoggerConfig) (*StructuredLogger, error) {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	logger := &StructuredLogger{
		config: config,
	}

	if config.OutputFile != "" {
		if err := logger.setupFileOutput(); err != nil {
			return nil, fmt.Errorf("failed to setup file output: %w", err)
		}
	} else {
		logger.output = os.Stderr
	}

	return logger, nil
}

// setupFileOutput configures file-based logging with rotation
func (l *StructuredLogger) setupFileOutput() error {
	// Ensure directory exists
	dir := filepath.Dir(l.config.OutputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open or create log file
	file, err := os.OpenFile(l.config.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	l.file = file
	l.output = file
	l.currentSize = stat.Size()

	return nil
}

// Error logs an error message
func (l *StructuredLogger) Error(msg string, fields ...Field) {
	l.log(LogLevelError, msg, fields...)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(msg string, fields ...Field) {
	l.log(LogLevelWarn, msg, fields...)
}

// Info logs an info message
func (l *StructuredLogger) Info(msg string, fields ...Field) {
	l.log(LogLevelInfo, msg, fields...)
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(msg string, fields ...Field) {
	l.log(LogLevelDebug, msg, fields...)
}

// SetLevel sets the logging level
func (l *StructuredLogger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.config.Level = level
}

// GetLevel returns the current logging level
func (l *StructuredLogger) GetLevel() LogLevel {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.config.Level
}

// Close closes the logger and any open files
func (l *StructuredLogger) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		l.output = os.Stderr
		return err
	}
	return nil
}

// log writes a log entry with the specified level and fields
func (l *StructuredLogger) log(level LogLevel, msg string, fields ...Field) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Check if we should log at this level
	if level > l.config.Level {
		return
	}

	// Check if we need to rotate the log file
	if l.file != nil && l.currentSize > l.config.MaxSize {
		if err := l.rotateLogFile(); err != nil {
			// If rotation fails, continue logging to current file
			fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
		}
	}

	// Format the log entry
	entry := l.formatLogEntry(level, msg, fields...)

	// Write to output
	n, err := l.output.Write([]byte(entry))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log entry: %v\n", err)
		return
	}

	// Update current size if writing to file
	if l.file != nil {
		l.currentSize += int64(n)
	}
}

// formatLogEntry formats a log entry with timestamp, level, message, and fields
func (l *StructuredLogger) formatLogEntry(level LogLevel, msg string, fields ...Field) string {
	timestamp := time.Now().Format(l.config.TimeFormat)

	// Start with timestamp, level, and message
	entry := fmt.Sprintf("%s [%s] %s", timestamp, level.String(), msg)

	// Add structured fields
	if len(fields) > 0 {
		entry += " |"
		for _, field := range fields {
			entry += fmt.Sprintf(" %s=%v", field.Key, field.Value)
		}
	}

	entry += "\n"
	return entry
}

// rotateLogFile rotates the current log file
func (l *StructuredLogger) rotateLogFile() error {
	if l.file == nil {
		return nil
	}

	// Close current file
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close current log file: %w", err)
	}

	// Rotate existing files
	if err := l.rotateExistingFiles(); err != nil {
		return fmt.Errorf("failed to rotate existing files: %w", err)
	}

	// Create new log file
	file, err := os.OpenFile(l.config.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}

	l.file = file
	l.output = file
	l.currentSize = 0

	return nil
}

// rotateExistingFiles rotates existing log files (file.log -> file.log.1 -> file.log.2, etc.)
func (l *StructuredLogger) rotateExistingFiles() error {
	baseFile := l.config.OutputFile

	// Remove the oldest file if it exists
	oldestFile := fmt.Sprintf("%s.%d", baseFile, l.config.MaxFiles)
	if _, err := os.Stat(oldestFile); err == nil {
		if err := os.Remove(oldestFile); err != nil {
			return fmt.Errorf("failed to remove oldest log file: %w", err)
		}
	}

	// Rotate files from newest to oldest
	for i := l.config.MaxFiles - 1; i >= 1; i-- {
		oldName := fmt.Sprintf("%s.%d", baseFile, i)
		newName := fmt.Sprintf("%s.%d", baseFile, i+1)

		if _, err := os.Stat(oldName); err == nil {
			if err := os.Rename(oldName, newName); err != nil {
				return fmt.Errorf("failed to rotate log file %s to %s: %w", oldName, newName, err)
			}
		}
	}

	// Move current file to .1
	rotatedName := fmt.Sprintf("%s.1", baseFile)
	if err := os.Rename(baseFile, rotatedName); err != nil {
		return fmt.Errorf("failed to rotate current log file: %w", err)
	}

	return nil
}

// NewField creates a new logging field
func NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Global logger instance
var globalLogger Logger

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	if globalLogger == nil {
		// Create default logger if none exists
		config := DefaultLoggerConfig()
		logger, err := NewStructuredLogger(config)
		if err != nil {
			// Fallback to a simple logger if structured logger fails
			return &SimpleLogger{level: LogLevelInfo}
		}
		globalLogger = logger
	}
	return globalLogger
}

// Convenience functions for global logging
func Error(msg string, fields ...Field) {
	GetGlobalLogger().Error(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

func Info(msg string, fields ...Field) {
	GetGlobalLogger().Info(msg, fields...)
}

func Debug(msg string, fields ...Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

// SimpleLogger is a fallback logger that writes to stderr
type SimpleLogger struct {
	level LogLevel
	mutex sync.RWMutex
}

// Error logs an error message
func (s *SimpleLogger) Error(msg string, fields ...Field) {
	s.log(LogLevelError, msg, fields...)
}

// Warn logs a warning message
func (s *SimpleLogger) Warn(msg string, fields ...Field) {
	s.log(LogLevelWarn, msg, fields...)
}

// Info logs an info message
func (s *SimpleLogger) Info(msg string, fields ...Field) {
	s.log(LogLevelInfo, msg, fields...)
}

// Debug logs a debug message
func (s *SimpleLogger) Debug(msg string, fields ...Field) {
	s.log(LogLevelDebug, msg, fields...)
}

// SetLevel sets the logging level
func (s *SimpleLogger) SetLevel(level LogLevel) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.level = level
}

// GetLevel returns the current logging level
func (s *SimpleLogger) GetLevel() LogLevel {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.level
}

// Close is a no-op for SimpleLogger
func (s *SimpleLogger) Close() error {
	return nil
}

// log writes a simple log entry
func (s *SimpleLogger) log(level LogLevel, msg string, fields ...Field) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if level > s.level {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("%s [%s] %s", timestamp, level.String(), msg)

	if len(fields) > 0 {
		entry += " |"
		for _, field := range fields {
			entry += fmt.Sprintf(" %s=%v", field.Key, field.Value)
		}
	}

	fmt.Fprintln(os.Stderr, entry)
}
