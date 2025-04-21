package m_api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// API Registration
//---------------------

// RegisterAction registers a handler in x_action,
// and exposes it as a public API if the Public option is set.
func RegisterAction(name string, handler typ.Handler, opts APIOption) error {
	// Register in x_action
	act.Register(name, handler)

	// Wrap as a public API endpoint if the Public option is set
	if opts.Public {
		return RegisterAPI(name, handler, opts)
	}
	return nil
}

// RegisterAPI registers a public API endpoint for the action.
func RegisterAPI(name string, handler typ.Handler, opts APIOption) error {
	// Ensure the action is properly registered as a public API
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

//---------------------
// Module Implementation
//---------------------

// ApiModule represents the API module that implements the IModule interface.
type ApiModule struct {
	Module          typ.IModule
	handlerRegistry map[string]typ.Handler
	logClient       typ.ILogClient // Client for logging

}

// Name returns the name of the API module.
func (m *ApiModule) Name() string {
	return m.Module.Name()
}

// Init initializes the API action module.
func (m *ApiModule) Init() error {
	log.Println("Initializing API Module")

	// Initialize the handler registry
	m.handlerRegistry = make(map[string]typ.Handler)

	// Register the /docs endpoint
	registerDocsEndpoint()

	return nil
}

// Stop stops the module (specific shutdown logic can be added).
func (s *ApiModule) Start() error {
	// Here you can implement stop logic if needed
	return nil
}

// Stop stops the API action module.
func (m *ApiModule) Stop() error {
	log.Println("Stopping API Module")
	return nil
}

// NewApiModule creates and returns an instance of the API action module.
func NewApiModule(service typ.IService) *ApiModule {
	// Create a new module using NewModule
	module := mod.NewModule("api", service, nil, nil, nil)

	// Return the ApiModule with the created module
	return &ApiModule{
		Module: module,
	}
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
	Public bool     // If true, the action is exposed via /api/<name>
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

//---------------------
// Handlers for Actions
//---------------------

var _ typ.IModule = (*ApiModule)(nil)

// Actions returns a list of actions the module can perform.
func (m *ApiModule) Actions() []typ.ActionDef {
	return []typ.ActionDef{
		{
			Name: "api.register_action",
			Func: func(a typ.IAction) any {
				name := a.InputString(0)
				handlerName := a.InputString(1)

				// Check if the handler is registered
				handler, found := m.handlerRegistry[strings.ToLower(handlerName)]
				if !found {
					return fmt.Errorf("handler '%s' not found", handlerName)
				}

				opts := APIOption{
					Public: a.InputBool(2),
					Auth:   a.InputBool(3),
					Doc:    a.InputString(4),
					Tags:   []string{}, // Optional tags
				}

				// Register the action with the handler
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

//---------------------
// Docs Endpoint
//---------------------

// registerDocsEndpoint registers the /docs endpoint for the API module.
func registerDocsEndpoint() {
	// Here we pass an empty APIOption with Public set to true
	Register("api.docs", func(args []any) (any, error) {
		// Create a map to store the documentation
		docs := map[string]any{}

		// Iterate over the registered handlers
		for name := range act.Handlers {
			// Retrieve the API entry for each handler
			apiEntry, ok := Get(name)
			docs[name] = map[string]any{
				"handler": name,
				"public":  ok && apiEntry.Options.Public,
				"doc":     apiEntry.Options.Doc,
				"tags":    apiEntry.Options.Tags,
			}
		}

		// Return the generated documentation
		return docs, nil
	}, APIOption{Public: true}) // Pass the APIOption
}
