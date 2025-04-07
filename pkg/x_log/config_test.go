package x_log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoadConfig tests the behavior of LoadConfig with different configurations.
func TestLoadConfig(t *testing.T) {
	// Test when the config file doesn't exist (should return default config)
	t.Run("FileNotFound", func(t *testing.T) {
		// Path to a non-existent config file
		path := "./non_existent_config.json"

		// Attempt to load config from the non-existent path
		cfg, err := LoadConfig(path)

		// Assert that no error occurred and the default config is used
		assert.NoError(t, err)
		assert.Equal(t, defaultConfig, *cfg) // Config should match the default config
	})

	// Test when the config file exists and has valid JSON
	t.Run("ValidConfig", func(t *testing.T) {
		// Create a temporary config file with custom settings
		tmpFile, err := os.CreateTemp("", "test_config_*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		// Write a valid JSON config to the temp file
		customConfig := `{
			"Level": "debug",
			"LogFile": "logs/test.log",
			"ToConsole": true,
			"ToFile": true,
			"Style": "light",
			"MaxSize": 20,
			"MaxBackups": 10,
			"MaxAge": 30,
			"Compress": false
		}`
		_, err = tmpFile.WriteString(customConfig)
		if err != nil {
			t.Fatal(err)
		}

		// Load the config from the temp file
		cfg, err := LoadConfig(tmpFile.Name())
		if err != nil {
			t.Fatal(err)
		}

		// Assert the config values are correctly read
		assert.Equal(t, "debug", cfg.Level)
		assert.Equal(t, "logs/test.log", cfg.LogFile)
		assert.True(t, cfg.ToConsole)
		assert.True(t, cfg.ToFile)
		assert.Equal(t, "light", cfg.Style)
		assert.Equal(t, 20, cfg.MaxSize)
		assert.Equal(t, 10, cfg.MaxBackups)
		assert.Equal(t, 30, cfg.MaxAge)
		assert.False(t, cfg.Compress)
	})

	// Test when the config file contains invalid JSON
	t.Run("InvalidConfig", func(t *testing.T) {
		// Create a temporary file with invalid JSON
		tmpFile, err := os.CreateTemp("", "test_invalid_config_*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		// Write invalid JSON (missing closing brace)
		invalidConfig := `{
			"Level": "debug",
			"LogFile": "logs/test.log",
			"ToConsole": true,
			"ToFile": true,
			"Style": "light",
			"MaxSize": 20,
			"MaxBackups": 10,
			"MaxAge": 30,
			"Compress": false
		` // Missing closing brace
		_, err = tmpFile.WriteString(invalidConfig)
		if err != nil {
			t.Fatal(err)
		}

		// Attempt to load the config from the invalid JSON file
		_, err = LoadConfig(tmpFile.Name())

		// Assert that an error occurred due to invalid JSON
		assert.Error(t, err)
	})
}
