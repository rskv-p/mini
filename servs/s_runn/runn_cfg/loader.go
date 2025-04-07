package runn_cfg

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/mitchellh/mapstructure"
)

// Load loads the configuration from the specified file or environment variable
func Load(path string) error {
	// Set default config values
	config = defaultConfig

	// Determine the config file path (either from argument or environment variable)
	if path == "" {
		if envPath := os.Getenv("PROC_CFG"); envPath != "" {
			path = envPath
		} else {
			path = "./runn_config.json"
		}
	}

	// Read the configuration file
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Silently fallback to default values if the file doesn't exist
			return nil
		}
		return err
	}

	// Parse JSON data into a map
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Parse each configuration field
	if err := parseConfigFields(raw); err != nil {
		return err
	}

	// Parse preconfigured processes
	if v, ok := raw["preconfigured_processes"].([]interface{}); ok {
		for _, p := range v {
			var process PreconfiguredProcess
			if err := mapstructure.Decode(p, &process); err != nil {
				return err
			}
			config.PreconfiguredProcesses = append(config.PreconfiguredProcesses, process)
		}
	}

	return nil
}

// parseConfigFields parses the main configuration fields from the raw JSON data
func parseConfigFields(raw map[string]interface{}) error {
	if v, ok := raw["db_path"].(string); ok {
		config.DBPath = v
	}
	if v, ok := raw["log_level"].(string); ok {
		config.LogLevel = v
	}
	if v, ok := raw["log_to_file"].(bool); ok {
		config.LogToFile = v
	}
	if v, ok := raw["http_address"].(string); ok {
		config.HTTPAddress = v
	}
	if v, ok := raw["restart_max"].(float64); ok {
		config.RestartMax = int(v)
	}
	if v, ok := raw["timeout_sec"].(float64); ok {
		config.Timeout = time.Duration(v) * time.Second
	}
	if v, ok := raw["jwt_secret"].(string); ok {
		config.JwtSecret = v
	}
	if v, ok := raw["admin_password"].(string); ok {
		config.AdminDefaultPassword = v
	}
	if v, ok := raw["auth_enabled"].(bool); ok {
		config.AuthEnabled = v
	}
	return nil
}
