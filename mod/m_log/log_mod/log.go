package log_mod

import (
	"fmt"
	"sync"

	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/mod/m_act/act_core"
	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_bus/bus_type"
	"github.com/rskv-p/mini/mod/m_cfg/cfg_type"
	"github.com/rskv-p/mini/mod/m_log/log_core"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// LogModule
//---------------------

// LogModule is responsible for logging actions and subscribing to log events.
type LogModule struct {
	Module       typ.IModule            // The module implementing IModule interface
	client       bus_type.IBusClient    // Client for publishing logs
	logs         []string               // Slice for storing logs
	mu           sync.Mutex             // Mutex for safe concurrent access to logs
	ConfigClient cfg_type.IConfigClient // Client to fetch log configuration from m_cfg
	logConfig    *log_core.Config       // Log configuration object
}

// Ensure LogModule implements the IModule interface.
var _ typ.IModule = (*LogModule)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the name of the module.
func (m *LogModule) GetName() string {
	return m.Module.GetName()
}

// Stop stops the module and performs cleanup.
func (m *LogModule) Stop() error {
	// No need for log_core.Sync() in this simplified version
	return nil
}

// Stop stops the module and performs cleanup.
func (m *LogModule) Start() error {
	// No need for log_core.Sync() in this simplified version
	return nil
}

// Stop stops the module and performs cleanup.
func (m *LogModule) Init() error {
	// No need for log_core.Sync() in this simplified version
	return nil
}

//---------------------
// Actions
//---------------------

// Actions returns a list of actions for the logging module.
func (m *LogModule) GetActions() []act_type.ActionDef {
	return []act_type.ActionDef{
		// Actions for logging at different levels
		{Name: "log.trace", Func: m.handleLog("trace"), Public: true},
		{Name: "log.debug", Func: m.handleLog("debug"), Public: true},
		{Name: "log.info", Func: m.handleLog("info"), Public: true},
		{Name: "log.warn", Func: m.handleLog("warn"), Public: true},
		{Name: "log.error", Func: m.handleLog("error"), Public: true},
		{Name: "log.get_all", Func: m.handleGetAllLogs, Public: true},
	}
}

//---------------------
// Log Handling
//---------------------

// handleLog handles logging actions based on the log level.
func (m *LogModule) handleLog(level string) func(act_type.IAction) any {
	return func(a act_type.IAction) any {
		action, ok := a.(*act_core.Action)
		if !ok {
			return "Invalid action type"
		}

		// Format the log message
		message := fmt.Sprintf("[%s] %v", level, action.Inputs)

		// Prepare the context for the log (currently empty, can be expanded)
		context := map[string]interface{}{
			"inputs": action.Inputs, // You can add more context if needed
		}

		// Check if the log level should be logged based on the current config
		if shouldLog(m.logConfig.Level, level) {
			// Log the message based on the level, with context
			switch level {
			case "trace":
				log_core.Trace(message, context)
			case "debug":
				log_core.Debug(message, context)
			case "info":
				log_core.Info(message, context)
			case "warn":
				log_core.Warn(message, context)
			case "error":
				log_core.Error(message, context)
			}
		}

		// Add the log to memory
		m.addLog(message)

		// Publish the log message via the client
		m.client.Publish(fmt.Sprintf("logs.%s", level), []byte(message))

		return nil
	}
}

// handleGetAllLogs returns all collected logs.
func (m *LogModule) handleGetAllLogs(a act_type.IAction) any {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]any{"logs": m.logs}
}

// addLog adds a new log entry to the slice.
func (m *LogModule) addLog(log string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, log)
}

//---------------------
// Log Subscribers
//---------------------

// initializeLogSubscribers subscribes to real-time log updates.
func (m *LogModule) initializeLogSubscribers(client bus_type.IBusClient) {
	// Subscribe to various log levels for real-time updates
	client.SubscribeWithQueue("logs.trace", "", func(subject string, msg []byte) {
		m.addLog(fmt.Sprintf("[TRACE] %s: %s", subject, string(msg)))
	})

	client.SubscribeWithQueue("logs.debug", "", func(subject string, msg []byte) {
		m.addLog(fmt.Sprintf("[DEBUG] %s: %s", subject, string(msg)))
	})

	client.SubscribeWithQueue("logs.info", "", func(subject string, msg []byte) {
		m.addLog(fmt.Sprintf("[INFO] %s: %s", subject, string(msg)))
	})

	client.SubscribeWithQueue("logs.warn", "", func(subject string, msg []byte) {
		m.addLog(fmt.Sprintf("[WARN] %s: %s", subject, string(msg)))
	})

	client.SubscribeWithQueue("logs.error", "", func(subject string, msg []byte) {
		m.addLog(fmt.Sprintf("[ERROR] %s: %s", subject, string(msg)))
	})
}

//---------------------
// Module Creation
//---------------------

// NewLogModule creates a new instance of LogModule and subscribes to log events.
func NewLogModule(service typ.IService, client bus_type.IBusClient, configClient cfg_type.IConfigClient) *LogModule {
	// Create a new module using NewModule
	module := mod.NewModule("log", service, nil, nil, nil)

	// Fetch log configuration from m_cfg
	logConfig := loadLogConfig(configClient)

	// Return the LogModule with the created module and log configuration
	logModule := &LogModule{
		Module:       module,
		client:       client,
		ConfigClient: configClient, // Store the config client to access other settings if needed
		logConfig:    logConfig,
	}

	// Initialize log subscribers
	logModule.initializeLogSubscribers(client)

	// Register actions for the module
	for _, action := range logModule.GetActions() {
		act_core.Register(action.Name, action.Func)
	}

	return logModule
}

// loadLogConfig loads the log configuration from the config client.
func loadLogConfig(configClient cfg_type.IConfigClient) *log_core.Config {
	// Fetch the log configuration from m_cfg using IConfigClient
	logConfig := configClient.GetConfig("log_config")
	if logConfig == nil {
		// If the config is nil, return a default configuration
		fmt.Println("Log configuration not found, using default configuration")
		return &log_core.Config{
			Level:  "DEBUG",   // Default log level
			Format: "console", // Default log format
		}
	}

	// Attempt to type assert the config value into the expected type
	config, ok := logConfig.(*log_core.Config)
	if !ok {
		fmt.Println("Invalid log config type, using default configuration")
		return &log_core.Config{
			Level:  "DEBUG",   // Default log level
			Format: "console", // Default log format
		}
	}

	// Return the valid configuration
	return config
}

//---------------------
// ILogClient Implementation
//---------------------

// Implement the ILogClient methods for each log level.

func (m *LogModule) Trace(message string) {
	m.handleLog("trace")(&act_core.Action{
		Inputs: []any{message},
	})
}

func (m *LogModule) Debug(message string) {
	m.handleLog("debug")(&act_core.Action{
		Inputs: []any{message},
	})
}

func (m *LogModule) Info(message string) {
	m.handleLog("info")(&act_core.Action{
		Inputs: []any{message},
	})
}

func (m *LogModule) Warn(message string) {
	m.handleLog("warn")(&act_core.Action{
		Inputs: []any{message},
	})
}

func (m *LogModule) Error(message string) {
	m.handleLog("error")(&act_core.Action{
		Inputs: []any{message},
	})
}

//---------------------
// Helper Function
//---------------------

// shouldLog checks if the log level should be logged based on the configured log level.
func shouldLog(configuredLevel, logLevel string) bool {
	levels := map[string]int{
		"TRACE": 0,
		"DEBUG": 1,
		"INFO":  2,
		"WARN":  3,
		"ERROR": 4,
	}

	// Compare the configured log level with the log level of the current message
	return levels[logLevel] >= levels[configuredLevel]
}
