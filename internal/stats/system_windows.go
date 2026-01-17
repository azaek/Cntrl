//go:build windows

package stats

import (
	"os"
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
		Platform: "win32", // Always win32 for compatibility
	}, nil
}

// getSystemUptime returns system uptime in seconds
func getSystemUptime() int64 {
	ret, _, _ := procGetTickCount64.Call()
	// GetTickCount64 returns milliseconds
	return int64(ret) / 1000
}

func getPlatformName() string {
	return "win32"
}
