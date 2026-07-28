package collector

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// ServiceStatus represents the status of a monitored service.
type ServiceStatus struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
	PID     int    `json:"pid,omitempty"`
}

// CollectServices checks the status of each named service.
func CollectServices(names []string) []ServiceStatus {
	if len(names) == 0 {
		return nil
	}

	results := make([]ServiceStatus, 0, len(names))
	for _, name := range names {
		s := ServiceStatus{Name: name}

		switch runtime.GOOS {
		case "linux":
			s.Running, s.PID = checkLinuxService(name)
		default:
			// Windows/macOS — return not-running gracefully
			s.Running = false
		}

		results = append(results, s)
	}

	return results
}

// checkLinuxService tries systemctl first, then falls back to /proc scan.
func checkLinuxService(name string) (running bool, pid int) {
	// Try systemctl
	if isSystemdAvailable() {
		out, err := exec.Command("systemctl", "is-active", name).Output()
		if err == nil && strings.TrimSpace(string(out)) == "active" {
			// Try to get the main PID
			pidOut, err := exec.Command("systemctl", "show", "-p", "MainPID", "--value", name).Output()
			if err == nil {
				if p, err := strconv.Atoi(strings.TrimSpace(string(pidOut))); err == nil && p > 0 {
					return true, p
				}
			}
			return true, 0
		}
	}

	// Fallback: scan /proc/*/comm
	return scanProcComm(name)
}

// isSystemdAvailable checks for systemctl on PATH.
func isSystemdAvailable() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}

// scanProcComm walks /proc looking for a process whose comm matches name.
func scanProcComm(name string) (bool, int) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		log.Printf("[services] cannot read /proc: %v", err)
		return false, 0
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		commPath := fmt.Sprintf("/proc/%d/comm", pid)
		data, err := os.ReadFile(commPath)
		if err != nil {
			continue
		}

		comm := strings.TrimSpace(string(data))
		if comm == name {
			return true, pid
		}
	}

	return false, 0
}
