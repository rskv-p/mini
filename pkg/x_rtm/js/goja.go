package js

import (
	"errors"
	"fmt"

	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/pkg/x_rtm"
	"github.com/rskv-p/mini/typ"

	"github.com/dop251/goja"
)

var ErrNoScript = errors.New("js_runtime: no script provided in Inputs[0]")

// JSRuntime is a JavaScript runtime based on Goja.
type JSRuntime struct {
	vm *goja.Runtime
}

// Ensure JSRuntime implements the Runtime interface.
var _ x_rtm.Runtime = (*JSRuntime)(nil)

// Init prepares the JS VM and logs the initialization.
func (r *JSRuntime) Init() error {
	r.vm = goja.New()
	x_log.RootLogger().Structured().Info("JSRuntime initialized")
	return nil
}

// Execute runs JS code from action.Inputs[0] and returns the result or error.
func (r *JSRuntime) Execute(action typ.IAction) (any, error) {
	if action == nil {
		return nil, fmt.Errorf("js_runtime: action cannot be nil")
	}

	// Check if action contains the script
	if !action.NumberOfInputsIs(1) {
		return nil, ErrNoScript
	}

	code := action.InputString(0)
	if code == "" {
		return nil, ErrNoScript
	}

	// Log the execution of the code
	x_log.RootLogger().Structured().Info("Executing JS code", x_log.FString("code", code))

	// Run the JavaScript code using Goja runtime
	val, err := r.vm.RunString(code)
	if err != nil {
		x_log.RootLogger().Structured().Error("Error executing JS code", x_log.FString("code", code), x_log.FError(err))
		return nil, fmt.Errorf("js_runtime: execution error: %w", err)
	}

	// Log the successful execution
	x_log.RootLogger().Structured().Info("JS code executed successfully", x_log.FString("result", fmt.Sprintf("%v", val.Export())))

	return val.Export(), nil
}

// Dispose clears the JS VM and logs the disposal.
func (r *JSRuntime) Dispose() {
	if r.vm != nil {
		// Dispose of the JS runtime
		r.vm = nil
		x_log.RootLogger().Structured().Info("JSRuntime disposed")
	}
}
