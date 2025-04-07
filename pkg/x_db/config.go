package x_db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// LoadConfig loads JSON config from file or environment fallback.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv("XDB_CONFIG")
		if path == "" {
			path = defaultConfigPath
		}
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &defaultCfg, nil
		}
		return nil, fmt.Errorf("cannot read DB config from %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse DB config from %s: %w", path, err)
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

// applyDefaults fills in default values for missing fields
func applyDefaults(cfg *Config) {
	if cfg.Type == "" {
		cfg.Type = defaultCfg.Type
	}
	if cfg.DSN == "" {
		cfg.DSN = defaultCfg.DSN
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = defaultCfg.LogLevel
	}
}
