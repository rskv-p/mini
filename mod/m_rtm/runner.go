package m_rtm

import (
	"fmt"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/pkg/x_rtm"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// RuntimeModule
//---------------------

// RuntimeModule manages various runtimes (JS, exec, etc.) and executes actions.
type RuntimeModule struct {
	Module    typ.IModule              // The module implementing IModule interface
	runtimes  map[string]x_rtm.Runtime // Map of registered runtimes
	logClient typ.ILogClient           // Client for logging
}

// Ensure RuntimeModule implements the IModule interface.
var _ typ.IModule = (*RuntimeModule)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the name of the module.
func (m *RuntimeModule) Name() string {
	return m.Module.Name()
}

// Start starts the module (specific logic can be added).
func (s *RuntimeModule) Start() error {
	// Log module start
	s.logClient.Info("Starting Runtime module", map[string]interface{}{"module": s.Name()})
	return nil
}

// Stop stops the module and cleans up all resources.
func (m *RuntimeModule) Stop() error {
	// Log module stop
	m.logClient.Info("Stopping Runtime module", map[string]interface{}{"module": m.Name()})

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
func (m *RuntimeModule) Actions() []typ.ActionDef {
	return []typ.ActionDef{
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
func (m *RuntimeModule) handleExecuteAction(a typ.IAction) any {
	action, ok := a.(*act.Action)
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
func (m *RuntimeModule) RegisterRuntime(name string, runtime x_rtm.Runtime) {
	m.runtimes[name] = runtime
	// Log the registration of a runtime
	m.logClient.Info(fmt.Sprintf("Registered runtime: %s", name), map[string]interface{}{"runtime": name})
}

// ExecuteAction executes the specified action in the given runtime.
func (m *RuntimeModule) ExecuteAction(runtimeName string, action *act.Action) any {
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
	m.RegisterRuntime("exec", &x_rtm.ExecRuntime{Client: &x_rtm.Client{}})
	m.RegisterRuntime("js", &x_rtm.JSRuntime{Client: &x_rtm.Client{}})
	return nil
}

//---------------------
// Module Creation
//---------------------

// NewRuntimeModule creates a new instance of RuntimeModule using NewModule constructor.
func NewRuntimeModule(service typ.IService, logClient typ.ILogClient) *RuntimeModule {
	// Create a new module using NewModule, passing name, service, actions, and nil for OnInit and OnStop
	module := mod.NewModule("runtime", service, nil, nil, nil)

	// Log module creation
	logClient.Info("Creating Runtime module", map[string]interface{}{"module": "runtime"})

	// Return the RuntimeModule with the created module
	runtimeModule := &RuntimeModule{
		Module:    module,
		runtimes:  make(map[string]x_rtm.Runtime),
		logClient: logClient,
	}

	// Register actions for the module
	for _, action := range runtimeModule.Actions() {
		act.Register(action.Name, action.Func)
	}

	// Log successful creation of module
	logClient.Info("Runtime module created successfully", map[string]interface{}{"module": "runtime"})

	return runtimeModule
}
