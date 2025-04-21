package mock

import (
	"fmt"

	"github.com/rskv-p/mini/pkg/x_rtm"
	"github.com/rskv-p/mini/typ"
)

// MockRuntime is a fake runtime used for testing or fallback execution.
type MockRuntime struct {
	Initialized bool        // Whether Init() was called
	Disposed    bool        // Whether Dispose() was called
	Result      any         // Optional result to return from Execute()
	Err         error       // Optional error to return from Execute()
	CalledWith  typ.IAction // Last action passed to Execute()
}

// Ensure interface compliance
var _ x_rtm.Runtime = (*MockRuntime)(nil)

// Init initializes the MockRuntime and logs the initialization.
func (m *MockRuntime) Init() error {
	m.Initialized = true
	//	x_log.RootLogger().Structured().Info("MockRuntime initialized")
	return nil
}

// Execute runs the action and returns the pre-defined result or error.
// Logs the action being executed for better tracking in tests.
func (m *MockRuntime) Execute(action typ.IAction) (any, error) {
	m.CalledWith = action
	///x_log.RootLogger().Structured().Info("Executing action", x_log.FString("action_name", action.GetName()))

	if m.Err != nil {
		//	x_log.RootLogger().Structured().Error("MockRuntime execution failed", x_log.FError(m.Err))
		return nil, m.Err
	}

	if m.Result != nil {
		//		x_log.RootLogger().Structured().Info("MockRuntime execution successful", x_log.FString("result", fmt.Sprintf("%v", m.Result)))
		return m.Result, nil
	}

	// Default mocked result if no result is defined
	mockedResult := fmt.Sprintf("mocked result for %s", action.GetName())
	//	x_log.RootLogger().Structured().Info("Returning default mocked result", x_log.FString("result", mockedResult))
	return mockedResult, nil
}

// Dispose disposes the MockRuntime and logs the disposal.
func (m *MockRuntime) Dispose() {
	m.Disposed = true
	//x_log.RootLogger().Structured().Info("MockRuntime disposed")
}
