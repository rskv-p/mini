package runn_serv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const statePath = "_data/data/runn.state.json"

// StateEntry holds persisted info about a running process
type StateEntry struct {
	Pid     int       `json:"pid"`
	Started time.Time `json:"started"`
}

// RunnState represents the full saved state
type RunnState struct {
	Processes map[string]StateEntry `json:"processes"`
}

// SaveState writes current processes to runn.state.json
func (r *RunnService) SaveState() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	state := RunnState{
		Processes: make(map[string]StateEntry, len(r.processes)),
	}
	for name, proc := range r.processes {
		state.Processes[name] = StateEntry{
			Pid:     proc.Pid,
			Started: proc.Started,
		}
	}

	_ = os.MkdirAll(filepath.Dir(statePath), 0755)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

// LoadState loads previous runn state from file
func LoadState() (*RunnState, error) {
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, err
	}

	var state RunnState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}
