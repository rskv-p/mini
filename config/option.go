// file: mini/config/option.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Option is a functional config initializer.
type Option func(*Config) error

// WithDefaults sets initial config values.
func WithDefaults(defaults map[string]any) Option {
	return func(c *Config) error {
		for k, v := range defaults {
			c.values[k] = v
		}
		return nil
	}
}

// FromJSON loads config from a JSON file.
func FromJSON(path string) Option {
	return func(c *Config) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read config file: %w", err)
		}
		data = ReplaceEnvVars(data)

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse config json: %w", err)
		}
		for k, v := range raw {
			c.values[strings.ToLower(k)] = v
		}
		return nil
	}
}

// FromEnv loads config values from environment variables with prefix.
func FromEnv(prefix string) Option {
	return func(c *Config) error {
		for _, e := range os.Environ() {
			if strings.HasPrefix(e, prefix) {
				kv := strings.SplitN(e, "=", 2)
				if len(kv) == 2 {
					key := strings.ToLower(strings.TrimPrefix(kv[0], prefix))
					c.values[key] = ParseEnvValue(kv[1])
				}
			}
		}
		return nil
	}
}

// parseEnvValue tries to interpret strings like "true", "123", etc.
func ParseEnvValue(v string) any {
	v = strings.TrimSpace(v)
	if strings.EqualFold(v, "true") {
		return true
	}
	if strings.EqualFold(v, "false") {
		return false
	}
	if i, err := strconv.Atoi(v); err == nil {
		return i
	}
	return v
}

// replaceEnvVars replaces ${ENV_VAR} in raw JSON string.
func ReplaceEnvVars(data []byte) []byte {
	return []byte(os.Expand(string(data), func(key string) string {
		return os.Getenv(key)
	}))
}
