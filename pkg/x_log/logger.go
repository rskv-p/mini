// file:mini/pkg/x_log/logger.go
package x_log

import (
	"os"
	"path/filepath"
	"strings"
)

//---------------------
// TYPES
//---------------------

type Level int

type Format int

type Logger interface {
	TraceEnabled() bool
	DebugEnabled() bool

	Trace(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)

	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)

	Structured() StructuredLogger
}

type StructuredLogger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

type Field = any

//---------------------
// LOG LEVELS
//---------------------

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

//---------------------
// LOG FORMATS
//---------------------

const (
	FormatConsole Format = iota
	FormatJson
)

//---------------------
// ENVIRONMENT
//---------------------

const (
	EnvKeyLogCtx        = "MINI_LOG_CTX"
	EnvKeyLogDateFormat = "MINI_LOG_DTFORMAT"
	EnvKeyLogLevel      = "MINI_LOG_LEVEL"
	EnvKeyLogFormat     = "MINI_LOG_FORMAT"
	EnvKeyLogSeparator  = "MINI_LOG_SEPARATOR"
	EnvLogConsoleStream = "MINI_LOG_CONSOLE_STREAM"
	EnvLogFilePath      = "MINI_LOG_FILE"
	EnvLogFileMaxMB     = "MINI_LOG_FILE_MAX_MB"
	EnvLogFileMaxAge    = "MINI_LOG_FILE_MAX_AGE"
	EnvLogFileMaxBack   = "MINI_LOG_FILE_BACKUPS"
	EnvLogFileCompress  = "MINI_LOG_FILE_COMPRESS"

	DefaultLogDateFormat = "01-02 15:04:05"
	DefaultLogLevel      = DebugLevel
	DefaultLogFormat     = FormatConsole
	DefaultLogSeparator  = " "
)

//---------------------
// GLOBALS
//---------------------

var (
	rootLogger   Logger
	ctxLogging   bool
	traceEnabled = true
	logStyles    = DefaultStylesDark()
)

//---------------------
// INITIALIZATION
//---------------------

func init() {
	cfg := LoadConfigFromEnv()
	if strings.ToLower(os.Getenv(EnvKeyLogCtx)) == "true" {
		ctxLogging = true
	}
	traceEnabled = cfg.WithTrace
	rootLogger = newZapRootLoggerWithOutput(cfg)
	SetLogLevel(rootLogger, cfg.Level)
}

//---------------------
// ACCESSORS
//---------------------

func CtxLoggingEnabled() bool {
	return ctxLogging
}

func RootLogger() Logger {
	return rootLogger
}

func SetLogLevel(logger Logger, level Level) {
	setZapLogLevel(logger, level)
}

func Sync() {
	zapSync(rootLogger)
}

//---------------------
// CHILD LOGGER HELPERS
//---------------------

func ChildLogger(logger Logger, name string) Logger {
	child, err := newZapChildLogger(logger, name)
	if err != nil {
		rootLogger.Warnf("unable to create child logger [%s]: %s", name, err)
		return logger
	}
	return child
}

func ChildLoggerWithFields(logger Logger, fields ...Field) Logger {
	child, err := newZapChildLoggerWithFields(logger, fields...)
	if err != nil {
		rootLogger.Warnf("unable to create child logger with fields: %s", err)
		return logger
	}
	return child
}

func CreateLoggerFromRef(logger Logger, contributionType, ref string) Logger {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimSuffix(ref, "/")
	dirs := strings.Split(ref, "/")

	switch {
	case len(dirs) >= 3:
		name := dirs[len(dirs)-1]
		acType := dirs[len(dirs)-2]
		if acType == "activity" || acType == "trigger" || acType == "connector" {
			category := dirs[len(dirs)-3]
			return ChildLogger(logger, strings.ToLower(category+"."+acType+"."+name))
		}
		return ChildLogger(logger, strings.ToLower(acType+"."+contributionType+"."+name))
	default:
		return ChildLogger(logger, strings.ToLower(contributionType+"."+filepath.Base(ref)))
	}
}

//---------------------
// UTILITIES
//---------------------

func ToLogLevel(levelStr string) Level {
	switch strings.ToUpper(levelStr) {
	case "TRACE":
		return DebugLevel
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	default:
		return DefaultLogLevel
	}
}

func getLogSeparator() string {
	if v, ok := os.LookupEnv(EnvKeyLogSeparator); ok && len(v) > 0 {
		return v
	}
	return DefaultLogSeparator
}

//---------------------
// STYLE HELPERS
//---------------------

func StyledLevel(level string) string {
	if s, ok := logStyles.Levels[level]; ok {
		return s.Render(strings.ToUpper(level))
	}
	return logStyles.DefaultKeyStyle.Render(strings.ToUpper(level))
}

func StyledKey(key string) string {
	if s, ok := logStyles.Keys[key]; ok {
		return s.Render(key)
	}
	return logStyles.DefaultKeyStyle.Render(key)
}

func StyledValue(key, value string) string {
	if s, ok := logStyles.Values[key]; ok {
		return s.Render(value)
	}
	return logStyles.DefaultValueStyle.Render(value)
}

func RenderStyledFields(fields map[string]string) string {
	var sb strings.Builder
	for k, v := range fields {
		sb.WriteString(StyledKey(k))
		sb.WriteString("=")
		sb.WriteString(StyledValue(k, v))
		sb.WriteString("  ")
	}
	return sb.String()
}

//---------------------
// DEFAULT LOGGER SHORTCUTS
//---------------------

func Trace(args ...any) { rootLogger.Trace(args...) }
func Debug(args ...any) { rootLogger.Debug(args...) }
func Info(args ...any)  { rootLogger.Info(args...) }
func Warn(args ...any)  { rootLogger.Warn(args...) }
func Error(args ...any) { rootLogger.Error(args...) }

func Tracef(format string, args ...any) { rootLogger.Tracef(format, args...) }
func Debugf(format string, args ...any) { rootLogger.Debugf(format, args...) }
func Infof(format string, args ...any)  { rootLogger.Infof(format, args...) }
func Warnf(format string, args ...any)  { rootLogger.Warnf(format, args...) }
func Errorf(format string, args ...any) { rootLogger.Errorf(format, args...) }
