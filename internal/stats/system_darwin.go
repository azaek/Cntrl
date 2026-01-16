//go:build darwin

package stats

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	cachedHostname string
)

// GetSystemStats collects all system statistics
func GetSystemStats(hostnameOverride string, gpuEnabled bool) (*SystemStats, error) {
	// Memory
	memory, err := GetMemoryStats()
	if err != nil {
		return nil, err
	}

	// CPU
	cpu, err := GetCpuStats()
	if err != nil {
		return nil, err
	}

	// GPU (optional)
	var gpu *GpuStats
	if gpuEnabled {
		gpu, _ = GetGpuStats() // Ignore errors, GPU is optional
	}

	// Disks
	disks, err := GetDiskStats()
	if err != nil {
		disks = []DiskStats{} // Empty array on error
	}

	// Hostname
	hostname := hostnameOverride
	if hostname == "" {
		if cachedHostname == "" {
			cachedHostname, _ = os.Hostname()
		}
		hostname = cachedHostname
	}

	// Uptime
	uptime := getSystemUptime()

	return &SystemStats{
		Memory:   *memory,
		Cpu:      *cpu,
		Gpu:      gpu,
		Disks:    disks,
		Uptime:   uptime,
		Hostname: hostname,
		Platform: "darwin",
	}, nil
}

// getSystemUptime returns system uptime in seconds using sysctl
func getSystemUptime() int64 {
	// Get boot time using sysctl
	out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
	if err != nil {
		return 0
	}

	// Output format: { sec = 1234567890, usec = 123456 } ...
	outStr := string(out)
	if strings.Contains(outStr, "sec =") {
		parts := strings.Split(outStr, "sec =")
		if len(parts) >= 2 {
			secPart := strings.TrimSpace(parts[1])
			secPart = strings.Split(secPart, ",")[0]
			if bootTime, err := strconv.ParseInt(strings.TrimSpace(secPart), 10, 64); err == nil {
				// Calculate uptime from boot time
				now := currentTimeSeconds()
				return now - bootTime
			}
		}
	}

	return 0
}

// currentTimeSeconds returns current Unix timestamp
func currentTimeSeconds() int64 {
	out, err := exec.Command("date", "+%s").Output()
	if err != nil {
		return 0
	}
	timestamp, _ := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	return timestamp
}
