package collector

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// SystemMetrics holds all system-level metrics.
type SystemMetrics struct {
	Hostname      string     `json:"hostname"`
	OS            string     `json:"os"`
	Arch          string     `json:"arch"`
	Kernel        string     `json:"kernel"`
	UptimeSeconds uint64     `json:"uptime_seconds"`
	CPUPercent    float64    `json:"cpu_percent"`
	CPUCores      int        `json:"cpu_cores"`
	CPUModel      string     `json:"cpu_model"`
	LoadAvg       [3]float64 `json:"load_avg"`
	RAMTotalBytes uint64     `json:"ram_total_bytes"`
	RAMUsedBytes  uint64     `json:"ram_used_bytes"`
	RAMAvailable  uint64     `json:"ram_available_bytes"`
	SwapTotal     uint64     `json:"swap_total_bytes"`
	SwapUsed      uint64     `json:"swap_used_bytes"`
	Disks         []DiskInfo `json:"disks"`
	NetRXBytes    uint64     `json:"net_rx_bytes"`
	NetTXBytes    uint64     `json:"net_tx_bytes"`
	ProcessCount  int        `json:"process_count"`
}

// DiskInfo represents a single mounted disk.
type DiskInfo struct {
	Path       string `json:"path"`
	TotalBytes uint64 `json:"total_bytes"`
	UsedBytes  uint64 `json:"used_bytes"`
	FreeBytes  uint64 `json:"free_bytes"`
	Filesystem string `json:"filesystem"`
}

// skipFilesystems lists pseudo-filesystems to ignore during disk collection.
var skipFilesystems = map[string]bool{
	"tmpfs":     true,
	"devtmpfs":  true,
	"devfs":     true,
	"proc":      true,
	"sysfs":     true,
	"cgroup":    true,
	"cgroup2":   true,
	"pstore":    true,
	"debugfs":   true,
	"securityfs": true,
	"configfs":  true,
	"fusectl":   true,
	"mqueue":    true,
	"hugetlbfs": true,
	"overlay":   true,
	"squashfs":  true,
	"nsfs":      true,
	"tracefs":   true,
	"binfmt_misc": true,
}

// CollectSystem gathers system metrics using gopsutil.
func CollectSystem() (*SystemMetrics, error) {
	m := &SystemMetrics{
		Disks: []DiskInfo{},
	}

	// Hostname
	if h, err := os.Hostname(); err == nil {
		m.Hostname = h
	}

	// OS and Arch
	m.OS = runtime.GOOS
	m.Arch = runtime.GOARCH

	// Kernel version
	if info, err := host.Info(); err == nil {
		m.Kernel = info.KernelVersion
	}

	// Uptime
	if uptime, err := host.Uptime(); err == nil {
		m.UptimeSeconds = uptime
	}

	// CPU info
	if cpuInfos, err := cpu.Info(); err == nil && len(cpuInfos) > 0 {
		m.CPUModel = cpuInfos[0].ModelName
		// Count total logical cores across all physical CPUs
		total := 0
		for _, ci := range cpuInfos {
			total += int(ci.Cores)
		}
		if total > 0 {
			m.CPUCores = total
		}
	}
	// Fallback: use cpu.Counts if Info didn't yield cores
	if m.CPUCores == 0 {
		if count, err := cpu.Counts(true); err == nil {
			m.CPUCores = count
		}
	}

	// CPU percent — 100ms sample, aggregate across all cores
	if percents, err := cpu.Percent(100*time.Millisecond, false); err == nil && len(percents) > 0 {
		m.CPUPercent = percents[0]
	}

	// Load average (returns zeroes on Windows, which is fine)
	if avg, err := load.Avg(); err == nil && avg != nil {
		m.LoadAvg = [3]float64{avg.Load1, avg.Load5, avg.Load15}
	}

	// RAM
	if vmem, err := mem.VirtualMemory(); err == nil {
		m.RAMTotalBytes = vmem.Total
		m.RAMUsedBytes = vmem.Used
		m.RAMAvailable = vmem.Available
	}

	// Swap
	if swap, err := mem.SwapMemory(); err == nil {
		m.SwapTotal = swap.Total
		m.SwapUsed = swap.Used
	}

	// Disks — real filesystems only
	if parts, err := disk.Partitions(false); err == nil {
		for _, part := range parts {
			fsLower := strings.ToLower(part.Fstype)
			if fsLower == "" || skipFilesystems[fsLower] {
				continue
			}
			// Skip /proc, /sys, /dev mount paths
			mp := part.Mountpoint
			if strings.HasPrefix(mp, "/proc") || strings.HasPrefix(mp, "/sys") || mp == "/dev" {
				continue
			}

			usage, err := disk.Usage(part.Mountpoint)
			if err != nil || usage == nil {
				continue
			}

			m.Disks = append(m.Disks, DiskInfo{
				Path:       part.Mountpoint,
				TotalBytes: usage.Total,
				UsedBytes:  usage.Used,
				FreeBytes:  usage.Free,
				Filesystem: part.Fstype,
			})
		}
	}

	// Network — sum across all non-loopback interfaces
	if counters, err := net.IOCounters(true); err == nil {
		for _, c := range counters {
			name := strings.ToLower(c.Name)
			if name == "lo" || strings.HasPrefix(name, "lo:") {
				continue
			}
			m.NetRXBytes += c.BytesRecv
			m.NetTXBytes += c.BytesSent
		}
	}

	// Process count
	if pids, err := process.Pids(); err == nil {
		m.ProcessCount = len(pids)
	}

	return m, nil
}
