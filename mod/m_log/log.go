package x_mod

import (
	"fmt"
	"sync"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// Log Module
//---------------------

// LogModule handles logging actions and subscribes to log events.
type LogModule struct {
	client typ.IClient // Client used to publish logs
	logs   []string    // Store the logs
	mu     sync.Mutex  // Mutex to ensure safe concurrent access to logs
}

// NewLogModule creates a new LogModule and subscribes to log events.
func NewLogModule(client typ.IClient) typ.IModule {
	module := &LogModule{
		client: client,
	}
	module.initializeLogSubscribers(client) // Initialize log subscribers
	return &mod.Module{
		ModName: "log",
		Acts: []typ.ActionDef{
			// Define log actions for different levels
			{Name: "log.trace", Func: module.handleLog("trace"), Public: true},
			{Name: "log.debug", Func: module.handleLog("debug"), Public: true},
			{Name: "log.info", Func: module.handleLog("info"), Public: true},
			{Name: "log.warn", Func: module.handleLog("warn"), Public: true},
			{Name: "log.error", Func: module.handleLog("error"), Public: true},
			{Name: "log.get_all", Func: module.handleGetAllLogs, Public: true},
		},
		OnInit: func() error {
			// Initialize log level from configuration
			cfg := x_log.LoadConfigFromEnv()
			x_log.SetLogLevel(x_log.RootLogger(), cfg.Level)
			return nil
		},
		OnStop: func() error {
			// Synchronize logs on stop
			x_log.Sync()
			return nil
		},
	}
}

//---------------------
// Log Handlers
//---------------------

// handleLog handles log actions based on the log level.
func (m *LogModule) handleLog(level string) func(typ.IAction) any {
	return func(a typ.IAction) any {
		action, ok := a.(*act.Action)
		if !ok {
			return "Invalid action type"
		}

		// Format the log message
		message := fmt.Sprintf("[%s] %v", level, action.Inputs)
		// Log the message based on the level
		switch level {
		case "trace":
			x_log.Trace(action.Inputs...)
		case "debug":
			x_log.Debug(action.Inputs...)
		case "info":
			x_log.Info(action.Inputs...)
		case "warn":
			x_log.Warn(action.Inputs...)
		case "error":
			x_log.Error(action.Inputs...)
		}
		// Store the log
		m.addLog(message)

		// Publish the log message through the client
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

//---------------------
// Log Management
//---------------------

// addLog appends a new log entry to the logs slice.
func (m *LogModule) addLog(log string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, log)
}

// initializeLogSubscribers subscribes to real-time log updates.
func (m *LogModule) initializeLogSubscribers(client typ.IClient) {
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
// ILogClient Interface Methods
//---------------------

// Implement ILogClient interface methods for each log level

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
