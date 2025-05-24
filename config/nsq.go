package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"time"
)

// NSQSettings defines settings for embedded or remote NSQ.
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

// DefaultNSQ returns recommended NSQ settings.
func DefaultNSQ() NSQSettings {
	return NSQSettings{
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
	}
}

// Validate checks required NSQ values.
func (nsq NSQSettings) Validate() error {
	var missing []string
	if nsq.TCPAddress == "" {
		missing = append(missing, "tcp")
	}
	if nsq.HTTPAddress == "" {
		missing = append(missing, "http")
	}
	if len(missing) > 0 {
		return errors.New("invalid NSQ config: " + fmt.Sprintf("%v", missing))
	}
	return nil
}
