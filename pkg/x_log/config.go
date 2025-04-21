// file:mini/pkg/x_log/config.go
package x_log

import (
	"os"
	"strconv"
	"strings"
)

//---------------------
// Log Configuration Struct
//---------------------

type Config struct {
	Name       string   // logger name
	Level      Level    // log level
	Format     Format   // console or json
	Outputs    []string // paths: stdout, stderr, or file paths
	WithTrace  bool     // enable TRACE level
	WithCaller bool     // include caller info

	// Rotation (for file output)
	RotateMaxMB    int  // max size in MB
	RotateMaxAge   int  // days to keep
	RotateBackups  int  // how many files
	RotateCompress bool // compress old files
}

func LoadConfigFromEnv() *Config {
	cfg := &Config{
		Name:           "mini",
		Level:          DefaultLogLevel,
		Format:         DefaultLogFormat,
		Outputs:        parseOutputPaths(),
		WithTrace:      false,
		WithCaller:     strings.ToLower(os.Getenv("MINI_LOG_CALLER")) == "true",
		RotateMaxMB:    getIntEnv(EnvLogFileMaxMB, 100),
		RotateMaxAge:   getIntEnv(EnvLogFileMaxAge, 7),
		RotateBackups:  getIntEnv(EnvLogFileMaxBack, 5),
		RotateCompress: strings.ToLower(os.Getenv(EnvLogFileCompress)) == "true",
	}

	switch strings.ToUpper(os.Getenv(EnvKeyLogLevel)) {
	case "TRACE":
		cfg.Level = DebugLevel
		cfg.WithTrace = true
	case "DEBUG":
		cfg.Level = DebugLevel
	case "INFO":
		cfg.Level = InfoLevel
	case "WARN":
		cfg.Level = WarnLevel
	case "ERROR":
		cfg.Level = ErrorLevel
	}

	if strings.ToUpper(os.Getenv(EnvKeyLogFormat)) == "JSON" {
		cfg.Format = FormatJson
	}

	return cfg
}

func parseOutputPaths() []string {
	var paths []string
	stream := strings.ToLower(os.Getenv(EnvLogConsoleStream))
	if stream == "stdout" || stream == "stderr" {
		paths = append(paths, stream)
	}
	if file := os.Getenv(EnvLogFilePath); file != "" {
		paths = append(paths, file)
	}
	if len(paths) == 0 {
		paths = append(paths, "stdout")
	}
	return paths
}

func getIntEnv(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if v, err := strconv.Atoi(val); err == nil {
			return v
		}
	}
	return fallback
}
