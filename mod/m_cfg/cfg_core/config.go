package cfg_core

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	mu     sync.RWMutex // Mutex to ensure thread safety
	values = map[string]any{}
	path   string // Path to the loaded configuration file
)

//---------------------
// Load and Save Config
//---------------------

// LoadConfig loads the configuration from a specified file.
func LoadConfig(file string) error {
	mu.Lock()
	defer mu.Unlock()

	// Read the file
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", file, err)
	}

	// Parse the JSON data into the configuration map
	if err := json.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("failed to unmarshal config data from %s: %w", file, err)
	}

	// Save the path of the loaded file
	path = file
	return nil
}

// SaveConfig saves the current configuration to a file.
func SaveConfig(file string) error {
	mu.Lock()
	defer mu.Unlock()

	// Convert the configuration map to JSON
	data, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %w", err)
	}

	// Write the data to the specified file
	if err := os.WriteFile(file, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", file, err)
	}

	return nil
}

//---------------------
// Get, Set, Delete
//---------------------

// Get retrieves a configuration value by key.
func Get(moduleName, key string) any {
	mu.RLock()
	defer mu.RUnlock()
	return values[key]
}

// Set sets a configuration value for the specified key.
func Set(moduleName, key string, val any) {
	mu.Lock()
	defer mu.Unlock()
	values[key] = val
}

// Delete removes a key from the configuration.
func Delete(moduleName, key string) {
	mu.Lock()
	defer mu.Unlock()
	delete(values, key)
}

//---------------------
// Full Configuration
//---------------------

// All returns a copy of the entire configuration.
func All() map[string]any {
	mu.RLock()
	defer mu.RUnlock()

	// Clone the map to prevent race conditions during access
	clone := make(map[string]any, len(values))
	for k, v := range values {
		clone[k] = v
	}
	return clone
}

//---------------------
// Reload Configuration
//---------------------

// Reload reloads the configuration from the file if the file path is set.
func Reload(file string) error {
	mu.Lock()
	defer mu.Unlock()

	// If no file path is set, return an error
	if path == "" {
		return fmt.Errorf("no config file loaded")
	}

	// If a new file path is provided, load the configuration again
	if file != "" {
		path = file
	}

	// Read the configuration file again
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Parse the JSON data into the configuration map
	if err := json.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("failed to unmarshal config data from %s: %w", path, err)
	}

	return nil
}
