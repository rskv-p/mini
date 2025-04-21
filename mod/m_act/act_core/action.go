package act_core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_bus/bus_type"
)

//---------------------
// Types
//---------------------

type Action struct {
	ID      string // Resource ID
	Name    string // Full name (e.g. "models.user.get")
	Group   string // Top-level namespace
	Method  string // Method name
	Handler string // Registered handler key
	Inputs  []any  // Input arguments
	output  *any   // Execution result
	ctx     context.Context
	Bus     bus_type.IBusClient // Access to the bus
}

//---------------------
// Registry
//---------------------

var Handlers = map[string]act_type.Handler{}

func Register(name string, handler act_type.Handler) {
	Handlers[strings.ToLower(name)] = handler
}

func RegisterGroup(name string, group map[string]act_type.Handler) {
	for method, handler := range group {
		full := fmt.Sprintf("%s.%s", strings.ToLower(name), strings.ToLower(method))
		Handlers[full] = handler
	}
}

func (a *Action) GetName() string {
	return a.Name
}

func Alias(name, alias string) {
	if h, ok := Handlers[strings.ToLower(name)]; ok {
		Handlers[strings.ToLower(alias)] = h
	} else {
		log.Printf("Error: action %s does not exist", name)
	}
}

func Exists(name string) bool {
	name = strings.ToLower(name)
	return Handlers[name] != nil || strings.HasPrefix(name, "script.")
}

//---------------------
// Factory
//---------------------

// NewAction creates a new Action, optionally with a context.
func NewAction(name string, ctx context.Context, args ...any) *Action {
	action, err := createAction(name, args...)
	if err != nil {
		log.Printf("Error creating action: %v", err)
		panic(fmt.Errorf("action: %w", err))
	}

	// Set context if provided
	if ctx != nil {
		action = action.WithContext(ctx).(*Action)
	}
	return action
}

// createAction is a helper function to create an Action without the context part.
func createAction(name string, args ...any) (*Action, error) {
	act := &Action{Name: name, Inputs: args}
	return act.init() // returns initialized Action
}

//---------------------
// Builders
//---------------------

func (a *Action) WithContext(ctx context.Context) act_type.IAction {
	a.ctx = ctx
	return a
}

func (a *Action) Context() context.Context {
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	return a.ctx
}

//---------------------
// Execution
//---------------------

// Execute executes the action with enhanced error handling and logging.
func (a *Action) Execute() (err error) {
	ctx := a.Context()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	hd, err := a.handler()
	if err != nil {
		log.Printf("Handler not found for action %s with inputs %v: %v", a.Name, a.Inputs, err)
		return err
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if rec := recover(); rec != nil {
				err = catch(rec)
			}
		}()
		val := hd(a)
		a.output = &val
	}()

	select {
	case <-ctx.Done():
		log.Printf("Action %s execution timed out: %v", a.Name, ctx.Err())
		return ctx.Err()
	case <-done:
		return err
	}
}

// Exec executes the action and handles errors.
func (a *Action) Exec() (any, error) {
	hd, err := a.handler()
	if err != nil {
		return nil, err
	}
	defer func() { err = catch(recover()) }()
	return hd(a), nil
}

// Run runs the action, enhanced with error handling.
func (a *Action) Run() any {
	hd, err := a.handler()
	if err != nil {
		log.Fatalf("Action execution failed for %s: %v", a.Name, err)
	}
	defer a.Release()
	return hd(a)
}

func (a *Action) Release() {
	a.output = nil
}

func (a *Action) Dispose() {
	a.Inputs = nil
	a.output = nil
	a.ctx = nil
}

func (a *Action) Output() any {
	if a.output != nil {
		return *a.output
	}
	return nil
}

//---------------------
// Internal
//---------------------

// SetOutput sets the output field of the Action.
func (a *Action) SetOutput(val any) {
	a.output = &val
}

func (a *Action) handler() (act_type.Handler, error) {
	if h, ok := Handlers[a.Handler]; ok {
		return h, nil
	}
	return nil, fmt.Errorf("handler not found: %s for action: %s", a.Handler, a.Name)
}

// Improved init with better error handling and flexibility.
func (a *Action) init() (*Action, error) {
	fields := strings.Split(a.Name, ".")
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid action name: %s", a.Name)
	}

	a.Group = fields[0]
	a.Method = fields[len(fields)-1]

	switch a.Group {
	case "schema", "file", "task", "cron":
		a.ID = strings.ToLower(strings.Join(fields[1:len(fields)-1], "."))
		a.Handler = strings.ToLower(fmt.Sprintf("%s.%s", a.Group, a.Method))

	case "flow":
		a.Handler = a.Group
		a.ID = strings.ToLower(strings.Join(fields[1:], "."))

	case "script", "http":
		if a.Group == "http" {
			a.Handler = strings.ToLower(fmt.Sprintf("%s.%s", a.Group, a.Method))
		} else {
			a.ID = strings.ToLower(strings.Join(fields[1:len(fields)-1], "."))
			a.Handler = a.Group
		}

	default:
		a.Handler = strings.ToLower(a.Name)
	}

	return a, nil
}

//---------------------
// Debug
//---------------------

func (a Action) String() string {
	in, _ := json.MarshalIndent(a.Inputs, "", "  ")
	return fmt.Sprintf("Action: %s\nInputs:\n%s\n", a.Name, string(in))
}

//---------------------
// Recover helper
//---------------------

func catch(rec any) error {
	if rec == nil {
		return nil
	}
	if err, ok := rec.(error); ok {
		return err
	}
	return fmt.Errorf("%v", rec)
}
