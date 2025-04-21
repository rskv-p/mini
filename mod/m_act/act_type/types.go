package act_type

import "context"

// IAction defines the core behavior that an action must implement.
type IAction interface {
	// GetName returns the name of the action
	GetName() string

	// WithContext sets and returns the context associated with the action
	WithContext(ctx context.Context) IAction

	// Context returns the context associated with the action
	Context() context.Context

	// Execute runs the action and returns any error encountered
	Execute() (err error)

	// Exec runs the action and returns the result or error
	Exec() (any, error)

	// Run runs the action and returns the result
	Run() any

	// Release releases any resources used by the action
	Release()

	// Dispose clears the inputs and results of the action
	Dispose()

	// Output returns the result of the action after execution
	Output() any

	// String returns a string representation of the action
	String() string

	// ValidateInputsNumber checks if the number of inputs is sufficient
	ValidateInputsNumber(length int) error

	// NumberOfInputs returns the number of inputs to the action
	NumberOfInputs() int

	// NumberOfInputsIs checks if the number of inputs matches the specified number
	NumberOfInputsIs(num int) bool

	// InputNotNull ensures that a specific input is not nil
	InputNotNull(i int) error

	// InputString parses a string input at the specified index
	InputString(i int, defaults ...string) string

	// InputInt parses an integer input at the specified index
	InputInt(i int, defaults ...int) int

	// InputUint32 parses a uint32 input at the specified index
	InputUint32(i int, defaults ...uint32) uint32

	// InputBool parses a boolean input at the specified index
	InputBool(i int, defaults ...bool) bool

	// InputURL parses a URL input at the specified index and retrieves a specific key
	InputURL(i int, key string, defaults ...string) string

	// InputMap parses an input as a map
	InputMap(i int) map[string]any

	// InputArray parses an input as an array
	InputArray(i int) []any

	// InputStrings parses an input as an array of strings
	InputStrings(i int) []string

	// InputsRecords parses an input as an array of records (maps)
	InputsRecords(i int) []map[string]any
}

//---------------------
// Action and Handler
//---------------------

// Handler defines the function signature for handlers that process actions.
type Handler func(IAction) any

// ActionDef represents a definition of an action, including its name and handler.
type ActionDef struct {
	Name   string  // Action name
	Func   Handler // Function to handle the action
	Public bool    // Future use (e.g., to expose the action via an API)
}
