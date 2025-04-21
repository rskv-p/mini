// file:mini/pkg/x_log/zap.go
package x_log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var traceLogger *zap.SugaredLogger

//---------------------
// ZAP LOGGER IMPL
//---------------------

type zapLoggerImpl struct {
	loggerLevel *zap.AtomicLevel
	mainLogger  *zap.SugaredLogger
}

func (l *zapLoggerImpl) DebugEnabled() bool {
	return l.loggerLevel.Enabled(zapcore.DebugLevel)
}

func (l *zapLoggerImpl) TraceEnabled() bool {
	return traceEnabled && l.DebugEnabled()
}

func (l *zapLoggerImpl) Trace(args ...any) {
	if traceEnabled {
		traceLogger.Debug(args...)
	}
}

func (l *zapLoggerImpl) Tracef(format string, args ...any) {
	if traceEnabled {
		traceLogger.Debugf(format, args...)
	}
}

func (l *zapLoggerImpl) Debug(args ...any)         { l.mainLogger.Debug(args...) }
func (l *zapLoggerImpl) Info(args ...any)          { l.mainLogger.Info(args...) }
func (l *zapLoggerImpl) Warn(args ...any)          { l.mainLogger.Warn(args...) }
func (l *zapLoggerImpl) Error(args ...any)         { l.mainLogger.Error(args...) }
func (l *zapLoggerImpl) Debugf(f string, a ...any) { l.mainLogger.Debugf(f, a...) }
func (l *zapLoggerImpl) Infof(f string, a ...any)  { l.mainLogger.Infof(f, a...) }
func (l *zapLoggerImpl) Warnf(f string, a ...any)  { l.mainLogger.Warnf(f, a...) }
func (l *zapLoggerImpl) Errorf(f string, a ...any) { l.mainLogger.Errorf(f, a...) }

func (l *zapLoggerImpl) Structured() StructuredLogger {
	return &zapStructuredLoggerImpl{lvl: l.loggerLevel, zl: l.mainLogger.Desugar()}
}

//---------------------
// STRUCTURED LOGGER
//---------------------

type zapStructuredLoggerImpl struct {
	lvl *zap.AtomicLevel
	zl  *zap.Logger
}

func (l *zapStructuredLoggerImpl) Trace(msg string, fields ...Field) {
	if traceEnabled {
		l.zl.Debug(msg, convertFields(fields)...) // trace -> debug
	}
}

func (l *zapStructuredLoggerImpl) Debug(msg string, fields ...Field) {
	l.zl.Debug(msg, convertFields(fields)...)
}

func (l *zapStructuredLoggerImpl) Info(msg string, fields ...Field) {
	l.zl.Info(msg, convertFields(fields)...)
}

func (l *zapStructuredLoggerImpl) Warn(msg string, fields ...Field) {
	l.zl.Warn(msg, convertFields(fields)...)
}

func (l *zapStructuredLoggerImpl) Error(msg string, fields ...Field) {
	l.zl.Error(msg, convertFields(fields)...)
}

func convertFields(fields []Field) []zap.Field {
	zFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zFields[i] = f.(zap.Field)
	}
	return zFields
}

//---------------------
// ROOT LOGGER FROM CONFIG
//---------------------

func newZapRootLoggerWithOutput(cfg *Config) Logger {
	encoderCfg := newEncoderConfig()
	writer := buildZapOutputWithRotation(cfg)

	atomicLevel := zap.NewAtomicLevelAt(toZapLogLevel(cfg.Level))

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		writer,
		atomicLevel,
	)

	opts := []zap.Option{zap.AddCallerSkip(1)}
	if cfg.WithCaller {
		opts = append(opts, zap.AddCaller())
	}

	zl := zap.New(core, opts...)
	if cfg.WithTrace {
		traceLogger = zl.Sugar()
	}

	return &zapLoggerImpl{
		loggerLevel: &atomicLevel,
		mainLogger:  zl.Named(cfg.Name).Sugar(),
	}
}

//---------------------
// ENCODER CONFIG
//---------------------

func newEncoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "timestamp"
	cfg.ConsoleSeparator = getLogSeparator()

	cfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		formatted := fmt.Sprintf("[%02d-%02d %02d:%02d:%02d]",
			t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(),
		)
		enc.AppendString(logStyles.Timestamp.Render(formatted))
	}

	cfg.EncodeName = func(name string, enc zapcore.PrimitiveArrayEncoder) {
		styled := logStyles.DefaultKeyStyle.Render("[" + name + "]")
		enc.AppendString(styled + " -")
	}

	cfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		level := strings.ToLower(l.String())
		if style, ok := logStyles.Levels[level]; ok {
			enc.AppendString(style.Render(padLevel(level)))
		} else {
			enc.AppendString(padLevel(level))
		}
	}

	cfg.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.CallerKey = "caller"
	return cfg
}

func padLevel(level string) string {
	switch level {
	case "info":
		return "INF"
	case "warn":
		return "WRN"
	case "error":
		return "ERR"
	case "panic":
		return "PNC"
	case "debug":
		return "DBG"
	default:
		return strings.ToUpper(level)
	}
}

//---------------------
// OUTPUT WRITER
//---------------------

func buildZapOutputWithRotation(cfg *Config) zapcore.WriteSyncer {
	var writers []zapcore.WriteSyncer

	for _, path := range cfg.Outputs {
		switch path {
		case "stdout":
			writers = append(writers, zapcore.Lock(os.Stdout))
		case "stderr":
			writers = append(writers, zapcore.Lock(os.Stderr))
		default:
			writers = append(writers, zapcore.AddSync(&lumberjack.Logger{
				Filename:   path,
				MaxSize:    cfg.RotateMaxMB,
				MaxBackups: cfg.RotateBackups,
				MaxAge:     cfg.RotateMaxAge,
				Compress:   cfg.RotateCompress,
			}))
		}
	}

	if len(writers) == 0 {
		return zapcore.AddSync(os.Stdout)
	}
	return zapcore.NewMultiWriteSyncer(writers...)
}

//---------------------
// HELPERS
//---------------------

func toZapLogLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel, TraceLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func traceLevelEncoder(_ zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[TRACE]")
}

func setZapLogLevel(logger Logger, level Level) {
	if impl, ok := logger.(*zapLoggerImpl); ok {
		impl.loggerLevel.SetLevel(toZapLogLevel(level))
	}
}

func newZapChildLogger(logger Logger, name string) (Logger, error) {
	if impl, ok := logger.(*zapLoggerImpl); ok {
		newZl := impl.mainLogger.Named(name)
		return &zapLoggerImpl{
			loggerLevel: impl.loggerLevel,
			mainLogger:  newZl,
		}, nil
	}
	return nil, fmt.Errorf("invalid zapLoggerImpl")
}

func newZapChildLoggerWithFields(logger Logger, fields ...Field) (Logger, error) {
	if impl, ok := logger.(*zapLoggerImpl); ok {
		newZl := impl.mainLogger.With(fields...)
		return &zapLoggerImpl{
			loggerLevel: impl.loggerLevel,
			mainLogger:  newZl,
		}, nil
	}
	return nil, fmt.Errorf("invalid zapLoggerImpl")
}

func zapSync(logger Logger) {
	if impl, ok := logger.(*zapLoggerImpl); ok {
		_ = impl.mainLogger.Sync()
	}
}
