package x_log

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ---------- Log Level ----------
type Level int

const (
	DebugLevel Level = -4
	InfoLevel  Level = 0
	WarnLevel  Level = 4
	ErrorLevel Level = 8
	FatalLevel Level = 12
)

// Logger is an alias for zerolog.Logger
type Logger = zerolog.Logger

// ---------- Config ----------
type Config struct {
	Level       string `json:"Level"`       // log level (debug, info, etc)
	LogFile     string `json:"LogFile"`     // path to log file
	ToConsole   bool   `json:"ToConsole"`   // output to stderr
	ToFile      bool   `json:"ToFile"`      // output to file
	ColoredFile bool   `json:"ColoredFile"` // use ANSI color in file
	Style       string `json:"Style"`       // "dark" or "light"
	MaxSize     int    `json:"MaxSize"`     // file size in MB
	MaxBackups  int    `json:"MaxBackups"`  // rotated files to keep
	MaxAge      int    `json:"MaxAge"`      // days to keep logs
	Compress    bool   `json:"Compress"`    // gzip old logs
}

//
// ---------- Initialization ----------

// Init allows flexible initialization:
// Init()                   → default config and "main" module
// Init("cfg.json")         → config path and default "main" module
// Init("cfg.json", "api")  → full control
func Init(args ...string) {
	var (
		path   string
		module = "main"
	)
	if len(args) > 0 {
		path = args[0]
	}
	if len(args) > 1 && args[1] != "" {
		module = args[1]
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		log.Error().Err(err).Msg("failed to load logger config")
		cfg = &defaultConfig
	}
	InitWithConfig(cfg, module)
}

// initWithConfig sets global logger from Config
func InitWithConfig(cfg *Config, module string) {
	if cfg == nil {
		cfg = &defaultConfig
	}
	log.Logger = NewWithConfig(cfg, module)
}

// New returns a scoped logger (does not affect global log.Logger)
func New(module ...string) zerolog.Logger {
	base := log.Logger.With()
	name := "main"
	if len(module) > 0 && module[0] != "" {
		name = module[0]
	}
	return base.Str("module", name).Logger()
}

//
// ---------- Internal logger builder ----------

func NewWithConfig(cfg *Config, module string) zerolog.Logger {
	zerolog.TimeFieldFormat = "02-01 15:04:05"
	zerolog.TimestampFunc = time.Now

	var writers []io.Writer
	styles := DefaultStylesByName(cfg.Style)

	// Console writer (ensure correct output to console)
	if cfg.ToConsole || isatty.IsTerminal(os.Stdout.Fd()) {
		styles.Out = os.Stderr
		// Use ConsoleWriterWithStyles for consistent console output
		consoleWriter := ConsoleWriterWithStyles(styles)
		writers = append(writers, consoleWriter)
	}

	// File writer
	if cfg.ToFile && cfg.LogFile != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		if cfg.ColoredFile {
			styles.Out = fileWriter
			writers = append(writers, ConsoleWriterWithStyles(styles)) // Colored output to file
		} else {
			writers = append(writers, fileWriter) // Plain output to file
		}
	}

	if len(writers) == 0 {
		writers = append(writers, io.Discard) // If no writers, discard logs
	}

	builder := zerolog.New(io.MultiWriter(writers...)).
		With().
		Timestamp().
		Caller().
		Str("module", module)

	logger := builder.Logger()

	// Global level setup
	if lvl, err := zerolog.ParseLevel(strings.ToLower(cfg.Level)); err == nil {
		zerolog.SetGlobalLevel(lvl)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Log initialization
	logger.Info().
		Str("level", cfg.Level).
		Str("log_file", cfg.LogFile).
		Str("module", module).
		Bool("console", cfg.ToConsole).        // Correct usage of Bool()
		Bool("file", cfg.ToFile).              // Correct usage of Bool()
		Bool("colored_file", cfg.ColoredFile). // Correct usage of Bool()
		Str("style", cfg.Style).
		Msg("logger initialized")

	return logger
}

//
// ---------- Aliases (shortcuts) ----------

func Trace() *zerolog.Event { return log.Trace() }
func Debug() *zerolog.Event { return log.Debug() }
func Info() *zerolog.Event  { return log.Info() }
func Warn() *zerolog.Event  { return log.Warn() }
func Error() *zerolog.Event { return log.Error() }
func Fatal() *zerolog.Event { return log.Fatal() }
func Panic() *zerolog.Event { return log.Panic() }

//
// ---------- Structured logging ----------

// WithFields logs map as fields (on Info level)
func WithFields(fields map[string]any) *zerolog.Event {
	e := log.Info()
	for k, v := range fields {
		e = e.Interface(k, v)
	}
	return e
}

// WithFieldsAt logs map as fields at given level
func WithFieldsAt(level string, fields map[string]any) *zerolog.Event {
	var e *zerolog.Event
	switch strings.ToLower(level) {
	case "trace":
		e = log.Trace()
	case "debug":
		e = log.Debug()
	case "warn":
		e = log.Warn()
	case "error":
		e = log.Error()
	case "fatal":
		e = log.Fatal()
	case "panic":
		e = log.Panic()
	default:
		e = log.Info()
	}
	for k, v := range fields {
		e = e.Interface(k, v)
	}
	return e
}

//
// ---------- Context integration ----------

type ctxKey struct{}

var loggerKey ctxKey = struct{}{}

// From extracts logger from context, or returns global logger
func From(ctx context.Context) *zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zerolog.Logger); ok {
		return logger
	}
	return &log.Logger
}

// WithLogger injects logger into context
func WithLogger(ctx context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
