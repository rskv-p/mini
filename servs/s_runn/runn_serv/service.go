package runn_serv

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/rskv-p/mini/servs/s_runn/runn_cfg"
)

type ProcessInfo struct {
	Cmd     *exec.Cmd
	Pid     int
	Started time.Time
}

type RunnService struct {
	cfg       runn_cfg.RunnConfig
	processes map[string]*ProcessInfo
	mu        sync.Mutex
}

func New(cfg runn_cfg.RunnConfig) *RunnService {
	return &RunnService{
		cfg:       cfg,
		processes: map[string]*ProcessInfo{},
	}
}

// StartServiceLocal starts a service by name
func (r *RunnService) StartServiceLocal(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.processes[name]; ok {
		return nil // already running
	}

	var conf *runn_cfg.ServiceConfig
	for i := range r.cfg.Services {
		if r.cfg.Services[i].Name == name {
			conf = &r.cfg.Services[i]
			break
		}
	}
	if conf == nil {
		return nil // not found
	}

	// Prepare command
	cmd := exec.Command(conf.Path, conf.Args...)

	// Setup logging to _data/log/<name>.log
	_ = os.MkdirAll("_data/log", 0755)
	logPath := filepath.Join("_data/log", name+".log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Start process
	if err := cmd.Start(); err != nil {
		return err
	}

	r.processes[name] = &ProcessInfo{
		Cmd:     cmd,
		Pid:     cmd.Process.Pid,
		Started: time.Now(),
	}

	_ = r.SaveState() // persist state

	return nil
}

// StopServiceLocal stops a running service
func (r *RunnService) StopServiceLocal(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	proc, ok := r.processes[name]
	if !ok {
		return nil
	}

	if err := proc.Cmd.Process.Kill(); err != nil {
		return err
	}
	delete(r.processes, name)
	return nil
}

// StopAllServicesLocal kills all running services
func (r *RunnService) StopAllServicesLocal() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, proc := range r.processes {
		_ = proc.Cmd.Process.Kill()
		delete(r.processes, name)
	}
}

// ListLocal returns names of all running services
func (r *RunnService) ListLocal() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	names := make([]string, 0, len(r.processes))
	for name := range r.processes {
		names = append(names, name)
	}
	return names
}

// GetProcess returns the ProcessInfo for a running service
func (r *RunnService) GetProcess(name string) *ProcessInfo {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.processes[name]
}
