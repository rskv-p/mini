package nats_cfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rskv-p/mini/pkg/x_log"
)

// ---------- Constants ----------

// Constants for config file paths and environment variables
const (
	ApplicationName        = "Nats service"
	ApplicationVersion     = "1.0.0"        // Default version
	ApplicationDescription = "NATS Service" // Default description
	ApplicationQueueGroup  = "default_group"
	ApplicationHost        = "127.0.0.1"
	ApplicationPort        = 4222
	DefaultConfig          = ".data/cfg/nats.config.json" // Default config file path
	LocalConfig            = ".nats.config.json"          // Local config file path
	GlobalConfig           = "NATS_CONFIG"                // Environment variable for custom config file path
)

// ---------- Default Config ----------

var defaultLogger = x_log.Config{
	Level:       "info",
	LogFile:     ".data/log/nats.log",
	ToConsole:   true,
	ToFile:      true,
	ColoredFile: false,
	Style:       "dark",
	MaxSize:     10,
	MaxBackups:  5,
	MaxAge:      7,
	Compress:    true,
}

// defaultConfig provides default values for the fields that can be overridden in the config file.
var defaultConfig = NatsConfig{
	Name:        ApplicationName,
	Version:     ApplicationVersion,     // Default version
	Description: ApplicationDescription, // Default description
	QueueGroup:  ApplicationQueueGroup,  // Default queue group
	Host:        ApplicationHost,        // Default host
	Port:        ApplicationPort,        // Default port
	JetStream:   true,                   // Default JetStream setting
	Logger:      defaultLogger,          // Default logger config
}

// ---------- Config Structure ----------

// NatsConfig is the structure for the NATS service configuration
type NatsConfig struct {
	Name        string       `json:"name" default:""`          // Service name
	Version     string       `json:"version" default:""`       // Service version
	Description string       `json:"description" default:""`   // Description
	QueueGroup  string       `json:"queue_group" default:""`   // Queue group for endpoints
	Host        string       `json:"host" default:""`          // NATS bind host
	Port        int          `json:"port" default:"0"`         // NATS bind port
	JetStream   bool         `json:"jetstream" default:"true"` // Enable JetStream
	Logger      x_log.Config `json:"Logger"`                   // Logger configuration
}

// ---------- LoadConfig Function ----------

// LoadConfig checks multiple locations for the configuration file and loads it.
// It looks in the following order:
// 1. Local config file (.nats_config.json)
// 2. Global config environment variable (NATS_CONFIG)
// 3. Default config file (_data/nats_config.json)
func LoadConfig() (*NatsConfig, error) {
	// Define possible paths to check for the config file
	paths := []string{
		LocalConfig,
		os.Getenv(GlobalConfig), // Check the environment variable
		DefaultConfig,           // Default config path
	}

	// Iterate through paths and try to load the config
	for _, path := range paths {
		if path == "" {
			continue // Skip empty paths
		}

		// Try reading and parsing the config from the path
		cfg, err := readConfig(path)
		if err == nil {
			// If config is successfully loaded, apply defaults if necessary
			applyDefaults(cfg)

			// Return the successfully loaded config
			return cfg, nil
		}

		// If file is not found, skip and continue with the next path
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
	}

	// If no valid config file is found, return an error
	return nil, fmt.Errorf("no valid config file found")
}

// ---------- applyDefaults Function ----------

// applyDefaults fills missing config values from defaultConfig
func applyDefaults(cfg *NatsConfig) {
	// Apply default values if necessary
	if cfg.Name == "" {
		cfg.Name = defaultConfig.Name
	}
	if cfg.Version == "" {
		cfg.Version = defaultConfig.Version
	}
	if cfg.Description == "" {
		cfg.Description = defaultConfig.Description
	}
	if cfg.QueueGroup == "" {
		cfg.QueueGroup = defaultConfig.QueueGroup
	}
	if cfg.Host == "" {
		cfg.Host = defaultConfig.Host
	}
	if cfg.Port == 0 {
		cfg.Port = defaultConfig.Port
	}
	if cfg.Logger == (x_log.Config{}) {
		cfg.Logger = defaultConfig.Logger
	}
}

// ---------- readConfig Function ----------

// readConfig reads a JSON config file from the specified path and unmarshals it into an NatsConfig object.
func readConfig(path string) (*NatsConfig, error) {
	// Read the file data
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err // Return error if file reading fails
	}

	// Unmarshal the JSON data into the NatsConfig struct
	var cfg NatsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err) // Return error if parsing fails
	}

	// Return the parsed config
	return &cfg, nil
}
