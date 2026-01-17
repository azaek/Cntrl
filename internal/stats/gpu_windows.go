//go:build windows

package stats

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	gpuCache     *GpuStats
	gpuCacheMu   sync.RWMutex
	gpuCacheTime time.Time
	gpuCacheTTL  = 2 * time.Second // Cache GPU stats for 2 seconds
)

// GetGpuStats retrieves GPU statistics
// Attempts detection in order: NVIDIA, AMD, Intel
func GetGpuStats() (*GpuStats, error) {
	// Check cache
	gpuCacheMu.RLock()
	if gpuCache != nil && time.Since(gpuCacheTime) < gpuCacheTTL {
		cached := *gpuCache
		gpuCacheMu.RUnlock()
		return &cached, nil
	}
	gpuCacheMu.RUnlock()

	// Try NVIDIA first
	stats, err := getNvidiaStats()
	if err == nil && stats != nil {
		cacheGpuStats(stats)
		return stats, nil
	}

	// Try AMD
	stats, err = getAmdStats()
	if err == nil && stats != nil {
		cacheGpuStats(stats)
		return stats, nil
	}

	// Try Intel
	stats, err = getIntelStats()
	if err == nil && stats != nil {
		cacheGpuStats(stats)
		return stats, nil
	}

	// No GPU detected
	return nil, nil
}

func cacheGpuStats(stats *GpuStats) {
	gpuCacheMu.Lock()
	defer gpuCacheMu.Unlock()
	gpuCache = stats
	gpuCacheTime = time.Now()
}

// getNvidiaStats uses nvidia-smi to get NVIDIA GPU stats
func getNvidiaStats() (*GpuStats, error) {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=utilization.gpu,utilization.memory,temperature.gpu,name,memory.total,memory.used",
		"--format=csv,noheader,nounits")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	line := strings.TrimSpace(string(output))
	parts := strings.Split(line, ", ")
	if len(parts) < 6 {
		return nil, nil
	}

	utilGpu, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	// parts[1] is memory utilization (not used currently)
	temp, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
	name := strings.TrimSpace(parts[3])
	vramTotal, _ := strconv.Atoi(strings.TrimSpace(parts[4]))
	vramUsed, _ := strconv.Atoi(strings.TrimSpace(parts[5]))

	// Calculate VRAM percentage
	vramPercent := 0
	if vramTotal > 0 {
		vramPercent = (vramUsed * 100) / vramTotal
	}

	return &GpuStats{
		Vendor:         "NVIDIA",
		Model:          name,
		Vram:           &vramTotal,
		VramUsed:       &vramPercent,
		TemperatureGpu: &temp,
		UtilizationGpu: &utilGpu,
	}, nil
}

// getAmdStats uses AMD's rocm-smi or falls back to WMIC
func getAmdStats() (*GpuStats, error) {
	// Try rocm-smi first
	stats, err := getAmdRocmStats()
	if err == nil && stats != nil {
		return stats, nil
	}

	// Fallback to WMIC for basic info
	return getAmdWmicStats()
}

func getAmdRocmStats() (*GpuStats, error) {
	// rocm-smi --showtemp --showuse --showmeminfo vram
	cmd := exec.Command("rocm-smi", "--showtemp", "--showuse", "--showmeminfo", "vram", "--csv")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse rocm-smi CSV output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var temp, util int
	var model string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "GPU") && strings.Contains(line, "Temp") {
			continue // header
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			temp, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			util, _ = strconv.Atoi(strings.TrimSpace(parts[2]))
		}
	}

	// Get model name separately
	cmdName := exec.Command("rocm-smi", "--showproductname")
	cmdName.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	nameOutput, _ := cmdName.Output()
	if len(nameOutput) > 0 {
		lines := strings.Split(string(nameOutput), "\n")
		for _, l := range lines {
			if strings.Contains(l, "Card") {
				parts := strings.SplitN(l, ":", 2)
				if len(parts) == 2 {
					model = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	if model == "" {
		model = "AMD GPU"
	}

	return &GpuStats{
		Vendor:         "AMD",
		Model:          model,
		TemperatureGpu: &temp,
		UtilizationGpu: &util,
	}, nil
}

func getAmdWmicStats() (*GpuStats, error) {
	// Use WMIC to detect AMD GPU
	cmd := exec.Command("wmic", "path", "win32_VideoController", "where", "AdapterCompatibility like '%AMD%' or name like '%Radeon%'", "get", "name,AdapterRAM", "/format:csv")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Node") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			vramBytes, _ := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
			vramMB := int(vramBytes / 1024 / 1024)
			name := strings.TrimSpace(parts[2])

			if strings.Contains(strings.ToLower(name), "radeon") || strings.Contains(strings.ToLower(name), "amd") {
				return &GpuStats{
					Vendor: "AMD",
					Model:  name,
					Vram:   &vramMB,
				}, nil
			}
		}
	}

	return nil, nil
}

// getIntelStats detects Intel integrated graphics
func getIntelStats() (*GpuStats, error) {
	cmd := exec.Command("wmic", "path", "win32_VideoController", "where", "AdapterCompatibility like '%Intel%' or name like '%Intel%'", "get", "name,AdapterRAM", "/format:csv")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Node") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			vramBytes, _ := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
			vramMB := int(vramBytes / 1024 / 1024)
			name := strings.TrimSpace(parts[2])

			if strings.Contains(strings.ToLower(name), "intel") {
				return &GpuStats{
					Vendor: "Intel",
					Model:  name,
					Vram:   &vramMB,
				}, nil
			}
		}
	}

	return nil, nil
}
