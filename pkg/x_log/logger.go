package x_log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap/zapcore"
)

// wrappedLogger is a minimal implementation of Logger with styling support.
type wrappedLogger struct {
	styles     *Styles
	showTime   bool
	timeFormat string
	format     OutputFormat
	minLevel   zapcore.Level
	writer     io.Writer
}

// SetStyles applies custom styles or resets to default.
func (l *wrappedLogger) SetStyles(s *Styles) {
	if s == nil {
		s = DefaultStyles()
	}
	l.styles = s
}

// SetReportTimestamp toggles timestamp visibility.
func (l *wrappedLogger) SetReportTimestamp(enabled bool) {
	l.showTime = enabled
}

// SetTimeFormat sets the timestamp format.
func (l *wrappedLogger) SetTimeFormat(format string) {
	l.timeFormat = format
}

// SetOutput sets the log output writer.
func (l *wrappedLogger) SetOutput(w io.Writer) {
	if w != nil {
		l.writer = w
	}
}

// Configure applies log format and minimum level.
func (l *wrappedLogger) Configure(format OutputFormat, lvl Level) error {
	l.format = format
	l.minLevel = zapcore.Level(lvl)
	l.writer = os.Stderr // default output
	return nil
}

// log renders and writes a full log line.
func (l *wrappedLogger) log(level zapcore.Level, msg string, kvs ...any) {
	if level < l.minLevel {
		return
	}
	var line string
	if l.format == OutputJSON {
		line = l.formatJSON(level, msg, kvs)
	} else {
		line = l.renderLine(level, msg, kvs)
	}
	if l.writer == nil {
		l.writer = os.Stderr
	}
	fmt.Fprintln(l.writer, line)
}

func (l *wrappedLogger) formatJSON(level zapcore.Level, msg string, kvs []any) string {
	entry := map[string]any{
		"level":   level.String(),
		"message": msg,
	}
	if l.showTime {
		entry["time"] = time.Now().Format(l.timeFormat)
	}
	for i := 0; i < len(kvs); i += 2 {
		key := fmt.Sprint(kvs[i])
		if i+1 < len(kvs) {
			entry[key] = kvs[i+1]
		} else {
			entry[key] = "<missing>"
		}
	}
	data, _ := json.Marshal(entry)
	return string(data)
}

// --- Log methods by level ---

func (l *wrappedLogger) Debug(args ...any) {
	l.log(zapcore.DebugLevel, fmt.Sprint(args...))
}

func (l *wrappedLogger) Debugw(msg string, kvs ...any) {
	l.log(zapcore.DebugLevel, msg, kvs...)
}

func (l *wrappedLogger) Info(args ...any) {
	l.log(zapcore.InfoLevel, fmt.Sprint(args...))
}

func (l *wrappedLogger) Infow(msg string, kvs ...any) {
	l.log(zapcore.InfoLevel, msg, kvs...)
}

func (l *wrappedLogger) Warn(args ...any) {
	l.log(zapcore.WarnLevel, fmt.Sprint(args...))
}

func (l *wrappedLogger) Warnw(msg string, kvs ...any) {
	l.log(zapcore.WarnLevel, msg, kvs...)
}

func (l *wrappedLogger) Error(args ...any) {
	l.log(zapcore.ErrorLevel, fmt.Sprint(args...))
}

func (l *wrappedLogger) Errorw(msg string, kvs ...any) {
	l.log(zapcore.ErrorLevel, msg, kvs...)
}
