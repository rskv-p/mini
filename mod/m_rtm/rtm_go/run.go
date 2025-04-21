package rtm_go

import (
	"fmt"
	"reflect"

	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_rtm/rtm_core"
)

//---------------------
// Go Runtime
//---------------------

// GoRuntime is a runtime that executes Go functions.
type GoRuntime struct{}

// Ensure GoRuntime implements the Runtime interface.
var _ rtm_core.Runtime = (*GoRuntime)(nil)

//---------------------
// Runtime Initialization
//---------------------

// Init prepares the Go runtime (no special initialization needed for Go functions).
func (r *GoRuntime) Init() error {
	// No specific initialization for Go functions is needed.
	//x_log.RootLogger().Structured().Info("GoRuntime initialized")
	return nil
}

//---------------------
// Function Execution
//---------------------

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

	// Log the execution attempt
	//x_log.RootLogger().Structured().Info("Executing Go function", x_log.FString("function", funcName))

	// Execute the Go function by looking it up in a registry or using reflection
	result, err := executeGoFunction(funcName, action)
	if err != nil {
		// Log the error and return it
		//	x_log.RootLogger().Structured().Error("Error executing Go function", x_log.FString("function", funcName), x_log.FError(err))
		return nil, fmt.Errorf("error executing Go function: %w", err)
	}

	// Log the result of the function execution
	//x_log.RootLogger().Structured().Info("Go function executed successfully", x_log.FString("result", fmt.Sprintf("%v", result)))

	return result, nil
}

//---------------------
// Resource Cleanup
//---------------------

// Dispose cleans up the Go runtime (no resources to release in this case).
func (r *GoRuntime) Dispose() {
	// No resources to clean up in Go runtime.
	//x_log.RootLogger().Structured().Info("GoRuntime disposed")
}

//---------------------
// Function Execution Helper
//---------------------

// executeGoFunction is a helper function to find and run a Go function based on its name.
func executeGoFunction(funcName string, action act_type.IAction) (any, error) {
	// Map Go function names to actual Go functions.
	goFuncs := map[string]interface{}{
		"exampleFunction": exampleFunction,
		// Add more function mappings here as needed
	}

	// Look for the function in the map
	fn, exists := goFuncs[funcName]
	if !exists {
		return nil, fmt.Errorf("Go function '%s' not found", funcName)
	}

	// Use reflection to invoke the function dynamically
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("expected function, got %v", fnValue.Kind())
	}

	// Prepare the input parameters
	params := []reflect.Value{reflect.ValueOf(action)}

	// Call the function
	result := fnValue.Call(params)

	// Return the result
	if len(result) > 0 {
		return result[0].Interface(), nil
	}

	return nil, nil
}

//---------------------
// Example Function
//---------------------

// exampleFunction is a simple Go function that could be called through GoRuntime.
func exampleFunction(action act_type.IAction) (any, error) {
	// Example function logic goes here
	return "Executed example function!", nil
}
