package m_rtm

import (
	"fmt"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/pkg/x_rtm"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// Runtime Module
//---------------------

// RuntimeModule is responsible for managing different runtimes (JS, exec, etc.) and executing actions.
type RuntimeModule struct {
	runtimes map[string]x_rtm.Runtime // Map of registered runtimes
}

//---------------------
// Module Lifecycle
//---------------------

// NewRuntimeModule creates a new RuntimeModule instance.
func NewRuntimeModule() *RuntimeModule {
	return &RuntimeModule{
		runtimes: make(map[string]x_rtm.Runtime),
	}
}

// RegisterRuntime adds a runtime to the module.
func (m *RuntimeModule) RegisterRuntime(name string, runtime x_rtm.Runtime) {
	m.runtimes[name] = runtime
}

// ExecuteAction executes the given action in the specified runtime.
func (m *RuntimeModule) ExecuteAction(runtimeName string, action *act.Action) any {
	runtime, ok := m.runtimes[runtimeName]
	if !ok {
		return fmt.Errorf("runtime not found: %s", runtimeName)
	}

	// Execute the action using the selected runtime
	result, err := runtime.Execute(action)
	if err != nil {
		return fmt.Errorf("error executing action: %v", err)
	}

	return result
}

// Actions returns a list of actions available in the RuntimeModule.
func (m *RuntimeModule) Actions() []typ.ActionDef {
	return []typ.ActionDef{
		{
			Name: "runtime.execute", // Execute action in the specified runtime
			Func: func(a typ.IAction) any {
				action, ok := a.(*act.Action)
				if !ok {
					return fmt.Errorf("invalid action type")
				}

				runtimeName := action.InputString(0)
				return m.ExecuteAction(runtimeName, action)
			},
			Public: true,
		},
	}
}

//---------------------
// Module Initialization and Cleanup
//---------------------

// Init initializes the module and registers available runtimes.
func (m *RuntimeModule) Init() error {
	// Register the available runtimes
	m.RegisterRuntime("exec", &x_rtm.ExecRuntime{Client: &x_rtm.Client{}})
	m.RegisterRuntime("js", &x_rtm.JSRuntime{Client: &x_rtm.Client{}})
	return nil
}

// Stop stops the module and disposes of all runtimes.
func (m *RuntimeModule) Stop() error {
	// Dispose all registered runtimes
	for _, runtime := range m.runtimes {
		runtime.Dispose()
	}
	return nil
}

// Name returns the name of the module.
func (m *RuntimeModule) Name() string {
	return "runtime"
}
