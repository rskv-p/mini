package log_core

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

//---------------------
// Log Levels
//---------------------

// Define basic log levels
const (
	InfoLevel  = "INFO"
	ErrorLevel = "ERROR"
	DebugLevel = "DEBUG"
	WarnLevel  = "WARN"
	TraceLevel = "TRACE"
)

//---------------------
// Log Configuration Struct
//---------------------

type Config struct {
	Name       string   // Logger name
	Level      string   // Log level (e.g., INFO, ERROR, etc.)
	Format     string   // Log format (e.g., console or json)
	Outputs    []string // Log outputs (stdout, stderr, or file paths)
	WithCaller bool     // Include caller info in logs
}

//---------------------
// Configuration Loading
//---------------------

// LoadConfigFromEnv loads the log configuration from environment variables.
func LoadConfigFromEnv() *Config {
	cfg := &Config{
		Name:       "mini",    // Default name
		Level:      "DEBUG",   // Default level
		Format:     "console", // Default format
		Outputs:    parseOutputPaths(),
		WithCaller: strings.ToLower(os.Getenv("MINI_LOG_CALLER")) == "true",
	}

	// Load the log level from environment variable
	if level := strings.ToUpper(os.Getenv("MINI_LOG_LEVEL")); level != "" {
		cfg.Level = level
	}

	// Load log format from environment variable
	if format := strings.ToUpper(os.Getenv("MINI_LOG_FORMAT")); format != "" {
		cfg.Format = format
	}

	return cfg
}

// parseOutputPaths parses the output paths (stdout, stderr, or file paths).
func parseOutputPaths() []string {
	var paths []string
	stream := strings.ToLower(os.Getenv("MINI_LOG_CONSOLE_STREAM"))
	if stream == "stdout" || stream == "stderr" {
		paths = append(paths, stream)
	}
	if file := os.Getenv("MINI_LOG_FILE"); file != "" {
		paths = append(paths, file)
	}
	if len(paths) == 0 {
		paths = append(paths, "stdout")
	}
	return paths
}

//---------------------
// Simple Logger Implementation
//---------------------

// SimpleLogger is a basic implementation of the Logger interface with additional context.
type SimpleLogger struct {
	level   string
	outputs []string
}

// NewLogger creates a new instance of SimpleLogger with the specified configuration.
func NewLogger(cfg *Config) *SimpleLogger {
	return &SimpleLogger{
		level:   cfg.Level,
		outputs: cfg.Outputs,
	}
}

// Log function logs a message based on the log level and additional context.
func (l *SimpleLogger) Log(level string, msg string, context map[string]interface{}) {
	if shouldLog(l.level, level) {
		logMessage := fmt.Sprintf("[%s] %s: %s", level, l.getCallerInfo(), msg)
		// Add context to the log message
		if len(context) > 0 {
			for key, value := range context {
				logMessage += fmt.Sprintf(" | %s=%v", key, value)
			}
		}

		fmt.Println(logMessage)
	}
}

// Info logs an info message with optional context.
func (l *SimpleLogger) Info(msg string, context map[string]interface{}) {
	l.Log(InfoLevel, msg, context)
}

// Error logs an error message with optional context.
func (l *SimpleLogger) Error(msg string, context map[string]interface{}) {
	l.Log(ErrorLevel, msg, context)
}

// Debug logs a debug message with optional context.
func (l *SimpleLogger) Debug(msg string, context map[string]interface{}) {
	l.Log(DebugLevel, msg, context)
}

// Warn logs a warn message with optional context.
func (l *SimpleLogger) Warn(msg string, context map[string]interface{}) {
	l.Log(WarnLevel, msg, context)
}

// Trace logs a trace message with optional context.
func (l *SimpleLogger) Trace(msg string, context map[string]interface{}) {
	l.Log(TraceLevel, msg, context)
}

//---------------------
// Helper Functions
//---------------------

// shouldLog checks if a message should be logged based on the log level.
func shouldLog(currentLevel, messageLevel string) bool {
	levels := []string{TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel}
	currentLevelIndex := indexOf(levels, currentLevel)
	messageLevelIndex := indexOf(levels, messageLevel)
	return messageLevelIndex >= currentLevelIndex
}

// indexOf returns the index of the level in the list.
func indexOf(levels []string, level string) int {
	for i, l := range levels {
		if l == level {
			return i
		}
	}
	return -1
}

// getCallerInfo retrieves caller info if needed (could be expanded later).
func (l *SimpleLogger) getCallerInfo() string {
	// Get the caller function and file/line
	_, file, line, ok := runtime.Caller(2) // Skip 2 levels to get the caller's file and line number
	if ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return "unknown"
}

//---------------------
// Global Logger
//---------------------

var globalLogger *SimpleLogger

// GetLogger returns the global logger instance.
func GetLogger() *SimpleLogger {
	if globalLogger == nil {
		cfg := LoadConfigFromEnv()
		globalLogger = NewLogger(cfg)
	}
	return globalLogger
}

//---------------------
// Log Shortcuts
//---------------------

// Info logs an info message with optional context.
func Info(msg string, context map[string]interface{}) {
	GetLogger().Info(msg, context)
}

// Error logs an error message with optional context.
func Error(msg string, context map[string]interface{}) {
	GetLogger().Error(msg, context)
}

// Debug logs a debug message with optional context.
func Debug(msg string, context map[string]interface{}) {
	GetLogger().Debug(msg, context)
}

// Warn logs a warn message with optional context.
func Warn(msg string, context map[string]interface{}) {
	GetLogger().Warn(msg, context)
}

// Trace logs a trace message with optional context.
func Trace(msg string, context map[string]interface{}) {
	GetLogger().Trace(msg, context)
}
