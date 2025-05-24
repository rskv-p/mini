package config_test

import (
	"os"
	"testing"

	"github.com/rskv-p/mini/config"
	"github.com/stretchr/testify/assert"
)

func TestWithDefaults(t *testing.T) {
	cfg, err := config.New(config.WithDefaults(map[string]any{
		"foo": "bar",
		"num": 123,
	}))
	assert.NoError(t, err)
	assert.Equal(t, "bar", cfg.MustString("foo"))
	v, ok := cfg.Get("num")
	assert.True(t, ok)
	assert.Equal(t, 123, v)
}

func TestFromJSON_Success(t *testing.T) {
	tmpFile := "test_config.json"
	content := `{"Port": 8080, "Log_Level": "debug", "Token": "${MY_TOKEN}"}`
	_ = os.WriteFile(tmpFile, []byte(content), 0644)
	_ = os.Setenv("MY_TOKEN", "secret-token")

	cfg, err := config.New(config.FromJSON(tmpFile))
	assert.NoError(t, err)
	assert.Equal(t, "debug", cfg.MustString("log_level"))
	assert.Equal(t, "secret-token", cfg.MustString("token"))

	_ = os.Remove(tmpFile)
	_ = os.Unsetenv("MY_TOKEN")
}

func TestFromJSON_InvalidFile(t *testing.T) {
	cfg, err := config.New(config.FromJSON("nonexistent.json"))
	assert.Nil(t, cfg)
	assert.Error(t, err)
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	tmpFile := "invalid.json"
	_ = os.WriteFile(tmpFile, []byte("{bad json"), 0644)

	cfg, err := config.New(config.FromJSON(tmpFile))
	assert.Nil(t, cfg)
	assert.Error(t, err)

	_ = os.Remove(tmpFile)
}

func TestFromEnv(t *testing.T) {
	_ = os.Setenv("SRV_TEST_A", "true")
	_ = os.Setenv("SRV_TEST_B", "123")
	_ = os.Setenv("SRV_TEST_C", "value")

	cfg, err := config.New(config.FromEnv("SRV_TEST_"))
	assert.NoError(t, err)

	valA, okA := cfg.Get("a")
	valB, okB := cfg.Get("b")
	valC, okC := cfg.Get("c")

	assert.True(t, okA)
	assert.True(t, okB)
	assert.True(t, okC)

	assert.IsType(t, true, valA)
	assert.Equal(t, true, valA)

	assert.IsType(t, 0, valB)
	assert.Equal(t, 123, valB)

	assert.IsType(t, "string", valC)
	assert.Equal(t, "value", valC)

	_ = os.Unsetenv("SRV_TEST_A")
	_ = os.Unsetenv("SRV_TEST_B")
	_ = os.Unsetenv("SRV_TEST_C")
}

func TestParseEnvValue(t *testing.T) {
	assert.Equal(t, true, config.ParseEnvValue("true"))
	assert.Equal(t, false, config.ParseEnvValue("false"))
	assert.Equal(t, "foo", config.ParseEnvValue("foo"))
	assert.NotNil(t, config.ParseEnvValue("123")) // will still return string, not int
}

func TestReplaceEnvVars(t *testing.T) {
	_ = os.Setenv("DEMO_KEY", "demo-value")
	in := []byte(`{"key": "${DEMO_KEY}"}`)
	out := config.ReplaceEnvVars(in)
	assert.Contains(t, string(out), "demo-value")
	_ = os.Unsetenv("DEMO_KEY")
}
