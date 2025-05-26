// file: mini/logger/logger.go
package logger

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

var _ ILogger = (*Logger)(nil)
var _ LoggerEntry = (*entry)(nil)

// ----------------------------------------------------
// Interfaces
// ----------------------------------------------------

type ILogger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	WithContext(contextID string) ILogger
	With(key string, value any) LoggerEntry
	SetLevel(level string)
	Clone() ILogger
}

type LoggerEntry interface {
	With(key string, value any) LoggerEntry
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Clone() LoggerEntry
}

// ----------------------------------------------------
// Constants and levels
// ----------------------------------------------------

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

var levelOrder = map[string]int{
	LevelDebug: 1,
	LevelInfo:  2,
	LevelWarn:  3,
	LevelError: 4,
}

// ----------------------------------------------------
// Logger implementation
// ----------------------------------------------------

type Logger struct {
	service   string
	contextID string
	level     string
}

func NewLogger(serviceName, level string) ILogger {
	return &Logger{
		service: serviceName,
		level:   normalizeLevel(level),
	}
}

func (l *Logger) SetLevel(level string) {
	l.level = normalizeLevel(level)
}

func (l *Logger) WithContext(contextID string) ILogger {
	return &Logger{
		service:   l.service,
		contextID: contextID,
		level:     l.level,
	}
}

func (l *Logger) Clone() ILogger {
	return &Logger{
		service:   l.service,
		contextID: l.contextID,
		level:     l.level,
	}
}

func (l *Logger) With(key string, value any) LoggerEntry {
	return &entry{
		parent: l,
		fields: map[string]any{key: value},
	}
}

func (l *Logger) Debug(msg string, args ...any) { l.log(LevelDebug, msg, args...) }
func (l *Logger) Info(msg string, args ...any)  { l.log(LevelInfo, msg, args...) }
func (l *Logger) Warn(msg string, args ...any)  { l.log(LevelWarn, msg, args...) }
func (l *Logger) Error(msg string, args ...any) { l.log(LevelError, msg, args...) }

func (l *Logger) log(level, msg string, args ...any) {
	if !shouldLog(l.level, level) {
		return
	}
	prefix := fmt.Sprintf("[%s][%s]", strings.ToUpper(level), l.service)
	if l.contextID != "" {
		prefix += fmt.Sprintf("[cid:%s]", l.contextID)
	}
	log.Printf("%s %s", prefix, fmt.Sprintf(msg, args...))
}

// ----------------------------------------------------
// Entry (structured log builder)
// ----------------------------------------------------

type entry struct {
	parent *Logger
	fields map[string]any
}

func (e *entry) With(key string, value any) LoggerEntry {
	if e.fields == nil {
		e.fields = make(map[string]any)
	}
	e.fields[key] = value
	return e
}

func (e *entry) Clone() LoggerEntry {
	copied := make(map[string]any, len(e.fields))
	for k, v := range e.fields {
		copied[k] = v
	}
	return &entry{
		parent: e.parent.Clone().(*Logger),
		fields: copied,
	}
}

func (e *entry) Debug(msg string, args ...any) { e.log(LevelDebug, msg, args...) }
func (e *entry) Info(msg string, args ...any)  { e.log(LevelInfo, msg, args...) }
func (e *entry) Warn(msg string, args ...any)  { e.log(LevelWarn, msg, args...) }
func (e *entry) Error(msg string, args ...any) { e.log(LevelError, msg, args...) }

func (e *entry) log(level, msg string, args ...any) {
	if !shouldLog(e.parent.level, level) {
		return
	}

	prefix := fmt.Sprintf("[%s][%s]", strings.ToUpper(level), e.parent.service)
	if e.parent.contextID != "" {
		prefix += fmt.Sprintf("[cid:%s]", e.parent.contextID)
	}

	// format fields
	fieldParts := make([]string, 0, len(e.fields))
	for k, v := range e.fields {
		fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
	}
	sort.Strings(fieldParts)

	meta := ""
	if len(fieldParts) > 0 {
		meta = " | " + strings.Join(fieldParts, " ")
	}

	log.Printf("%s %s%s", prefix, fmt.Sprintf(msg, args...), meta)
}

// ----------------------------------------------------
// Helpers
// ----------------------------------------------------

func normalizeLevel(level string) string {
	switch strings.ToLower(level) {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		return strings.ToLower(level)
	default:
		return LevelInfo
	}
}

func shouldLog(current, incoming string) bool {
	c := levelOrder[normalizeLevel(current)]
	i := levelOrder[normalizeLevel(incoming)]
	return i >= c
}

func (l *Logger) Level() string {
	return l.level
}
