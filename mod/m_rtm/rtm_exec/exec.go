package rtm_exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_rtm/rtm_core"
)

// ErrNoScript is returned when no script is provided in the inputs.
var ErrNoScript = errors.New("exec_runtime: no script provided in Inputs[0]")

// ExecRuntime runs system commands as actions.
type ExecRuntime struct {
	mu        sync.Mutex
	processes map[string]*exec.Cmd
}

// Ensure interface compliance
var _ rtm_core.Runtime = (*ExecRuntime)(nil)

// New creates a new exec runtime.
func New() *ExecRuntime {
	return &ExecRuntime{
		processes: make(map[string]*exec.Cmd),
	}
}

// Init initializes the runtime. Currently not needed but can be extended.
func (r *ExecRuntime) Init() error {
	// Initialization logic (if any) can be added here.
	return nil
}

// Dispose stops and cleans up all running processes.
func (r *ExecRuntime) Dispose() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, cmd := range r.processes {
		if err := cmd.Process.Kill(); err != nil {
			//	x_log.RootLogger().Structured().Error("failed to kill process", x_log.FString("name", name), x_log.FError(err))
		}
		delete(r.processes, name)
	}
}

// Execute runs the system command defined by the action's inputs.
func (r *ExecRuntime) Execute(action act_type.IAction) (any, error) {
	// Check if the action has at least one input (the script)
	if !action.NumberOfInputsIs(1) {
		return nil, ErrNoScript
	}

	// Get the command (script) from the first input
	code := action.InputString(0)
	if code == "" {
		return nil, ErrNoScript
	}

	// Collect remaining arguments
	args := make([]string, 0, action.NumberOfInputs()-1)
	for i := 1; i < action.NumberOfInputs(); i++ {
		args = append(args, action.InputString(i))
	}

	// Set up the context with a timeout (30 seconds)
	ctx := action.Context()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build the command to execute
	cmd := exec.CommandContext(ctx, code, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Track the command
	r.mu.Lock()
	r.processes[action.GetName()] = cmd
	r.mu.Unlock()

	// Run the command
	err := cmd.Start()
	if err != nil {
		r.mu.Lock()
		delete(r.processes, action.GetName())
		r.mu.Unlock()
		//x_log.RootLogger().Structured().Error("failed to start process", x_log.FString("name", action.GetName()), x_log.FError(err))
		return nil, fmt.Errorf("exec_runtime: failed to start process: %w", err)
	}

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		r.mu.Lock()
		delete(r.processes, action.GetName())
		r.mu.Unlock()
		//	x_log.RootLogger().Structured().Error("process failed", x_log.FString("name", action.GetName()), x_log.FError(err))
		return nil, fmt.Errorf("exec_runtime: %w: %s", err, stderr.String())
	}

	// Process completed successfully
	//x_log.RootLogger().Structured().Info("process completed", x_log.FString("name", action.GetName()), x_log.FString("stdout", stdout.String()))
	return map[string]any{
		"pid":    cmd.Process.Pid,
		"stdout": stdout.String(),
	}, nil
}

// List returns the names of running actions.
func (r *ExecRuntime) List() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	names := make([]string, 0, len(r.processes))
	for name := range r.processes {
		names = append(names, name)
	}
	return names
}

// Stop terminates a running process by action name.
func (r *ExecRuntime) Stop(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cmd, ok := r.processes[name]
	if !ok {
		return fmt.Errorf("exec_runtime: no such process: %s", name)
	}
	err := cmd.Process.Kill()
	delete(r.processes, name)
	if err != nil {
		//	x_log.RootLogger().Structured().Error("failed to kill process", x_log.FString("name", name), x_log.FError(err))
		return err
	}
	//	x_log.RootLogger().Structured().Info("process terminated", x_log.FString("name", name))
	return nil
}
