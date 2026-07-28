package collector

import (
	"log"
	"time"

	"github.com/p-mon/agent/internal/config"
)

// Payload is the complete metrics payload sent to the backend.
type Payload struct {
	Timestamp        int64           `json:"timestamp"`
	Hostname         string          `json:"hostname"`
	OS               string          `json:"os"`
	Arch             string          `json:"arch"`
	Kernel           string          `json:"kernel"`
	UptimeSeconds    uint64          `json:"uptime_seconds"`
	CPUPercent       float64         `json:"cpu_percent"`
	CPUCores         int             `json:"cpu_cores"`
	CPUModel         string          `json:"cpu_model"`
	LoadAvg          [3]float64      `json:"load_avg"`
	RAMTotalBytes    uint64          `json:"ram_total_bytes"`
	RAMUsedBytes     uint64          `json:"ram_used_bytes"`
	SwapTotalBytes   uint64          `json:"swap_total_bytes"`
	SwapUsedBytes    uint64          `json:"swap_used_bytes"`
	Disks            []DiskInfo      `json:"disks"`
	NetRXBytes       uint64          `json:"net_rx_bytes"`
	NetTXBytes       uint64          `json:"net_tx_bytes"`
	ProcessCount     int             `json:"process_count"`
	DockerContainers []ContainerInfo `json:"docker_containers,omitempty"`
	PM2Processes     []PM2Process    `json:"pm2_processes,omitempty"`
	Services         []ServiceStatus `json:"services,omitempty"`
	AgentVersion     string          `json:"agent_version"`
}

// Collect runs all enabled collectors and returns a single aggregated Payload.
// Individual collector failures are logged but never propagate — one broken
// collector must not prevent the others from reporting.
func Collect(cfg *config.Config, version string) Payload {
	p := Payload{
		Timestamp:    time.Now().Unix(),
		AgentVersion: version,
	}

	// System metrics
	if cfg.Collect.System {
		func() {
			defer recoverCollector("system")

			sys, err := CollectSystem()
			if err != nil {
				log.Printf("[system] collection error: %v", err)
				return
			}
			p.Hostname = sys.Hostname
			p.OS = sys.OS
			p.Arch = sys.Arch
			p.Kernel = sys.Kernel
			p.UptimeSeconds = sys.UptimeSeconds
			p.CPUPercent = sys.CPUPercent
			p.CPUCores = sys.CPUCores
			p.CPUModel = sys.CPUModel
			p.LoadAvg = sys.LoadAvg
			p.RAMTotalBytes = sys.RAMTotalBytes
			p.RAMUsedBytes = sys.RAMUsedBytes
			p.SwapTotalBytes = sys.SwapTotal
			p.SwapUsedBytes = sys.SwapUsed
			p.Disks = sys.Disks
			p.NetRXBytes = sys.NetRXBytes
			p.NetTXBytes = sys.NetTXBytes
			p.ProcessCount = sys.ProcessCount
		}()
	}

	// Docker containers
	if cfg.Collect.Docker {
		func() {
			defer recoverCollector("docker")
			if containers := CollectDocker(); len(containers) > 0 {
				p.DockerContainers = containers
			}
		}()
	}

	// PM2 processes
	if cfg.Collect.PM2 {
		func() {
			defer recoverCollector("pm2")
			if procs := CollectPM2(); len(procs) > 0 {
				p.PM2Processes = procs
			}
		}()
	}

	// Named services
	if len(cfg.Collect.Services) > 0 {
		func() {
			defer recoverCollector("services")
			if svcs := CollectServices(cfg.Collect.Services); len(svcs) > 0 {
				p.Services = svcs
			}
		}()
	}

	return p
}

// recoverCollector catches panics in a collector so one crash doesn't kill
// the entire collection cycle.
func recoverCollector(name string) {
	if r := recover(); r != nil {
		log.Printf("[%s] collector panic: %v", name, r)
	}
}
