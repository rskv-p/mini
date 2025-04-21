package rtm_mod

import (
	"fmt"

	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/mod/m_act/act_core"
	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_log/log_type"
	"github.com/rskv-p/mini/mod/m_rtm/rtm_core"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// RuntimeModule
//---------------------

// RuntimeModule manages various runtimes (JS, exec, etc.) and executes actions.
type RuntimeModule struct {
	Module    typ.IModule                 // The module implementing IModule interface
	runtimes  map[string]rtm_core.Runtime // Map of registered runtimes
	logClient log_type.ILogClient         // Client for logging
}

// Ensure RuntimeModule implements the IModule interface.
var _ typ.IModule = (*RuntimeModule)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the name of the module.
func (m *RuntimeModule) GetName() string {
	return m.Module.GetName()
}

// Start starts the module (specific logic can be added).
func (s *RuntimeModule) Start() error {
	// Log module start
	s.logClient.Info("Starting Runtime module", map[string]interface{}{"module": s.GetName()})
	return nil
}

// Stop stops the module and cleans up all resources.
func (m *RuntimeModule) Stop() error {
	// Log module stop
	m.logClient.Info("Stopping Runtime module", map[string]interface{}{"module": m.GetName()})

	// Stop and dispose of all registered runtimes
	for name, runtime := range m.runtimes {
		runtime.Dispose()
		// Log stopping of each runtime
		m.logClient.Info(fmt.Sprintf("Disposed runtime: %s", name), map[string]interface{}{"runtime": name})
	}
	return nil
}

//---------------------
// Actions
//---------------------

// Actions returns a list of actions the runtime module can perform.
func (m *RuntimeModule) GetActions() []act_type.ActionDef {
	return []act_type.ActionDef{
		{
			Name:   "runtime.execute",     // Execute action in the specified runtime
			Func:   m.handleExecuteAction, // Pass method reference for the handler
			Public: true,
		},
	}
}

//---------------------
// Action Handlers
//---------------------

// handleExecuteAction executes an action in the specified runtime.
func (m *RuntimeModule) handleExecuteAction(a act_type.IAction) any {
	action, ok := a.(*act_core.Action)
	if !ok {
		m.logClient.Error("Invalid action type", map[string]interface{}{"error": "invalid action type"})
		return fmt.Errorf("invalid action type")
	}

	runtimeName := action.InputString(0)

	// Log the action before execution
	m.logClient.Info(fmt.Sprintf("Executing action in runtime: %s", runtimeName), map[string]interface{}{"runtime": runtimeName})

	// Execute the action using the specified runtime
	return m.ExecuteAction(runtimeName, action)
}

//---------------------
// Runtime Registration and Execution
//---------------------

// RegisterRuntime adds a runtime to the module.
func (m *RuntimeModule) RegisterRuntime(name string, runtime rtm_core.Runtime) {
	m.runtimes[name] = runtime
	// Log the registration of a runtime
	m.logClient.Info(fmt.Sprintf("Registered runtime: %s", name), map[string]interface{}{"runtime": name})
}

// ExecuteAction executes the specified action in the given runtime.
func (m *RuntimeModule) ExecuteAction(runtimeName string, action *act_core.Action) any {
	runtime, ok := m.runtimes[runtimeName]
	if !ok {
		m.logClient.Error(fmt.Sprintf("Runtime not found: %s", runtimeName), map[string]interface{}{"runtime": runtimeName})
		return fmt.Errorf("runtime not found: %s", runtimeName)
	}

	// Execute the action using the selected runtime
	result, err := runtime.Execute(action)
	if err != nil {
		m.logClient.Error(fmt.Sprintf("Error executing action in runtime %s", runtimeName), map[string]interface{}{"runtime": runtimeName, "error": err})
		return fmt.Errorf("error executing action: %v", err)
	}

	// Log success
	m.logClient.Info(fmt.Sprintf("Successfully executed action in runtime: %s", runtimeName), map[string]interface{}{"runtime": runtimeName})
	return result
}

//---------------------
// Module Initialization
//---------------------

// Init initializes the module and registers available runtimes.
func (m *RuntimeModule) Init() error {
	// Register available runtimes
	m.RegisterRuntime("exec", &rtm_core.ExecRuntime{Client: &rtm_core.Client{}})
	m.RegisterRuntime("js", &rtm_core.JSRuntime{Client: &rtm_core.Client{}})
	return nil
}

//---------------------
// Module Creation
//---------------------

// NewRuntimeModule creates a new instance of RuntimeModule using NewModule constructor.
func NewRuntimeModule(service typ.IService, logClient log_type.ILogClient) *RuntimeModule {
	// Create a new module using NewModule, passing name, service, actions, and nil for OnInit and OnStop
	module := mod.NewModule("runtime", service, nil, nil, nil)

	// Log module creation
	logClient.Info("Creating Runtime module", map[string]interface{}{"module": "runtime"})

	// Return the RuntimeModule with the created module
	runtimeModule := &RuntimeModule{
		Module:    module,
		runtimes:  make(map[string]rtm_core.Runtime),
		logClient: logClient,
	}

	// Register actions for the module
	for _, action := range runtimeModule.GetActions() {
		act_core.Register(action.Name, action.Func)
	}

	// Log successful creation of module
	logClient.Info("Runtime module created successfully", map[string]interface{}{"module": "runtime"})

	return runtimeModule
}
