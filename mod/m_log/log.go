package m_log

import (
	"fmt"
	"sync"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// LogModule
//---------------------

// LogModule is responsible for logging actions and subscribing to log events.
type LogModule struct {
	Module       typ.IModule       // The module implementing IModule interface
	client       typ.IBusClient    // Client for publishing logs
	logs         []string          // Slice for storing logs
	mu           sync.Mutex        // Mutex for safe concurrent access to logs
	ConfigClient typ.IConfigClient // Client to fetch log configuration from m_cfg
	logConfig    *x_log.Config     // Log configuration object
}

// Ensure LogModule implements the IModule interface.
var _ typ.IModule = (*LogModule)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the name of the module.
func (m *LogModule) Name() string {
	return m.Module.Name()
}

// Stop stops the module and performs cleanup.
func (m *LogModule) Stop() error {
	// No need for x_log.Sync() in this simplified version
	return nil
}

// Stop stops the module and performs cleanup.
func (m *LogModule) Start() error {
	// No need for x_log.Sync() in this simplified version
	return nil
}

// Stop stops the module and performs cleanup.
func (m *LogModule) Init() error {
	// No need for x_log.Sync() in this simplified version
	return nil
}

//---------------------
// Actions
//---------------------

// Actions returns a list of actions for the logging module.
func (m *LogModule) Actions() []typ.ActionDef {
	return []typ.ActionDef{
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
func (m *LogModule) handleLog(level string) func(typ.IAction) any {
	return func(a typ.IAction) any {
		action, ok := a.(*act.Action)
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
				x_log.Trace(message, context)
			case "debug":
				x_log.Debug(message, context)
			case "info":
				x_log.Info(message, context)
			case "warn":
				x_log.Warn(message, context)
			case "error":
				x_log.Error(message, context)
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
func (m *LogModule) handleGetAllLogs(a typ.IAction) any {
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
func (m *LogModule) initializeLogSubscribers(client typ.IBusClient) {
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
func NewLogModule(service typ.IService, client typ.IBusClient, configClient typ.IConfigClient) *LogModule {
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
	for _, action := range logModule.Actions() {
		act.Register(action.Name, action.Func)
	}

	return logModule
}

// loadLogConfig loads the log configuration from the config client.
func loadLogConfig(configClient typ.IConfigClient) *x_log.Config {
	// Fetch the log configuration from m_cfg using IConfigClient
	logConfig := configClient.GetConfig("log_config").(*x_log.Config)

	// Optionally, configure additional log settings based on the loaded config
	// You can apply the settings from the logConfig object here (log level, format, etc.)

	return logConfig
}

//---------------------
// ILogClient Implementation
//---------------------

// Implement the ILogClient methods for each log level.

func (m *LogModule) Trace(message string) {
	m.handleLog("trace")(&act.Action{
		Inputs: []any{message},
	})
}

func (m *LogModule) Debug(message string) {
	m.handleLog("debug")(&act.Action{
		Inputs: []any{message},
	})
}

func (m *LogModule) Info(message string) {
	m.handleLog("info")(&act.Action{
		Inputs: []any{message},
	})
}

func (m *LogModule) Warn(message string) {
	m.handleLog("warn")(&act.Action{
		Inputs: []any{message},
	})
}

func (m *LogModule) Error(message string) {
	m.handleLog("error")(&act.Action{
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
