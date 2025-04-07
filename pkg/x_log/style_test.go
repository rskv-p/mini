package x_log

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestStylesCheck verifies that the correct colors are applied to different log levels
// based on the selected theme (dark theme in this case).
func TestStylesCheck(t *testing.T) {
	// Test dark theme styles for each log level
	styles := DefaultStylesByName("dark")

	// Verify the styles for different log levels
	assert.NotNil(t, styles.Levels[InfoLevel], "InfoLevel style should be defined")
	assert.NotNil(t, styles.Levels[WarnLevel], "WarnLevel style should be defined")
	assert.NotNil(t, styles.Levels[ErrorLevel], "ErrorLevel style should be defined")
	assert.NotNil(t, styles.Levels[FatalLevel], "FatalLevel style should be defined")
}

// TestLogLevelMapping checks if the log levels are correctly mapped to zerolog levels.
func TestLogLevelMapping(t *testing.T) {
	// Initialize with a config that sets the log level to debug
	cfg := &Config{
		Level: "debug",
	}
	InitWithConfig(cfg, "testModule")

	// Assert that the global level is set to DebugLevel
	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel(), "Global log level should be set to DebugLevel")
}

// TestApplyConfigToLogger tests if the Init function applies the correct configuration to the logger.
func TestApplyConfigToLogger(t *testing.T) {
	// Initialize the logger with a custom config
	cfg := &Config{
		Level: "error", // Set log level to "error"
	}
	InitWithConfig(cfg, "testModule")

	// Assert that the global level is set to ErrorLevel
	assert.Equal(t, zerolog.ErrorLevel, zerolog.GlobalLevel(), "Global log level should be set to ErrorLevel")
}

// TestApplyDefaultConfig verifies that the default config is used when no config is provided.
func TestApplyDefaultConfig(t *testing.T) {
	// Initialize the logger with the default config (no config file)
	Init()

	// Assert that the default log file path is used
	assert.Equal(t, "logs/app.log", defaultConfig.LogFile, "Default log file path should be 'logs/app.log'")

	// Assert that the default style is set to "dark"
	assert.Equal(t, "dark", defaultConfig.Style, "Default style should be 'dark'")

	// Assert the default max size of the log file is 10 MB
	assert.Equal(t, 10, defaultConfig.MaxSize, "Default MaxSize should be 10 MB")
}
