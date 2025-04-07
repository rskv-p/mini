package runn_serv

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/rskv-p/mini/servs/s_runn/runn_cfg"
	"gorm.io/gorm"
)

// Errors
var (
	ErrNotFound       = errors.New("proc: not found")
	ErrInitDatabase   = errors.New("proc: error initializing db")
	ErrAlreadyStarted = errors.New("proc: already started")
	ErrNotStarted     = errors.New("proc: not started")
)

// Hook function type
type HookFunc func(process *Process)

// Hooks for process events
type Hooks struct {
	OnStart HookFunc
	OnStop  HookFunc
}

// Service structure for process management
type Service struct {
	db              *gorm.DB
	activeProcesses map[uint64]*ActiveProcess
	activeMutex     sync.RWMutex
	hooks           Hooks
}

// DB access
func (service *Service) DB() *gorm.DB {
	return service.db
}

// Set hooks for start/stop events
func (service *Service) SetHooks(h Hooks) {
	service.hooks = h
}

// New service initialization with DB and active processes
func New(db *gorm.DB) (*Service, error) {
	if err := db.AutoMigrate(&Process{}); err != nil {
		return nil, ErrInitDatabase
	}

	service := &Service{
		db:              db,
		activeProcesses: make(map[uint64]*ActiveProcess),
	}

	activeIDs, err := service.getActiveProcesses()
	if err != nil {
		return nil, err
	}
	for _, id := range activeIDs {
		if err := service.Start(id); err != nil {
			return nil, err
		}
	}

	return service, nil
}

// Fetch active processes from DB
func (service *Service) getActiveProcesses() ([]uint64, error) {
	var processes []Process
	if err := service.db.Where("disabled = ?", false).Find(&processes).Error; err != nil {
		return nil, err
	}

	var ids []uint64
	for _, p := range processes {
		ids = append(ids, p.ID)
	}
	return ids, nil
}

// List all processes
func (service *Service) Processes() ([]*Process, error) {
	var list []Process
	if err := service.db.Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*Process, len(list))
	for i := range list {
		out[i] = &list[i]
	}
	return out, nil
}

// Get process by ID
func (service *Service) Get(id uint64) (*Process, error) {
	var proc Process
	if err := service.db.First(&proc, id).Error; err != nil {
		return nil, err
	}
	return &proc, nil
}

// Add a new process
func (service *Service) Add(cmd, dir string) (*Process, error) {
	proc := &Process{
		Cmd:      cmd,
		Dir:      dir,
		Status:   "stopped",
		Disabled: true,
	}
	if err := service.db.Create(proc).Error; err != nil {
		return nil, err
	}
	return proc, nil
}

// Remove a process by ID
func (service *Service) Remove(id uint64) error {
	service.activeMutex.Lock()
	defer service.activeMutex.Unlock()

	if _, exists := service.activeProcesses[id]; exists {
		return ErrAlreadyStarted
	}
	return service.db.Delete(&Process{}, id).Error
}

// Start a process by ID
func (service *Service) Start(id uint64) error {
	service.activeMutex.Lock()
	defer service.activeMutex.Unlock()

	if _, exists := service.activeProcesses[id]; exists {
		return ErrAlreadyStarted
	}

	var proc Process
	if err := service.db.First(&proc, id).Error; err != nil {
		return err
	}
	proc.Disabled = false
	if err := service.db.Save(&proc).Error; err != nil {
		return err
	}

	service.activeProcesses[id] = newActiveProcess(service, id)

	if service.hooks.OnStart != nil {
		if startedProc, err := service.Get(id); err == nil {
			service.hooks.OnStart(startedProc)
		}
	}

	return nil
}

// Get active process by ID
func (service *Service) getActive(id uint64) (*ActiveProcess, error) {
	service.activeMutex.RLock()
	defer service.activeMutex.RUnlock()

	ap, ok := service.activeProcesses[id]
	if !ok {
		return nil, ErrNotStarted
	}
	return ap, nil
}

// Stop a running process
func (service *Service) Stop(id uint64) error {
	ap, err := service.getActive(id)
	if err != nil {
		return err
	}
	return ap.sigkill()
}

// Pause a running process
func (service *Service) Pause(id uint64) error {
	ap, err := service.getActive(id)
	if err != nil {
		return err
	}
	return ap.sigstop()
}

// Resume a paused process
func (service *Service) Resume(id uint64) error {
	ap, err := service.getActive(id)
	if err != nil {
		return err
	}
	return ap.sigcont()
}

// Load preconfigured processes from config
func LoadPreconfiguredProcesses(db *gorm.DB) error {
	for _, p := range runn_cfg.C().PreconfiguredProcesses {
		var process Process
		if err := db.Where("name = ?", p.Name).First(&process).Error; err != nil {
			if err := CreateProcess(db, p.Cmd, p.Dir, p.Name, p.StartOnLaunch); err != nil {
				log.Printf("Failed to add preconfigured process: %s", err)
				return err
			}
		}
	}
	return nil
}

// Create a new process in DB
func CreateProcess(db *gorm.DB, cmd, dir, name string, startOnLaunch bool) error {
	process := Process{
		Cmd:    cmd,
		Dir:    dir,
		Status: "stopped",
	}

	if err := db.Create(&process).Error; err != nil {
		return err
	}

	log.Printf("Process %s added to DB.", name)

	if startOnLaunch {
		go StartProcess(db, &process)
	}

	return nil
}

// Start a given process
func StartProcess(db *gorm.DB, process *Process) error {
	cmd := exec.Command(process.Cmd)
	cmd.Dir = process.Dir
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting process %s: %v", process.Cmd, err)
		return err
	}

	process.Status = "running"
	if err := db.Save(process).Error; err != nil {
		log.Printf("Error updating process status %s: %v", process.Cmd, err)
		return err
	}

	log.Printf("Process %s started.", process.Cmd)
	return nil
}

// Start all processes in a group, considering dependencies
func (service *Service) StartProcessGroup(groupID uint64) error {
	var group ProcessGroup
	if err := service.DB().First(&group, groupID).Error; err != nil {
		return err
	}

	var processes []Process
	if err := service.DB().Where("group_id = ?", groupID).Find(&processes).Error; err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, process := range processes {
		wg.Add(1)
		go func(p Process) {
			defer wg.Done()

			for _, depID := range p.Dependencies {
				var dependentProcess Process
				if err := service.DB().First(&dependentProcess, depID).Error; err != nil {
					log.Printf("Error getting dependent process %v: %v", depID, err)
					continue
				}

				for dependentProcess.Status != "running" {
					log.Printf("Waiting for process %v to finish before starting %v", depID, p.ID)
					time.Sleep(time.Second)
					if err := service.DB().First(&dependentProcess, depID).Error; err != nil {
						log.Printf("Error checking dependent process %v: %v", depID, err)
						break
					}
				}
			}

			if err := StartProcess(service.DB(), &p); err != nil {
				log.Printf("Failed to start process %v: %v", p.ID, err)
			}
		}(process)
	}

	wg.Wait()
	return nil
}

// Update a process by ID
func (service *Service) UpdateProcess(id string, cmd string, dir string, dependencies []uint64) error {
	var process Process
	if err := service.DB().First(&process, id).Error; err != nil {
		return fmt.Errorf("process with ID %s not found: %v", id, err)
	}

	if cmd != "" {
		process.Cmd = cmd
	}
	if dir != "" {
		process.Dir = dir
	}
	if dependencies != nil {
		process.Dependencies = dependencies
	}

	if err := service.DB().Save(&process).Error; err != nil {
		return fmt.Errorf("failed to save process: %v", err)
	}

	return nil
}
