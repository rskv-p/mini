package config_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rskv-p/mini/config"
	"github.com/stretchr/testify/assert"
)

func TestConfig_New_WithDefaults(t *testing.T) {
	cfg, err := config.New(config.WithDefaults(map[string]any{
		"service_name": "demo",
		"bus_addr":     "127.0.0.1:4150",
		"log_level":    "debug",
		"port":         "8080",
	}))
	assert.NoError(t, err)
	assert.Equal(t, "demo", cfg.MustString("service_name"))
}

func TestConfig_GetAndMustString(t *testing.T) {
	cfg, _ := config.New(config.WithDefaults(map[string]any{
		"key1": "value1",
		"key2": 123,
	}))
	v, ok := cfg.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", v)
	assert.Equal(t, "value1", cfg.MustString("key1"))
	assert.Equal(t, "", cfg.MustString("key2")) // not a string
	assert.Equal(t, "", cfg.MustString("missing"))
}

func TestConfig_StringAndDump(t *testing.T) {
	cfg, _ := config.New(config.WithDefaults(map[string]any{
		"key": "value",
	}))
	str := cfg.String()
	assert.Contains(t, str, `"key": "value"`)

	var buf bytes.Buffer
	cfg.Dump(&buf)
	assert.Contains(t, buf.String(), `"key": "value"`)
}

func TestConfig_Validate(t *testing.T) {
	cfg, _ := config.New(config.WithDefaults(map[string]any{
		"service_name": "demo",
		"bus_addr":     "127.0.0.1:4150",
		"log_level":    "info",
		"port":         "8080",
	}))
	assert.NoError(t, cfg.Validate())

	// Missing required field
	cfg2, _ := config.New(config.WithDefaults(map[string]any{
		"service_name": "only",
	}))
	err := cfg2.Validate()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "bus_addr"))
}
