package runn_serv

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/shlex"
)

// Constants for process management
const (
	timeout     = time.Second * 10
	maxRestarts = 3
)

// Process represents a system process
type Process struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Cmd          string    `json:"cmd"`
	Dir          string    `json:"dir,omitempty"`
	Status       string    `json:"status"`
	Disabled     bool      `json:"disabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	RestartCount int       `json:"restart_count"`
	LastExitCode int       `json:"last_exit_code"`
	GroupID      uint64    `json:"group_id,omitempty"`
	Dependencies []uint64  `json:"dependencies"`
}

// ProcessGroup represents a group of processes
type ProcessGroup struct {
	ID           uint64   `gorm:"primaryKey"`
	Name         string   `json:"name"`
	Dependencies []uint64 `json:"depends_on"`
}

// ActiveProcess manages the execution of a process
type ActiveProcess struct {
	service      *Service
	id           uint64
	pidMutex     *sync.Mutex
	pid          int
	exitMutex    *sync.Mutex
	exit         bool
	restartCount int
}

// New ActiveProcess instance
func newActiveProcess(service *Service, id uint64) *ActiveProcess {
	ap := &ActiveProcess{
		service:   service,
		id:        id,
		pidMutex:  &sync.Mutex{},
		pid:       0,
		exitMutex: &sync.Mutex{},
		exit:      false,
	}
	ap.setStatus("running")
	go ap.start()
	return ap
}

// Set the process status in DB
func (ap *ActiveProcess) setStatus(status string) {
	var proc Process
	if err := ap.service.DB().First(&proc, ap.id).Error; err != nil {
		return
	}
	proc.Status = status
	_ = ap.service.DB().Save(&proc).Error
}

// Cleanup process state and DB
func (ap *ActiveProcess) cleanup() {
	dao := ap.service
	dao.activeMutex.Lock()
	defer dao.activeMutex.Unlock()
	delete(dao.activeProcesses, ap.id)

	var proc Process
	if err := dao.db.First(&proc, ap.id).Error; err == nil {
		proc.Status = "stopped"
		proc.Disabled = true
		proc.RestartCount = ap.restartCount
		_ = dao.db.Save(&proc).Error

		if dao.hooks.OnStop != nil {
			dao.hooks.OnStop(&proc)
		}
	}
}

// Start the process and handle restarts
func (ap *ActiveProcess) start() {
	defer ap.cleanup()

	var proc Process
	if err := ap.service.DB().First(&proc, ap.id).Error; err != nil {
		return
	}

	parts, err := shlex.Split(proc.Cmd)
	if err != nil || len(parts) == 0 {
		return
	}

	dir := expandHomeDir(proc.Dir)
	dir, err = filepath.Abs(dir)
	if err != nil {
		return
	}

	// Attempt to start process with retries
	for ap.restartCount < maxRestarts {
		if ap.shouldExit() {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Dir = dir
		cmd.Env = os.Environ()

		if err := cmd.Start(); err != nil {
			ap.setStatus("failed")
			ap.restartCount++
			time.Sleep(timeout)
			continue
		}

		// Track PID and wait for process exit
		ap.pidMutex.Lock()
		ap.pid = cmd.Process.Pid
		ap.pidMutex.Unlock()

		err = cmd.Wait()

		// Capture exit status
		exitCode := 0
		if exitErr, ok := err.(*exec.ExitError); ok {
			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = ws.ExitStatus()
			}
		}

		ap.restartCount++

		// Update DB with restart count and exit code
		dbProc := &Process{}
		if err := ap.service.DB().First(dbProc, ap.id).Error; err == nil {
			dbProc.RestartCount = ap.restartCount
			dbProc.LastExitCode = exitCode
			_ = ap.service.DB().Save(dbProc).Error
		}

		// Handle failure or success
		if err != nil {
			ap.setStatus("failed")
			time.Sleep(timeout)
			continue
		}

		ap.setStatus("running")
		return
	}

	// Final failure state after max restarts
	ap.setStatus("failed")
}

// Check if the process should exit
func (ap *ActiveProcess) shouldExit() bool {
	ap.exitMutex.Lock()
	defer ap.exitMutex.Unlock()
	return ap.exit
}

// Kill the process
func (ap *ActiveProcess) sigkill() error {
	ap.exitMutex.Lock()
	defer ap.exitMutex.Unlock()
	ap.setStatus("stopping")
	ap.exit = true
	if ap.pid == 0 {
		return errors.New("proc: bad pid")
	}
	return syscall.Kill(-ap.pid, syscall.SIGKILL)
}

// Pause the process
func (ap *ActiveProcess) sigstop() error {
	if err := syscall.Kill(-ap.pid, syscall.SIGSTOP); err != nil {
		return err
	}
	ap.setStatus("paused")
	return nil
}

// Resume the process
func (ap *ActiveProcess) sigcont() error {
	if err := syscall.Kill(-ap.pid, syscall.SIGCONT); err != nil {
		return err
	}
	ap.setStatus("running")
	return nil
}

// Expand tilde (~) to home directory
func expandHomeDir(path string) string {
	if strings.HasPrefix(path, "~") {
		if usr, err := user.Current(); err == nil {
			return filepath.Join(usr.HomeDir, path[1:])
		}
	}
	return path
}
