// file: arc/service/constant/constant.go
package constant

import "errors"

// ----------------------------------------------------
// Standard errors
// ----------------------------------------------------

var (
	ErrBadRequest       = errors.New("invalid request")
	ErrNotFound         = errors.New("resource not found")
	ErrEmptyMessage     = errors.New("message cannot be empty")
	ErrMissingHandler   = errors.New("no handler registered for topic")
	ErrNoRegistry       = errors.New("registry is empty")
	ErrNoAvailableNodes = errors.New("no service instance available")
	ErrEmptyNodeList    = errors.New("registry requires at least one node")
	ErrInvalidPath      = errors.New("registry requires at least one node")
)

// ----------------------------------------------------
// Config paths & keys
// ----------------------------------------------------

const (
	DefaultURL                      = "bus://127.0.0.1:4222"
	DefaultConfigFile               = "config.json"
	GogoConfigPath                  = "/etc/gogo/"
	ConfigFileName                  = ".config.json"
	MaxFileChunkSize                = 2 * 1024 * 1024 // 2MB
	ConfigBusAddress                = "bus_address"
	ConfigHCMemoryCriticalThreshold = "memory_critical"
	ConfigHCMemoryWarningThreshold  = "memory_warning"
	ConfigHCLoadCriticalThreshold   = "load_critical"
	ConfigHCLoadWarningThreshold    = "load_warning"
)

// ----------------------------------------------------
// Message types
// ----------------------------------------------------

const (
	MessageTypeRequest     = "request"
	MessageTypeResponse    = "response"
	MessageTypePublish     = "publish"
	MessageTypeHealthCheck = "healthCheck"
	MessageTypeStream      = "stream"
	MessageTypeEvent       = "event"
)

// ----------------------------------------------------
// Health check status
// ----------------------------------------------------

const (
	HealthOK       = 0
	HealthWarning  = 1
	HealthCritical = 2
)

// ----------------------------------------------------
// Body keys (standardized)
// ----------------------------------------------------

const (
	BodyKeyError  = "error"
	BodyKeyResult = "result"
	BodyKeyInput  = "input"
)

// ----------------------------------------------------
// Action / invoke keys
// ----------------------------------------------------

const (
	KeyActionID   = "id"
	KeyActionName = "action"
)

// ----------------------------------------------------
// Status flags (internal logic)
// ----------------------------------------------------

const (
	StatusOK       = 0
	StatusWarning  = 1
	StatusCritical = 2
)

// ----------------------------------------------------
// HTTP status codes (for compatibility / future use)
// ----------------------------------------------------

const (
	StatusBadRequest    = 400
	StatusNotFound      = 404
	StatusInternalError = 500
	StatusTimeout       = 504
)

// ----------------------------------------------------
// Misc
// ----------------------------------------------------

const OrganizationName = "run"

const (
	// MemoryWarningKey is the config key for memory warnings.
	MemoryWarningKey = "memory_warning"
	// MemoryCriticalKey is the config key for critical memory.
	MemoryCriticalKey = "memory_critical"
	// LoadWarningKey is the config key for load warnings.
	LoadWarningKey = "load_warning"
	// LoadCriticalKey is the config key for critical load.
	LoadCriticalKey = "load_critical"
)
