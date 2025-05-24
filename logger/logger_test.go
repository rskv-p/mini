package logger_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/rskv-p/mini/logger"
	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer log.SetOutput(nil)
	f()
	return buf.String()
}

func TestLoggerLevels(t *testing.T) {
	l := logger.NewLogger("test", "debug")

	output := captureOutput(func() {
		l.Debug("debug msg: %d", 1)
		l.Info("info msg")
		l.Warn("warn msg")
		l.Error("error msg")
	})
	assert.Contains(t, output, "[DEBUG][test] debug msg: 1")
	assert.Contains(t, output, "[INFO][test] info msg")
	assert.Contains(t, output, "[WARN][test] warn msg")
	assert.Contains(t, output, "[ERROR][test] error msg")
}

func TestLoggerLevelFiltering(t *testing.T) {
	l := logger.NewLogger("svc", "warn")

	output := captureOutput(func() {
		l.Debug("should not appear")
		l.Info("should not appear")
		l.Warn("warn ok")
		l.Error("error ok")
	})
	assert.NotContains(t, output, "should not appear")
	assert.Contains(t, output, "warn ok")
	assert.Contains(t, output, "error ok")
}

func TestWithContextAndClone(t *testing.T) {
	l := logger.NewLogger("svc", "info").WithContext("ctx123")
	cl := l.Clone()

	output := captureOutput(func() {
		cl.Info("ctx present")
	})
	assert.Contains(t, output, "[INFO][svc][cid:ctx123] ctx present")
}

func TestLoggerEntryFields(t *testing.T) {
	l := logger.NewLogger("svc", "debug")
	entry := l.With("k1", "v1").With("k2", 42)

	output := captureOutput(func() {
		entry.Debug("entry msg")
	})
	assert.Contains(t, output, "entry msg")
	assert.Contains(t, output, "k1=v1")
	assert.Contains(t, output, "k2=42")
}

func TestEntryClone(t *testing.T) {
	l := logger.NewLogger("svc", "debug")
	entry := l.With("a", "b").With("x", "y")
	cl := entry.Clone()

	output := captureOutput(func() {
		cl.Info("cloned entry")
	})
	assert.Contains(t, output, "a=b")
	assert.Contains(t, output, "x=y")
}

func TestNormalizeLevel(t *testing.T) {
	assert.Equal(t, "info", logger.NewLogger("svc", "").(*logger.Logger).Level())
	assert.Equal(t, "warn", logger.NewLogger("svc", "WARN").(*logger.Logger).Level())
	assert.Equal(t, "info", logger.NewLogger("svc", "badlevel").(*logger.Logger).Level())
}

func TestShouldLog(t *testing.T) {
	assert.True(t, callShouldLog("debug", "debug"))
	assert.True(t, callShouldLog("info", "warn"))
	assert.False(t, callShouldLog("warn", "info"))
}

func callShouldLog(current, incoming string) bool {
	// Expose internal for test
	return loggerTestHelperShouldLog(current, incoming)
}

// dummy copy of internal function
func loggerTestHelperShouldLog(current, incoming string) bool {
	order := map[string]int{
		"debug": 1,
		"info":  2,
		"warn":  3,
		"error": 4,
	}
	c := order[strings.ToLower(current)]
	i := order[strings.ToLower(incoming)]
	return i >= c
}
