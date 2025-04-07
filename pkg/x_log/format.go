package x_log

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

// renderLine formats the full log line with time, level, message and key-value pairs.
func (l *wrappedLogger) renderLine(level zapcore.Level, msg string, kvs []any) string {
	var b strings.Builder

	// Timestamp
	if l.showTime {
		timestamp := time.Now().Format(l.timeFormat)
		b.WriteString(l.styles.Time.Render(timestamp))
		b.WriteByte(' ')
	}

	// Level (styled or fallback)
	if style, ok := l.styles.Levels[level]; ok {
		b.WriteString(style.Render())
	} else {
		b.WriteString(level.CapitalString())
	}
	b.WriteByte(' ')

	// Message
	b.WriteString(msg)

	// Key-value fields
	if len(kvs) > 0 {
		b.WriteByte(' ')
		b.WriteString(strings.Join(l.renderKVs(kvs), " "))
	}

	return b.String()
}

// renderKVs formats and styles key-value pairs.
func (l *wrappedLogger) renderKVs(kvs []any) []string {
	result := make([]string, 0, len(kvs)/2)

	for i := 0; i < len(kvs); i += 2 {
		key := fmt.Sprint(kvs[i])

		var val string
		if i+1 < len(kvs) {
			val = fmt.Sprint(kvs[i+1])
		} else {
			val = "<missing>"
		}

		// Apply key style
		if style, ok := l.styles.Keys[key]; ok {
			key = style.Render(key)
		}

		// Apply value style
		if style, ok := l.styles.Values[key]; ok {
			val = style.Render(val)
		}

		result = append(result, fmt.Sprintf("%s=%s", key, val))
	}

	return result
}
