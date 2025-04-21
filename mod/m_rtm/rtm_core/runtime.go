package rtm_core

import (
	"errors"
	"fmt"

	"github.com/rskv-p/mini/mod/m_act/act_core"
	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_rtm/rtm_type"
)

// Runtime defines a runtime that can execute actions in a specific environment (JS, exec, etc.).
type Runtime interface {
	Init() error                                  // Initialize the runtime environment.
	Execute(action act_type.IAction) (any, error) // Execute the action in the runtime.
	Dispose()                                     // Clean up the resources used by the runtime.
}

// Client implements the RuntimeClient interface to execute actions in different runtimes.
type Client struct{}

// ExecuteAction executes the action in a given runtime.
func (c *Client) ExecuteAction(action act_type.IAction) (any, error) {
	// Placeholder for future implementation
	return fmt.Sprintf("Executed action: %s", action.GetName()), nil
}

// JSRuntime is a JavaScript runtime based on Goja.
type JSRuntime struct {
	Client rtm_type.RuntimeClient
}

// Init initializes the JS runtime environment.
func (r *JSRuntime) Init() error {
	// Initialize the JS runtime environment (e.g., setting up the JS interpreter).
	return nil
}

// Execute runs the action in the JS runtime.
func (r *JSRuntime) Execute(action act_type.IAction) (any, error) {
	// Use the client to execute the action (e.g., running JS code)
	if r.Client == nil {
		return nil, errors.New("JSRuntime client is not initialized")
	}
	return r.Client.ExecuteAction(action)
}

// Dispose disposes of JS runtime resources.
func (r *JSRuntime) Dispose() {
	// Dispose of JS runtime resources (e.g., clearing JS environment).
}

// ExecRuntime implements the Runtime interface to execute system commands (e.g., shell commands).
type ExecRuntime struct {
	Client rtm_type.RuntimeClient
}

// Init initializes the Exec runtime environment.
func (r *ExecRuntime) Init() error {
	// Initialize Exec environment (e.g., setting up necessary configurations for system commands).
	return nil
}

// Execute runs the action in the Exec runtime.
func (r *ExecRuntime) Execute(action act_type.IAction) (any, error) {
	// Use the client to execute the action (e.g., executing a system command)
	if r.Client == nil {
		return nil, errors.New("ExecRuntime client is not initialized")
	}
	return r.Client.ExecuteAction(action)
}

// Dispose disposes of Exec runtime resources.
func (r *ExecRuntime) Dispose() {
	// Clean up Exec runtime resources (e.g., clearing any system-related resources).
}

// GoRuntime is a Go runtime for executing Go functions.
type GoRuntime struct{}

// Init prepares the Go runtime (no special initialization needed for Go functions).
func (r *GoRuntime) Init() error {
	// No special initialization for Go functions
	return nil
}

// Execute runs a Go function wrapped in an action.
func (r *GoRuntime) Execute(action act_type.IAction) (any, error) {
	// Check if the action has at least one input (the function name)
	if !action.NumberOfInputsIs(1) {
		return nil, fmt.Errorf("no Go function provided in Inputs[0]")
	}

	// The first input should be the function name or type
	funcName := action.InputString(0)
	if funcName == "" {
		return nil, fmt.Errorf("invalid Go function name")
	}

	// Find and execute the Go function (simple lookup in a registry)
	result, err := executeGoFunction(funcName, action)
	if err != nil {
		return nil, fmt.Errorf("error executing Go function: %w", err)
	}

	return result, nil
}

// Dispose cleans up the Go runtime (no resources to release in this case).
func (r *GoRuntime) Dispose() {
	// No resources to clean up in Go runtime.
}

// executeGoFunction is a helper function to find and run a Go function based on its name.
func executeGoFunction(funcName string, action act_type.IAction) (any, error) {
	// Here we can map Go function names to actual Go function calls.
	goFuncs := map[string]func(act_type.IAction) (any, error){
		"exampleFunction": exampleFunction,
	}

	// Look for the function in the map
	fn, exists := goFuncs[funcName]
	if !exists {
		return nil, fmt.Errorf("Go function '%s' not found", funcName)
	}

	// Execute the function
	return fn(action)
}

// exampleFunction is a simple Go function that could be called through GoRuntime.
func exampleFunction(action act_type.IAction) (any, error) {
	// Example function logic goes here
	return "Executed example function!", nil
}

// RuntimeModule manages different runtimes (JS, Exec, Go, etc.) and executes actions in the appropriate runtime.
type RuntimeModule struct {
	runtimes map[string]Runtime
}

// NewRuntimeModule creates a new RuntimeModule instance.
func NewRuntimeModule() *RuntimeModule {
	return &RuntimeModule{
		runtimes: make(map[string]Runtime),
	}
}

// RegisterRuntime adds a new runtime to the module.
func (m *RuntimeModule) RegisterRuntime(name string, runtime Runtime) {
	m.runtimes[name] = runtime
}

// ExecuteAction executes the given action in the specified runtime.
func (m *RuntimeModule) ExecuteAction(runtimeName string, action *act_core.Action) (any, error) {
	runtime, ok := m.runtimes[runtimeName]
	if !ok {
		return nil, fmt.Errorf("runtime not found: %s", runtimeName)
	}

	// Execute the action using the selected runtime
	result, err := runtime.Execute(action)
	if err != nil {
		return nil, fmt.Errorf("error executing action: %v", err)
	}

	return result, nil
}

// Actions returns a list of actions available in the RuntimeModule.
func (m *RuntimeModule) GetActions() []act_type.ActionDef {
	return []act_type.ActionDef{
		{
			Name: "runtime.execute",
			Func: func(a act_type.IAction) any {
				action, ok := a.(*act_core.Action)
				if !ok {
					return fmt.Errorf("invalid action type")
				}

				runtimeName := action.InputString(0)
				result, err := m.ExecuteAction(runtimeName, action)
				if err != nil {
					return fmt.Errorf("execution failed: %v", err)
				}
				return result
			},
			Public: true,
		},
	}
}

// Init initializes the RuntimeModule by registering available runtimes.
func (m *RuntimeModule) Init() error {
	// Register all the available runtimes
	m.RegisterRuntime("exec", &ExecRuntime{Client: &Client{}})
	m.RegisterRuntime("js", &JSRuntime{Client: &Client{}})
	m.RegisterRuntime("go", &GoRuntime{})
	return nil
}

// Stop stops all runtimes and releases their resources.
func (m *RuntimeModule) Stop() error {
	for _, runtime := range m.runtimes {
		runtime.Dispose()
	}
	return nil
}

// Name returns the name of the RuntimeModule.
func (m *RuntimeModule) GetName() string {
	return "runtime"
}
