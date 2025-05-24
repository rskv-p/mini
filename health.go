package service

import (
	"log"
	"strconv"
	"sync"

	"github.com/rskv-p/mini/config"
	"github.com/rskv-p/mini/constant"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

// ToFloat converts various types to float64.
func ToFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	}
	return 0
}

// getFloat safely extracts float config values.
func getFloat(cfg config.IConfig, key string) float64 {
	return ToFloat(cfg.MustString(key))
}

// HealthProbe is a function that returns key + status.
type HealthProbe func() (key string, status int, info any)

var (
	healthProbesMu sync.RWMutex
	healthProbes   []HealthProbe
)

// RegisterHealthProbe adds a custom health-check function.
func RegisterHealthProbe(probe HealthProbe) {
	healthProbesMu.Lock()
	defer healthProbesMu.Unlock()
	healthProbes = append(healthProbes, probe)
}

// healthCheck evaluates memory, CPU, and registered probes.
func healthCheck(cfg config.IConfig) (int, map[string]any) {
	if cfg == nil {
		log.Println("[health] missing configuration")
		return constant.StatusWarning, nil
	}

	status := constant.StatusOK
	feedback := make(map[string]any)

	// --- Memory Check ---
	memCrit := getFloat(cfg, "hc_memory_critical")
	memWarn := getFloat(cfg, "hc_memory_warning")
	if vm, err := mem.VirtualMemory(); err == nil {
		used := vm.UsedPercent
		free := 100 - used
		switch {
		case memCrit > 0 && free < memCrit:
			msg := "Memory critical: used=" + strconv.FormatFloat(used, 'f', 1, 64) + "%"
			log.Println("[health] " + msg)
			feedback[constant.MemoryCriticalKey] = msg
			status |= constant.StatusCritical
		case memWarn > 0 && free < memWarn:
			msg := "Memory warning: used=" + strconv.FormatFloat(used, 'f', 1, 64) + "%"
			log.Println("[health] " + msg)
			feedback[constant.MemoryWarningKey] = msg
			status |= constant.StatusWarning
		}
	}

	// --- CPU Load Check ---
	loadCrit := getFloat(cfg, "hc_load_critical")
	loadWarn := getFloat(cfg, "hc_load_warning")
	if avg, err := load.Avg(); err == nil {
		cores := int32(1)
		if info, err := cpu.Info(); err == nil {
			cores = 0
			for _, c := range info {
				cores += c.Cores
			}
			if cores == 0 {
				cores = 1
			}
		}
		ratio := avg.Load5 / float64(cores)
		switch {
		case loadCrit > 0 && ratio > loadCrit:
			msg := "CPU load critical: load5=" + strconv.FormatFloat(ratio, 'f', 2, 64)
			log.Println("[health] " + msg)
			feedback[constant.LoadCriticalKey] = msg
			status |= constant.StatusCritical
		case loadWarn > 0 && ratio > loadWarn:
			msg := "CPU load warning: load5=" + strconv.FormatFloat(ratio, 'f', 2, 64)
			log.Println("[health] " + msg)
			feedback[constant.LoadWarningKey] = msg
			status |= constant.StatusWarning
		}
	}

	// --- Custom Health Probes ---
	healthProbesMu.RLock()
	for _, probe := range healthProbes {
		key, st, info := probe()
		if key == "" {
			continue
		}
		switch st {
		case constant.StatusCritical:
			status |= constant.StatusCritical
		case constant.StatusWarning:
			if status < constant.StatusCritical {
				status |= constant.StatusWarning
			}
		}
		feedback[key] = info
	}
	healthProbesMu.RUnlock()

	if status > constant.StatusCritical {
		status = constant.StatusCritical
	}
	return status, feedback
}
