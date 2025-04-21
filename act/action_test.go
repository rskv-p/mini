// file: mini/act/action_test.go
package act_test

import (
	"context"
	"testing"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/typ"

	"github.com/stretchr/testify/require"
)

func TestAction_Execute(t *testing.T) {
	//---------------------
	// Setup: Register a simple handler
	//---------------------

	// Register a test handler that processes the action
	act.Register("demo.test", func(a typ.IAction) any {
		// Cast the interface to *act.Action
		action, ok := a.(*act.Action)
		require.True(t, ok) // Ensure the interface is properly cast to *Action

		// Use the Inputs field of the *act.Action structure
		inputs := action.Inputs
		name := "default" // Default name if no input provided
		if len(inputs) > 0 {
			// If the first input is a string, use it as the name
			if str, ok := inputs[0].(string); ok {
				name = str
			}
		}
		// Return a greeting message
		return map[string]any{
			"greeting": "Hello " + name,
		}
	})

	//---------------------
	// Create action
	//---------------------
	// Create a new action with the context and a name input
	action := act.NewAction("demo.test", context.Background(), "Pasha")

	//---------------------
	// Execute the action
	//---------------------
	// Execute the action and check for errors
	err := action.Execute()
	require.NoError(t, err)

	//---------------------
	// Assert the output
	//---------------------
	// Get the output of the action
	val := action.Output()
	require.NotNil(t, val)

	// Assert the output is of the expected type and value
	result, ok := val.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Hello Pasha", result["greeting"])

	//---------------------
	// Dispose the action
	//---------------------
	// Dispose the action and ensure the output is cleared
	action.Dispose()
	require.Nil(t, action.Output())
}
