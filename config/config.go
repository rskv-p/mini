// file: arc/service/config/config.go
package config

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime and embedded NSQ settings.
type Config struct {
	ServiceName               string      `json:"service_name"`
	BusAddr                   string      `json:"bus_addr"`
	LogLevel                  string      `json:"log_level"`
	Port                      int         `json:"port"`
	DevMode                   bool        `json:"dev_mode"`
	HCMemoryCriticalThreshold float64     `json:"hc_memory_critical"`
	HCMemoryWarningThreshold  float64     `json:"hc_memory_warning"`
	HCLoadCriticalThreshold   float64     `json:"hc_load_critical"`
	HCLoadWarningThreshold    float64     `json:"hc_load_warning"`
	NSQ                       NSQSettings `json:"nsq"`
}

type NSQSettings struct {
	TCPAddress             string        `json:"tcp"`
	HTTPAddress            string        `json:"http"`
	BroadcastAddress       string        `json:"broadcast"`
	AdminHTTPAddress       string        `json:"admin_http"`
	MemQueueSize           int64         `json:"mem_queue"`
	MaxMsgSize             int64         `json:"max_msg_size"`
	MsgTimeout             time.Duration `json:"msg_timeout"`
	SyncEvery              int64         `json:"sync_every"`
	SyncTimeout            time.Duration `json:"sync_timeout"`
	MaxRdyCount            int64         `json:"max_rdy"`
	MaxOutputBufferSize    int64         `json:"output_buffer_size"`
	MaxOutputBufferTimeout time.Duration `json:"output_buffer_timeout"`
	ClientTimeout          time.Duration `json:"client_timeout"`
	DeflateEnabled         bool          `json:"deflate"`
	SnappyEnabled          bool          `json:"snappy"`
	TLSMinVersion          uint16        `json:"tls_min_version"`
	LogLevel               string        `json:"log_level"`
}

// Default returns a default config.
func Default() *Config {
	return &Config{
		ServiceName:               "default",
		BusAddr:                   "127.0.0.1:4150",
		LogLevel:                  "info",
		Port:                      8080,
		DevMode:                   false,
		HCMemoryCriticalThreshold: 10.0,
		HCMemoryWarningThreshold:  20.0,
		HCLoadCriticalThreshold:   1.5,
		HCLoadWarningThreshold:    1.0,
		NSQ: NSQSettings{
			TCPAddress:             "127.0.0.1:4150",
			HTTPAddress:            "127.0.0.1:4151",
			BroadcastAddress:       "127.0.0.1",
			AdminHTTPAddress:       "127.0.0.1:4171",
			MemQueueSize:           10000,
			MaxMsgSize:             1024768,
			MsgTimeout:             60 * time.Second,
			SyncEvery:              2500,
			SyncTimeout:            2 * time.Second,
			MaxRdyCount:            2500,
			MaxOutputBufferSize:    64 * 1024,
			MaxOutputBufferTimeout: 250 * time.Millisecond,
			ClientTimeout:          60 * time.Second,
			DeflateEnabled:         true,
			SnappyEnabled:          true,
			TLSMinVersion:          tls.VersionTLS12,
			LogLevel:               "debug",
		},
	}
}

// Load loads config from file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	data = replaceEnvVars(data)

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config json: %w", err)
	}
	return &cfg, nil
}

// LoadFromEnv loads config from environment using prefix.
func LoadFromEnv(prefix string) *Config {
	cfg := Default()

	cfg.ServiceName = getenvStr(prefix+"SERVICE_NAME", cfg.ServiceName)
	cfg.BusAddr = getenvStr(prefix+"BUS_ADDR", cfg.BusAddr)
	cfg.LogLevel = getenvStr(prefix+"LOG_LEVEL", cfg.LogLevel)
	cfg.Port = getenvInt(prefix+"PORT", cfg.Port)
	cfg.DevMode = getenvBool(prefix+"DEV_MODE", cfg.DevMode)
	cfg.HCMemoryCriticalThreshold = getenvFloat(prefix+"HC_MEMORY_CRITICAL", cfg.HCMemoryCriticalThreshold)
	cfg.HCMemoryWarningThreshold = getenvFloat(prefix+"HC_MEMORY_WARNING", cfg.HCMemoryWarningThreshold)
	cfg.HCLoadCriticalThreshold = getenvFloat(prefix+"HC_LOAD_CRITICAL", cfg.HCLoadCriticalThreshold)
	cfg.HCLoadWarningThreshold = getenvFloat(prefix+"HC_LOAD_WARNING", cfg.HCLoadWarningThreshold)

	cfg.NSQ.TCPAddress = getenvStr(prefix+"NSQ_TCP", cfg.NSQ.TCPAddress)
	cfg.NSQ.HTTPAddress = getenvStr(prefix+"NSQ_HTTP", cfg.NSQ.HTTPAddress)
	cfg.NSQ.AdminHTTPAddress = getenvStr(prefix+"NSQ_ADMIN_HTTP", cfg.NSQ.AdminHTTPAddress)
	cfg.NSQ.BroadcastAddress = getenvStr(prefix+"NSQ_BROADCAST", cfg.NSQ.BroadcastAddress)

	return cfg
}

// LoadWithFallback loads from SRV_CONFIG or env vars.
func LoadWithFallback() *Config {
	if path := os.Getenv("SRV_CONFIG"); path != "" {
		if cfg, err := Load(path); err == nil {
			return cfg
		}
	}
	return LoadFromEnv("SRV_")
}

// MustLoadFromEnv panics if config is invalid.
func MustLoadFromEnv() *Config {
	cfg := LoadWithFallback()
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}
	return cfg
}

// Validate checks config for required values.
func (cfg *Config) Validate() error {
	var missing []string
	if cfg.ServiceName == "" {
		missing = append(missing, "service_name")
	}
	if cfg.BusAddr == "" {
		missing = append(missing, "bus_addr")
	}
	if cfg.LogLevel == "" {
		missing = append(missing, "log_level")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		missing = append(missing, fmt.Sprintf("port(%d)", cfg.Port))
	}
	if cfg.NSQ.TCPAddress == "" || cfg.NSQ.HTTPAddress == "" {
		missing = append(missing, "nsq.tcp/nsq.http")
	}
	if len(missing) > 0 {
		return fmt.Errorf("invalid config: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (cfg *Config) String() string {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data)
}

func (cfg *Config) Dump(w io.Writer) {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	_, _ = w.Write(data)
}

// ----------------------------------------------------
// Env helpers
// ----------------------------------------------------

func getenvStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		v = strings.ToLower(v)
		return v == "1" || v == "true" || v == "yes"
	}
	return fallback
}

func getenvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed
		}
	}
	return fallback
}

// replaceEnvVars replaces ${ENV_VAR} in JSON with values from os.Getenv
func replaceEnvVars(data []byte) []byte {
	s := os.Expand(string(data), func(key string) string {
		return os.Getenv(key)
	})
	return []byte(s)
}
