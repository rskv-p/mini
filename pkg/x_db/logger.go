package x_db

import (
	"context"
	"time"

	"github.com/rskv-p/mini/pkg/x_log"
	"gorm.io/gorm/logger"
)

//
// ---------- GORM log adapter (based on x_log) ----------

// logAdapter implements GORM logger.Interface using x_log
type logAdapter struct {
	Logger        *x_log.Logger   // underlying zerolog-based logger
	LogLevel      logger.LogLevel // current log level
	SlowThreshold time.Duration   // duration to treat query as slow
}

// newlogAdapter creates a new log adapter for GORM using given logger and level
func newlogAdapter(zlog *x_log.Logger, level logger.LogLevel) logger.Interface {
	return &logAdapter{
		Logger:        zlog,
		LogLevel:      level,
		SlowThreshold: 200 * time.Millisecond,
	}
}

// LogMode sets the logging level for the adapter
func (l *logAdapter) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info logs an info-level message
func (l *logAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Logger.Info().Msgf(msg, data...)
	}
}

// Warn logs a warning-level message
func (l *logAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Logger.Warn().Msgf(msg, data...)
	}
}

// Error logs an error-level message
func (l *logAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Logger.Error().Msgf(msg, data...)
	}
}

// Trace logs SQL queries, highlighting slow or failed ones
func (l *logAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	e := l.Logger.With().
		Str("elapsed", elapsed.String()).
		Int64("rows", rows).
		Logger()

	switch {
	case err != nil && l.LogLevel >= logger.Error:
		e.Error().Err(err).Msg(sql)
	case elapsed > l.SlowThreshold && l.LogLevel >= logger.Warn:
		e.Warn().Msgf("SLOW SQL: %s", sql)
	case l.LogLevel >= logger.Info:
		e.Info().Msg(sql)
	}
}
