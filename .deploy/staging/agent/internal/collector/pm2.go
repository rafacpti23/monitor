package collector

import (
	"encoding/json"
	"log"
	"os/exec"
	"runtime"
)

// PM2Process represents a PM2-managed process.
type PM2Process struct {
	Name         string  `json:"name"`
	Status       string  `json:"status"`
	CPU          float64 `json:"cpu"`
	MemoryBytes  int64   `json:"memory"`
	PID          int     `json:"pid"`
	RestartCount int     `json:"restart_count"`
	UptimeMs     int64   `json:"uptime_ms"`
}

// CollectPM2 lists PM2 processes via `pm2 jlist`.
// Returns an empty slice (not an error) if PM2 is unavailable.
func CollectPM2() []PM2Process {
	if runtime.GOOS == "windows" {
		return nil
	}

	if _, err := exec.LookPath("pm2"); err != nil {
		return nil
	}

	out, err := exec.Command("pm2", "jlist").Output()
	if err != nil {
		log.Printf("[pm2] failed to run pm2 jlist: %v", err)
		return nil
	}

	var rawList []map[string]interface{}
	if err := json.Unmarshal(out, &rawList); err != nil {
		log.Printf("[pm2] failed to parse output: %v", err)
		return nil
	}

	procs := make([]PM2Process, 0, len(rawList))
	for _, raw := range rawList {
		p := PM2Process{}

		if name, ok := raw["name"].(string); ok {
			p.Name = name
		}

		if pid, ok := raw["pid"].(float64); ok {
			p.PID = int(pid)
		}

		// Status lives inside pm2_env
		if env, ok := raw["pm2_env"].(map[string]interface{}); ok {
			if status, ok := env["status"].(string); ok {
				p.Status = status
			}
			if restarts, ok := env["restart_time"].(float64); ok {
				p.RestartCount = int(restarts)
			}
			if uptime, ok := env["pm_uptime"].(float64); ok {
				p.UptimeMs = int64(uptime)
			}
		}

		// CPU and memory live inside monit
		if monit, ok := raw["monit"].(map[string]interface{}); ok {
			if c, ok := monit["cpu"].(float64); ok {
				p.CPU = c
			}
			if m, ok := monit["memory"].(float64); ok {
				p.MemoryBytes = int64(m)
			}
		}

		procs = append(procs, p)
	}

	return procs
}
