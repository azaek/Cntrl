//go:build darwin

package stats

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	cpuInfoCache     *cpuStaticInfo
	cpuInfoCacheOnce sync.Once
	currentCPULoad   float64
	loadMutex        sync.RWMutex
)

func init() {
	go backgroundCPULoadCollector()
}

func backgroundCPULoadCollector() {
	for {
		load := calculateCPULoad()
		loadMutex.Lock()
		currentCPULoad = load
		loadMutex.Unlock()
		time.Sleep(2 * time.Second)
	}
}

type cpuStaticInfo struct {
	Manufacturer  string
	Brand         string
	Cores         int
	PhysicalCores int
	Speed         float64
}

// getCPUStaticInfo retrieves static CPU info (cached)
func getCPUStaticInfo() *cpuStaticInfo {
	cpuInfoCacheOnce.Do(func() {
		cpuInfoCache = &cpuStaticInfo{
			Cores:         runtime.NumCPU(),
			PhysicalCores: getPhysicalCores(),
		}

		// Get CPU brand string using sysctl
		brandOut, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
		if err == nil {
			cpuInfoCache.Brand = strings.TrimSpace(string(brandOut))

			// Extract manufacturer from brand
			brandLower := strings.ToLower(cpuInfoCache.Brand)
			if strings.Contains(brandLower, "intel") {
				cpuInfoCache.Manufacturer = "Intel"
			} else if strings.Contains(brandLower, "apple") {
				cpuInfoCache.Manufacturer = "Apple"
			} else if strings.Contains(brandLower, "amd") {
				cpuInfoCache.Manufacturer = "AMD"
			} else {
				cpuInfoCache.Manufacturer = "Unknown"
			}
		}

		// Get CPU frequency (may not be available on Apple Silicon)
		freqOut, err := exec.Command("sysctl", "-n", "hw.cpufrequency").Output()
		if err == nil {
			if freq, err := strconv.ParseInt(strings.TrimSpace(string(freqOut)), 10, 64); err == nil {
				cpuInfoCache.Speed = float64(freq) / 1e9 // Convert Hz to GHz
			}
		}

		// Fallback values
		if cpuInfoCache.Brand == "" {
			// Try alternative for Apple Silicon
			chipOut, _ := exec.Command("sysctl", "-n", "machdep.cpu.brand").Output()
			if len(chipOut) > 0 {
				cpuInfoCache.Brand = strings.TrimSpace(string(chipOut))
			} else {
				cpuInfoCache.Brand = "Apple Silicon"
			}
			cpuInfoCache.Manufacturer = "Apple"
		}
	})

	return cpuInfoCache
}

// getPhysicalCores returns the number of physical CPU cores
func getPhysicalCores() int {
	out, err := exec.Command("sysctl", "-n", "hw.physicalcpu").Output()
	if err != nil {
		return runtime.NumCPU()
	}

	cores, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil || cores <= 0 {
		return runtime.NumCPU()
	}
	return cores
}

// calculateCPULoad calculates CPU load using top command
func calculateCPULoad() float64 {
	// Use top -l 1 to get a single sample of CPU usage
	out, err := exec.Command("top", "-l", "1", "-n", "0", "-s", "0").Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CPU usage:") {
			// Parse: CPU usage: 5.26% user, 10.52% sys, 84.21% idle
			parts := strings.Split(line, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.Contains(part, "idle") {
					idleStr := strings.Split(part, "%")[0]
					idleStr = strings.TrimSpace(idleStr)
					// Handle "CPU usage: X% idle" format
					if strings.Contains(idleStr, ":") {
						idleStr = strings.Split(idleStr, ":")[1]
						idleStr = strings.TrimSpace(idleStr)
					}
					if idle, err := strconv.ParseFloat(idleStr, 64); err == nil {
						return 100.0 - idle
					}
				}
			}
		}
	}

	return 0
}

// GetCpuStats retrieves current CPU statistics
func GetCpuStats() (*CpuStats, error) {
	info := getCPUStaticInfo()

	loadMutex.RLock()
	load := currentCPULoad
	loadMutex.RUnlock()

	return &CpuStats{
		Manufacturer:  info.Manufacturer,
		Brand:         info.Brand,
		Cores:         info.Cores,
		PhysicalCores: info.PhysicalCores,
		Speed:         info.Speed,
		CurrentLoad:   load,
	}, nil
}
