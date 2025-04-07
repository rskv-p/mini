package x_db

import (
	"context"
	"fmt"

	"github.com/rskv-p/mini/pkg/x_log"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//
// ---------- Public Interfaces ----------

// DB defines the interface for database layer
type DB interface {
	Migrate(ctx context.Context) error
	Close() error
	DB() *gorm.DB
}

// DbType represents supported database backends
type DbType string

const (
	DbSqlite   DbType = "sqlite"
	DbPostgres DbType = "postgres"
)

// Config defines database connection settings
type Config struct {
	Type      DbType `json:"Type"`      // "sqlite" or "postgres"
	DSN       string `json:"DSN"`       // connection string
	LogFile   string `json:"LogFile"`   // unused (log handled via x_log)
	LogLevel  string `json:"LogLevel"`  // silent, error, warn, info
	LogToFile bool   `json:"LogToFile"` // unused (log handled via x_log)
}

// DAO implements the DB interface using GORM
type DAO struct {
	db *gorm.DB
}

//
// ---------- Constructor ----------

// NewDatabase creates a new DAO with the provided config
func New(cfg Config) (*DAO, error) {
	var dialector gorm.Dialector

	switch cfg.Type {
	case DbSqlite:
		dialector = sqlite.Open(cfg.DSN)
	case DbPostgres:
		dialector = postgres.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported db type: %s", cfg.Type)
	}

	gormCfg := &gorm.Config{
		Logger: gormLoggerFromConfig(cfg),
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, err
	}

	return &DAO{db: db}, nil
}

//
// ---------- DAO Methods ----------

// Migrate performs automatic schema migration for the provided models
func (s *DAO) Migrate(ctx context.Context, models ...interface{}) error {
	return s.db.WithContext(ctx).AutoMigrate(models...)
}

// Close closes the underlying SQL connection
func (s *DAO) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DB exposes raw *gorm.DB for advanced usage
func (s *DAO) DB() *gorm.DB {
	return s.db
}

//
// ---------- Logging ----------

// gormLoggerFromConfig builds a GORM logger from Config using x_log
func gormLoggerFromConfig(cfg Config) logger.Interface {
	level := parseGormLogLevel(cfg.LogLevel)
	log := x_log.New("gorm")
	return newlogAdapter(&log, level)
}

// parseGormLogLevel maps string to GORM log level
func parseGormLogLevel(lvl string) logger.LogLevel {
	switch lvl {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	case "warn":
		return logger.Warn
	default:
		return logger.Warn
	}
}
