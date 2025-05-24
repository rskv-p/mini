// file: arc/service/health.go
package service

import (
	"log"
	"strconv"
	"sync"

	"github.com/rskv-p/mini/service/config"
	"github.com/rskv-p/mini/service/constant"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

// ----------------------------------------------------
// Configuration helper
// ----------------------------------------------------

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

// ----------------------------------------------------
// Health probe registration
// ----------------------------------------------------

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

// ----------------------------------------------------
// Core health-check logic
// ----------------------------------------------------

// healthCheck evaluates memory, CPU, and custom probes.
func healthCheck(cfg *config.Config) (int, map[string]any) {
	if cfg == nil {
		log.Println("[health] missing configuration")
		return constant.StatusWarning, nil
	}

	status := constant.StatusOK
	feedback := make(map[string]any)

	// memory usage check
	memCrit := cfg.HCMemoryCriticalThreshold
	memWarn := cfg.HCMemoryWarningThreshold
	if vm, err := mem.VirtualMemory(); err == nil {
		used := vm.UsedPercent
		free := 100 - used
		if memCrit > 0 && free < memCrit {
			msg := "Memory critical: used=" + strconv.FormatFloat(used, 'f', 1, 64) + "%"
			log.Println("[health] " + msg)
			feedback[constant.MemoryCriticalKey] = msg
			status |= constant.StatusCritical
		} else if memWarn > 0 && free < memWarn {
			msg := "Memory warning: used=" + strconv.FormatFloat(used, 'f', 1, 64) + "%"
			log.Println("[health] " + msg)
			feedback[constant.MemoryWarningKey] = msg
			status |= constant.StatusWarning
		}
	}

	// CPU load check
	loadCrit := cfg.HCLoadCriticalThreshold
	loadWarn := cfg.HCLoadWarningThreshold
	if avg, err := load.Avg(); err == nil {
		cores := int32(0)
		if info, err := cpu.Info(); err == nil {
			for _, c := range info {
				cores += c.Cores
			}
		}
		if cores == 0 {
			cores = 1
		}
		ratio := avg.Load5 / float64(cores)
		if loadCrit > 0 && ratio > loadCrit {
			msg := "CPU load critical: load5=" + strconv.FormatFloat(ratio, 'f', 2, 64)
			log.Println("[health] " + msg)
			feedback[constant.LoadCriticalKey] = msg
			status |= constant.StatusCritical
		} else if loadWarn > 0 && ratio > loadWarn {
			msg := "CPU load warning: load5=" + strconv.FormatFloat(ratio, 'f', 2, 64)
			log.Println("[health] " + msg)
			feedback[constant.LoadWarningKey] = msg
			status |= constant.StatusWarning
		}
	}

	// custom health probes
	healthProbesMu.RLock()
	for _, probe := range healthProbes {
		key, st, info := probe()
		if key == "" {
			continue
		}
		if st == constant.StatusCritical {
			status |= constant.StatusCritical
		} else if st == constant.StatusWarning && status < constant.StatusCritical {
			status |= constant.StatusWarning
		}
		feedback[key] = info
	}
	healthProbesMu.RUnlock()

	// cap status at critical
	if status > constant.StatusCritical {
		status = constant.StatusCritical
	}
	return status, feedback
}
