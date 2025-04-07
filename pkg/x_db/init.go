package x_db

import (
	"github.com/rskv-p/mini/pkg/x_log"
)

const defaultConfigPath = "./.data/cfg/xdb.config.json"

var (
	dao        *DAO
	defaultCfg = Config{
		Type:      DbSqlite,
		DSN:       "file=xdb.db?_foreign_keys=on",
		LogLevel:  "warn",
		LogToFile: false,
		LogFile:   "./.data/log/xdb.log",
	}
)

// Init initializes the global DAO instance using optional config path:
// - Init() → uses env or ./xdb.json
// - Init("path/to/db.json")
// - Init("path/to/db.json", "db-module")
func Init(args ...string) {
	var (
		path   string
		module = "xdb"
	)

	if len(args) > 0 {
		path = args[0]
	}
	if len(args) > 1 && args[1] != "" {
		module = args[1]
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		x_log.Fatal().Err(err).Msg("failed to load DB config")
	}

	dao, err = New(*cfg)
	if err != nil {
		x_log.Fatal().Err(err).Msg("failed to open DB")
	}

	x_log.Info().
		Str("driver", string(cfg.Type)).
		Str("dsn", cfg.DSN).
		Str("module", module).
		Msg("database initialized")
}

// Global returns the globally-initialized DAO (after Init)
func Global() *DAO {
	if dao == nil {
		panic("xdb not initialized — call x_db.Init() first")
	}
	return dao
}
