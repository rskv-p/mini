package runn_cfg

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
	ApplicationName        = "Runn service"
	ApplicationVersion     = "1.0.0"        // Default version
	ApplicationDescription = "Runn Service" // Default description
	ApplicationQueueGroup  = "default_group"
	ApplicationHost        = "127.0.0.1"
	ApplicationPort        = 4008
	DefaultConfig          = ".data/cfg/runn.config.json" // Default config file path
	LocalConfig            = ".runn.config.json"          // Local config file path
	GlobalConfig           = "RUNN_CONFIG"                // Environment variable for custom config file path
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
var defaultConfig = RunnConfig{
	Name:        ApplicationName,
	Version:     ApplicationVersion,     // Default version
	Description: ApplicationDescription, // Default description
	QueueGroup:  ApplicationQueueGroup,  // Default queue group
	Host:        ApplicationHost,        // Default host
	Port:        ApplicationPort,        // Default port
	Services: []ServiceConfig{
		{
			Name:        "Nats service",
			Path:        ".data/bim/s_nats",
			Args:        []string{},
			AutoRestart: true,
			DependsOn:   []string{},
		},
	},
	Logger: defaultLogger, // Default logger config
}

// ---------- Config Structure ----------

// RunnConfig is the structure for the service manager configuration.
type RunnConfig struct {
	Name        string          `json:"name"`        // Service name
	Version     string          `json:"version"`     // Service version
	Description string          `json:"description"` // Description
	QueueGroup  string          `json:"queue_group"` // Queue group for endpoints
	Host        string          `json:"host"`        // NATS bind host
	Port        int             `json:"port"`        // NATS bind port
	Services    []ServiceConfig `json:"services"`    // List of services to launch
	Logger      x_log.Config    `json:"Logger"`      // Logger configuration
}

// ServiceConfig defines the structure of a service configuration.
type ServiceConfig struct {
	Name        string   `json:"name"`         // Logical service name
	Path        string   `json:"path"`         // Path to binary or main.go
	Args        []string `json:"args"`         // Optional arguments
	AutoRestart bool     `json:"auto_restart"` // Restart on crash
	DependsOn   []string `json:"depends_on"`   // List of service names this depends on
}

// ---------- LoadConfig Function ----------

// LoadConfig checks multiple locations for the configuration file and loads it.
// It looks in the following order:
// 1. Local config file (.runn_config.json)
// 2. Global config environment variable (RUNN_CONFIG)
// 3. Default config file (_data/runn_config.json)
func LoadConfig() (*RunnConfig, error) {
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
func applyDefaults(cfg *RunnConfig) {
	if len(cfg.Name) == 0 {
		cfg.Name = defaultConfig.Name
	}

	if len(cfg.Version) == 0 {
		cfg.Version = defaultConfig.Version
	}

	if len(cfg.Description) == 0 {
		cfg.Description = defaultConfig.Description
	}

	if len(cfg.QueueGroup) == 0 {
		cfg.QueueGroup = defaultConfig.QueueGroup
	}

	if len(cfg.Host) == 0 {
		cfg.Host = defaultConfig.Host
	}

	if cfg.Port < 1 {
		cfg.Port = defaultConfig.Port
	}

	if len(cfg.Services) == 0 {
		cfg.Services = defaultConfig.Services
	}
	if cfg.Logger == (x_log.Config{}) {
		cfg.Logger = defaultConfig.Logger
	}
}

// ---------- readConfig Function ----------

// readConfig reads a JSON config file from the specified path and unmarshals it into an RunnConfig object.
func readConfig(path string) (*RunnConfig, error) {
	// Read the file data
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err // Return error if file reading fails
	}

	// Unmarshal the JSON data into the RunnConfig struct
	var cfg RunnConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err) // Return error if parsing fails
	}

	// Return the parsed config
	return &cfg, nil
}

func ResolveStartupOrder(cfg RunnConfig) ([]ServiceConfig, error) {
	graph := map[string][]string{}
	services := map[string]ServiceConfig{}

	for _, svc := range cfg.Services {
		services[svc.Name] = svc
		graph[svc.Name] = svc.DependsOn
	}

	visited := map[string]bool{}
	temp := map[string]bool{}
	result := []ServiceConfig{}

	var visit func(string) error
	visit = func(n string) error {
		if temp[n] {
			return fmt.Errorf("circular dependency detected on %q", n)
		}
		if !visited[n] {
			temp[n] = true
			for _, dep := range graph[n] {
				if _, ok := services[dep]; !ok {
					return fmt.Errorf("unknown dependency %q for service %q", dep, n)
				}
				if err := visit(dep); err != nil {
					return err
				}
			}
			visited[n] = true
			temp[n] = false
			result = append(result, services[n])
		}
		return nil
	}

	for name := range services {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}
