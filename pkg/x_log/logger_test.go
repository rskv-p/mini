// file:mini/pkg/x_log/logger_test.go
package x_log_test

import (
	"os"
	"testing"

	"github.com/rskv-p/mini/pkg/x_log"
)

func TestLoggerOutput(t *testing.T) {
	t.Setenv("MINI_LOG_LEVEL", "DEBUG")
	t.Setenv("MINI_LOG_FILE", "_build/test.log")
	t.Setenv("MINI_LOG_FILE_COMPRESS", "true")
	t.Setenv("MINI_LOG_FILE_MAX_MB", "1")
	t.Setenv("MINI_LOG_FILE_BACKUPS", "2")
	t.Setenv("MINI_LOG_FILE_MAX_AGE", "1")
	t.Setenv("MINI_LOG_FORMAT", "console")
	t.Setenv("MINI_LOG_CTX", "true")

	x_log.Debug("debug message")
	x_log.Info("info message")
	x_log.Warn("warn message")
	x_log.Error("error message")

	logger := x_log.RootLogger().Structured()
	logger.Info("structured log", x_log.FString("module", "test"), x_log.FInt("code", 200))

	t.Log("Log written to console and _build/test.log")
	t.Log("Waiting 1s to allow log flush...")
	t.Log("Check the output file manually if needed")
	t.Cleanup(func() {
		x_log.Sync()
		_ = os.Remove("_build/test.log")
	})
	t.Log("[OK] test completed")

}
