package x_cfg

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	mu     sync.RWMutex
	values = map[string]any{}
	path   string
)

// Load loads configuration from a JSON file.
func Load(file string) error {
	mu.Lock()
	defer mu.Unlock()

	// Read the file
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", file, err)
	}

	// Unmarshal JSON data into the config values
	if err := json.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("failed to unmarshal config data from %s: %w", file, err)
	}

	// Store the path of the loaded file
	path = file
	return nil
}

// Reload reloads the configuration from the original file.
func Reload() error {
	if path == "" {
		return fmt.Errorf("no config file loaded, cannot reload")
	}
	return Load(path)
}

// Get retrieves a value from the config by key.
func Get(key string) any {
	mu.RLock()
	defer mu.RUnlock()
	return values[key]
}

// Set adds or updates a value in the config.
func Set(key string, val any) {
	mu.Lock()
	defer mu.Unlock()
	values[key] = val
}

// Delete removes a key from the config.
func Delete(key string) {
	mu.Lock()
	defer mu.Unlock()
	delete(values, key)
}

// All returns a copy of the full config map.
func All() map[string]any {
	mu.RLock()
	defer mu.RUnlock()

	// Clone the map to avoid race conditions with concurrent access
	clone := make(map[string]any, len(values))
	for k, v := range values {
		clone[k] = v
	}
	return clone
}
