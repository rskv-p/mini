package x_log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

// TestInit tests if the Init function initializes the logger with default config.
func TestInit(t *testing.T) {
	Init() // Use default config
	assert.NotNil(t, log.Logger)
	assert.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel()) // Default level should be info
}

// TestInitWithConfig tests if InitWithConfig correctly sets up the logger.
func TestInitWithConfig(t *testing.T) {
	cfg := &Config{
		Level: "debug",
	}

	InitWithConfig(cfg, "testModule")
	assert.NotNil(t, log.Logger)
	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel()) // Level should be set to debug
}

// TestNew tests if the New function creates a scoped logger.
func TestNew(t *testing.T) {
	// Create a logger with the module "testModule"
	logger := New("testModule")

	// Capture the log output in a buffer
	var buf bytes.Buffer
	logger = logger.Output(&buf)

	// Log a test message
	logger.Info().Msg("Testing logger")

	// Check if the "module": "testModule" field is in the log output (JSON format)
	output := buf.String()

	// Assert that the log output contains the "module" key and the "testModule" value
	assert.Contains(t, output, `"module":"testModule"`)
}

// TestConsoleLogging tests if console logging works as expected.
func TestConsoleLogging(t *testing.T) {
	// Create a temporary buffer to capture output
	var buf bytes.Buffer
	consoleWriter := ConsoleWriterWithStyles(&Styles{
		Out: &buf,
	})
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	// Log a message
	logger.Info().Msg("Test message")

	// Check if the message is in the output buffer
	assert.Contains(t, buf.String(), "Test message")
}

// TestFileLogging tests if the file logging works correctly.
func TestFileLogging(t *testing.T) {
	// Temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatal("failed to create temp file:", err)
	}
	defer os.Remove(tmpFile.Name())

	// Create logger with file output
	cfg := &Config{
		ToFile:  true,
		LogFile: tmpFile.Name(),
		Level:   "info",
	}
	InitWithConfig(cfg, "testModule")

	// Log a message
	Info().Msg("Test file logging")

	// Read the file to check the content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal("failed to read log file:", err)
	}

	// Assert the message is in the log file
	assert.Contains(t, string(content), "Test file logging")
}

// TestContextLogger tests logging with context integration.
func TestContextLogger(t *testing.T) {
	// Create a logger with a "module" field
	logger := New("testModule")

	// Create a buffer to capture the log output
	var buf bytes.Buffer
	// Attach the logger to the context
	ctx := WithLogger(context.Background(), &logger)

	// Assign the logger to zerolog's global logger for output capture
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

	// Log a message using the logger
	logger.Info().Msg("Testing context logger")

	// Retrieve the logger from context and log another message
	ctxLogger := From(ctx)
	ctxLogger.Info().Msg("Message from context logger")

	// Check if the "module" field is present in the logger's context by inspecting the output
	output := buf.String()

	// Split the output into lines and unmarshal each line as a separate JSON object
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue // Skip empty lines
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Fatalf("Error unmarshaling log entry: %v", err)
		}

		// Check if the "module" field is present and matches the expected value
		if module, exists := logEntry["module"]; exists {
			if module != "testModule" {
				t.Errorf("Expected module 'testModule', got %v", module)
			}
		} else {
			t.Errorf("Missing 'module' field in log entry")
		}
	}
}

// TestLoggingLevels tests if the logger respects different logging levels.
func TestLoggingLevels(t *testing.T) {
	var buf bytes.Buffer
	consoleWriter := ConsoleWriterWithStyles(&Styles{
		Out: &buf,
	})
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	// Log messages at different levels
	logger.Debug().Msg("Debug message")
	logger.Info().Msg("Info message")
	logger.Warn().Msg("Warn message")
	logger.Error().Msg("Error message")

	// Assert the presence of each message in the buffer
	assert.Contains(t, buf.String(), "Debug message")
	assert.Contains(t, buf.String(), "Info message")
	assert.Contains(t, buf.String(), "Warn message")
	assert.Contains(t, buf.String(), "Error message")
}

// TestWithFields tests structured logging with custom fields.
func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	consoleWriter := ConsoleWriterWithStyles(&Styles{
		Out: &buf,
	})
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	// Log with fields
	logger.Info().Str("user", "john").Str("action", "login").Msg("User login")

	// Check if fields are logged correctly
	assert.Contains(t, buf.String(), "user=john")
	assert.Contains(t, buf.String(), "action=login")
	assert.Contains(t, buf.String(), "User login")
}

// TestErrorLogging tests logging errors correctly.
func TestErrorLogging(t *testing.T) {
	var buf bytes.Buffer
	consoleWriter := ConsoleWriterWithStyles(&Styles{
		Out: &buf,
	})
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	// Log an error
	err := fmt.Errorf("Sample error")
	logger.Error().Err(err).Msg("Test error logging")

	// Assert that the error is logged
	assert.Contains(t, buf.String(), "Sample error")
}
