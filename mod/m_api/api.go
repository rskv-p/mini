package m_api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// API Registration
//---------------------

// RegisterAction registers a handler in x_action,
// and exposes it as a public API if Public option is set.
func RegisterAction(name string, handler typ.Handler, opts APIOption) error {
	// Register in x_action
	act.Register(name, handler)

	// Wrap as public API endpoint if Public option is set
	if opts.Public {
		return RegisterAPI(name, handler, opts)
	}
	return nil
}

// RegisterAPI registers a public API endpoint for the action.
func RegisterAPI(name string, handler typ.Handler, opts APIOption) error {
	// Ensure we register the action properly as a public API
	Register(name, func(args []any) (any, error) {
		// Create a new action with context
		ctx := context.Background()
		action := act.NewAction(name, ctx, args...) // Pass context as the first argument

		// Execute the action
		err := action.Execute()
		if err != nil {
			// Log the error and return it
			log.Printf("Error executing action '%s': %v", name, err)
			return nil, fmt.Errorf("failed to execute action: %w", err)
		}

		// Get the output of the action
		output := action.Output()
		if output == nil {
			// Log if no output is returned
			log.Printf("Action '%s' returned nil output", name)
		}
		return output, nil
	}, opts)

	return nil
}

// Handler is the type for functions that handle actions
// Registry to map handler names to actual functions
var handlerRegistry = map[string]typ.Handler{}

//---------------------
// Module Implementation
//---------------------

type APIActionModule struct{}

// Name returns the name of the API module.
func (m *APIActionModule) Name() string {
	return "API Action Module"
}

// Init initializes the API action module.
func (m *APIActionModule) Init() error {
	log.Println("Initializing API Action Module")
	return nil
}

// Stop stops the API action module.
func (m *APIActionModule) Stop() error {
	log.Println("Stopping API Action Module")
	return nil
}

// NewAPIActionModule creates and returns an instance of the API Action Module
func NewAPIActionModule() *APIActionModule {
	return &APIActionModule{}
}

// RegisterHandler registers a handler for a given name
func RegisterHandler(name string, handler typ.Handler) error {
	// Ensure the handler registry contains the handler by name
	name = strings.ToLower(name)
	if _, exists := handlerRegistry[name]; exists {
		return fmt.Errorf("handler '%s' already registered", name)
	}
	handlerRegistry[name] = handler
	return nil
}

//---------------------
// API Registry
//---------------------

var registry = map[string]APIEntry{}

// Register adds a new API entry with metadata.
func Register(name string, handler func([]any) (any, error), opts APIOption) {
	name = strings.ToLower(name)
	registry[name] = APIEntry{
		Name:    name,
		Handler: handler,
		Options: opts,
	}
}

// Get retrieves an API entry by name.
func Get(name string) (APIEntry, bool) {
	entry, ok := registry[strings.ToLower(name)]
	return entry, ok
}

//---------------------
// API Types
//---------------------

// APIOption defines metadata for exposing an action via API.
type APIOption struct {
	Public bool     // If true, action is exposed via /api/<name>
	Auth   bool     // Reserved for future use (authorization)
	Doc    string   // Optional documentation string
	Tags   []string // Optional categorization
}

// APIEntry stores a handler and its associated metadata.
type APIEntry struct {
	Name    string                        // Action name
	Handler func(args []any) (any, error) // Wrapped handler function
	Options APIOption                     // API exposure options
}

//---------------------
// Error Handling
//---------------------

var ErrHandlerNotFound = errors.New("handler not found")

// Example usage of error handling in the API registration process
func (m *APIActionModule) Actions() []typ.ActionDef {
	return []typ.ActionDef{
		{
			Name: "api.register_action",
			Func: func(a typ.IAction) any {
				name := a.InputString(0)
				handlerName := a.InputString(1)

				// Check for valid handler registration
				handler, found := handlerRegistry[strings.ToLower(handlerName)]
				if !found {
					return fmt.Errorf("handler '%s' not found", handlerName)
				}

				opts := APIOption{
					Public: a.InputBool(2),
					Auth:   a.InputBool(3),
					Doc:    a.InputString(4),
					Tags:   []string{}, // Optional tags
				}

				// Register action with handler
				err := RegisterAction(name, handler, opts)
				if err != nil {
					return fmt.Errorf("error registering action: %v", err)
				}

				return "Action registered successfully"
			},
			Public: true,
		},
	}
}
