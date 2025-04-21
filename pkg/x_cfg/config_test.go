// file:mini/pkg/x_cfg/config_test.go
package x_cfg_test

import (
	"os"
	"testing"

	x_cfg "github.com/rskv-p/mini/pkg/x_cfg"

	"github.com/stretchr/testify/require"
)

func TestConfig_LoadGetSetDeleteReload(t *testing.T) {
	tmpFile := "test_config.json"

	//---------------------
	// Create test file
	//---------------------
	data := `{
		"debug": true,
		"port": 8080,
		"name": "mini-app"
	}`
	err := os.WriteFile(tmpFile, []byte(data), 0644)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	//---------------------
	// Load config
	//---------------------
	err = x_cfg.Load(tmpFile)
	require.NoError(t, err)

	//---------------------
	// Get
	//---------------------
	require.Equal(t, true, x_cfg.Get("debug"))
	require.Equal(t, 8080.0, x_cfg.Get("port")) // JSON numbers = float64
	require.Equal(t, "mini-app", x_cfg.Get("name"))

	//---------------------
	// Set
	//---------------------
	x_cfg.Set("version", "1.0.0")
	require.Equal(t, "1.0.0", x_cfg.Get("version"))

	//---------------------
	// Delete
	//---------------------
	x_cfg.Delete("name")
	require.Nil(t, x_cfg.Get("name"))

	//---------------------
	// All
	//---------------------
	all := x_cfg.All()
	require.Len(t, all, 3)
	require.Contains(t, all, "debug")
	require.Contains(t, all, "port")
	require.Contains(t, all, "version")

	//---------------------
	// Reload
	//---------------------
	err = x_cfg.Reload()
	require.NoError(t, err)
	require.Equal(t, "mini-app", x_cfg.Get("name")) // name restored from file
}
