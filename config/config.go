package config

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// IConfig defines the interface for accessing and validating config.
type IConfig interface {
	Validate() error
	String() string
	Dump(w io.Writer)
	Get(key string) (any, bool)
	MustString(key string) string
}

// Config is the default implementation of IConfig.
type Config struct {
	values map[string]any
}

// New creates a new config from default, file or environment.
func New(opts ...Option) (*Config, error) {
	cfg := &Config{values: make(map[string]any)}
	for _, o := range opts {
		if err := o(cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// Validate required fields.
func (c *Config) Validate() error {
	required := []string{"service_name", "bus_addr", "log_level", "port"}
	var missing []string
	for _, key := range required {
		if _, ok := c.values[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing config keys: %s", strings.Join(missing, ", "))
	}
	return nil
}

// String returns pretty-printed JSON.
func (c *Config) String() string {
	data, _ := json.MarshalIndent(c.values, "", "  ")
	return string(data)
}

// Dump writes JSON config to writer.
func (c *Config) Dump(w io.Writer) {
	data, _ := json.MarshalIndent(c.values, "", "  ")
	_, _ = w.Write(data)
}

// Get returns a value from config.
func (c *Config) Get(key string) (any, bool) {
	v, ok := c.values[key]
	return v, ok
}

// MustString returns a string value or fallback empty string.
func (c *Config) MustString(key string) string {
	v, _ := c.Get(key)
	s, _ := v.(string)
	return s
}
