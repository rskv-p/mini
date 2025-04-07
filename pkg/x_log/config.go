package x_log

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

//
// ---------- Defaults ----------

const defaultConfigPath = "./xlog.json"

var defaultConfig = Config{
	Level:       "info",
	LogFile:     "logs/app.log",
	ToConsole:   true,
	ToFile:      false,
	ColoredFile: false,
	Style:       "dark",
	MaxSize:     10, // MB
	MaxBackups:  5,  // rotated files
	MaxAge:      7,  // days
	Compress:    true,
}

//
// ---------- LoadConfig ----------

// LoadConfig reads JSON config from file.
// If path is empty, uses XLOG_CONFIG or ./xlog.json.
func LoadConfig(path string) (*Config, error) {
	// Resolve path
	if path == "" {
		path = os.Getenv("XLOG_CONFIG")
		if path == "" {
			path = defaultConfigPath
		}
	}

	// Read file
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Return default config if file not found
			return &defaultConfig, nil
		}
		return nil, fmt.Errorf("failed to read config from %s: %w", path, err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from %s: %w", path, err)
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

//
// ---------- Defaults Fill ----------

// applyDefaults fills missing config values from defaultConfig
func applyDefaults(cfg *Config) {
	if cfg.Level == "" {
		cfg.Level = defaultConfig.Level
	}
	if cfg.LogFile == "" {
		cfg.LogFile = defaultConfig.LogFile
	}
	if cfg.Style == "" {
		cfg.Style = defaultConfig.Style
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = defaultConfig.MaxSize
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = defaultConfig.MaxBackups
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = defaultConfig.MaxAge
	}
}
