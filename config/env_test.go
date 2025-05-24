package config_test

import (
	"os"
	"testing"

	"github.com/rskv-p/mini/config"
	"github.com/stretchr/testify/assert"
)

func TestGetEnvStr(t *testing.T) {
	_ = os.Setenv("ENV_STR", "hello")
	assert.Equal(t, "hello", config.GetEnvStr("ENV_STR", "default"))

	_ = os.Unsetenv("ENV_STR")
	assert.Equal(t, "default", config.GetEnvStr("ENV_STR", "default"))
}

func TestGetEnvInt(t *testing.T) {
	_ = os.Setenv("ENV_INT", "123")
	assert.Equal(t, 123, config.GetEnvInt("ENV_INT", 42))

	_ = os.Setenv("ENV_INT", "bad")
	assert.Equal(t, 42, config.GetEnvInt("ENV_INT", 42))

	_ = os.Unsetenv("ENV_INT")
	assert.Equal(t, 42, config.GetEnvInt("ENV_INT", 42))
}

func TestGetEnvFloat(t *testing.T) {
	_ = os.Setenv("ENV_FLOAT", "3.14")
	assert.InDelta(t, 3.14, config.GetEnvFloat("ENV_FLOAT", 1.0), 0.001)

	_ = os.Setenv("ENV_FLOAT", "bad")
	assert.Equal(t, 1.0, config.GetEnvFloat("ENV_FLOAT", 1.0))

	_ = os.Unsetenv("ENV_FLOAT")
	assert.Equal(t, 1.0, config.GetEnvFloat("ENV_FLOAT", 1.0))
}

func TestGetEnvBool(t *testing.T) {
	_ = os.Setenv("ENV_BOOL", "true")
	assert.True(t, config.GetEnvBool("ENV_BOOL", false))

	_ = os.Setenv("ENV_BOOL", "1")
	assert.True(t, config.GetEnvBool("ENV_BOOL", false))

	_ = os.Setenv("ENV_BOOL", "yes")
	assert.True(t, config.GetEnvBool("ENV_BOOL", false))

	_ = os.Setenv("ENV_BOOL", "false")
	assert.False(t, config.GetEnvBool("ENV_BOOL", true))

	_ = os.Setenv("ENV_BOOL", "0")
	assert.False(t, config.GetEnvBool("ENV_BOOL", true))

	_ = os.Setenv("ENV_BOOL", "no")
	assert.False(t, config.GetEnvBool("ENV_BOOL", true))

	_ = os.Setenv("ENV_BOOL", "invalid")
	assert.True(t, config.GetEnvBool("ENV_BOOL", true)) // fallback

	_ = os.Unsetenv("ENV_BOOL")
	assert.False(t, config.GetEnvBool("ENV_BOOL", false))
}
