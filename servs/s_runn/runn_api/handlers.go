package runn_api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rskv-p/mini/servs/s_runn/runn_serv"
)

// handleList retrieves and returns the list of processes.
func handleList(m *runn_serv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		procs, err := m.Processes()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = json.NewEncoder(w).Encode(procs)
	}
}

// handleStart starts the process with the given ID.
func handleStart(m *runn_serv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
		if err := m.Start(id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleStop stops the process with the given ID.
func handleStop(m *runn_serv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
		if err := m.Stop(id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// handlePause pauses the process with the given ID.
func handlePause(m *runn_serv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
		if err := m.Pause(id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleResume resumes the paused process with the given ID.
func handleResume(m *runn_serv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
		if err := m.Resume(id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleAdd adds a new process with the specified command and directory.
func handleAdd(m *runn_serv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Cmd string `json:"cmd"`
			Dir string `json:"dir"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		proc, err := m.Add(req.Cmd, req.Dir)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = json.NewEncoder(w).Encode(proc)
	}
}

// handleRemove removes the process with the given ID.
func handleRemove(m *runn_serv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
		if err := m.Remove(id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleStartGroup starts a group of processes with the given group ID.
func handleStartGroup(manager *runn_serv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID := chi.URLParam(r, "group_id")
		groupIDInt, err := strconv.Atoi(groupID)
		if err != nil {
			http.Error(w, "Invalid group ID", http.StatusBadRequest)
			return
		}

		if err := manager.StartProcessGroup(uint64(groupIDInt)); err != nil {
			http.Error(w, "Error starting process group", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// handleEdit updates the process with new data, including command, directory, and dependencies.
func handleEdit(manager *runn_serv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		var updatedData struct {
			Cmd          string   `json:"cmd"`
			Dir          string   `json:"dir"`
			Dependencies []uint64 `json:"dependencies"`
		}
		if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
			http.Error(w, "Invalid data", http.StatusBadRequest)
			return
		}

		if err := manager.UpdateProcess(id, updatedData.Cmd, updatedData.Dir, updatedData.Dependencies); err != nil {
			http.Error(w, "Error editing process", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
