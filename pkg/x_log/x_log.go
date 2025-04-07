// Package x_log provides an abstraction over zap logger with styling support.
package x_log

import (
	"errors"
	"log"
	"strings"
	"sync"

	"go.uber.org/zap/zapcore"
)

var (
	ErrInvalidLevelValue  = errors.New("invalid_level_value")
	ErrInvalidFormatValue = errors.New("invalid_format_value")

	globalMu sync.RWMutex
	global   Logger
)

type (
	Level        int8
	OutputFormat string

	Logger interface {
		Error(args ...any)
		Errorw(msg string, keysAndValues ...any)
		Debug(args ...any)
		Debugw(msg string, keysAndValues ...any)
		Info(args ...any)
		Infow(msg string, keysAndValues ...any)
		Warn(args ...any)
		Warnw(msg string, keysAndValues ...any)
	}

	ConfigurableLogger interface {
		Logger
		Configure(OutputFormat, Level) error
		SetStyles(*Styles)
		SetReportTimestamp(bool)
		SetTimeFormat(string)
		StdLogger() *log.Logger
	}
)

const (
	DebugLevel Level = Level(0)
	InfoLevel  Level = Level(1)
	WarnLevel  Level = Level(2)
	ErrorLevel Level = Level(3)

	OutputConsole OutputFormat = "console"
	OutputJSON    OutputFormat = "json"
)

func NewLogger() (ConfigurableLogger, error) {
	return &wrappedLogger{
		styles:     DefaultStyles(),
		showTime:   true,
		timeFormat: "01-02 15:04:05",
		format:     OutputConsole,
		minLevel:   zapcore.InfoLevel,
	}, nil
}

func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	default:
		return 0, ErrInvalidLevelValue
	}
}

func ParseFormat(s string) (OutputFormat, error) {
	switch strings.ToLower(s) {
	case "console":
		return OutputConsole, nil
	case "json":
		return OutputJSON, nil
	default:
		return OutputConsole, ErrInvalidFormatValue
	}
}

func SetGlobal(l Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	global = l
}

func L() Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return global
}

// Additions to wrappedLogger for stdlib compatibility
func (l *wrappedLogger) StdLogger() *log.Logger {
	return log.New(l, "", 0)
}

func (l *wrappedLogger) Write(p []byte) (n int, err error) {
	l.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}
