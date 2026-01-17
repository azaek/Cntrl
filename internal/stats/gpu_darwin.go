//go:build darwin

package stats

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	gpuCache     *GpuStats
	gpuCacheMu   sync.RWMutex
	gpuCacheTime time.Time
	gpuCacheTTL  = 2 * time.Second // Cache GPU stats for 2 seconds
)

// GetGpuStats retrieves GPU statistics using system_profiler
func GetGpuStats() (*GpuStats, error) {
	// Check cache
	gpuCacheMu.RLock()
	if gpuCache != nil && time.Since(gpuCacheTime) < gpuCacheTTL {
		cached := *gpuCache
		gpuCacheMu.RUnlock()
		return &cached, nil
	}
	gpuCacheMu.RUnlock()

	// Try NVIDIA first (if nvidia-smi is available)
	stats, err := getNvidiaStats()
	if err == nil && stats != nil {
		cacheGpuStats(stats)
		return stats, nil
	}

	// Fall back to system_profiler for integrated/Apple GPUs
	stats, err = getSystemProfilerGpuStats()
	if err == nil && stats != nil {
		cacheGpuStats(stats)
		return stats, nil
	}

	return nil, nil
}

func cacheGpuStats(stats *GpuStats) {
	gpuCacheMu.Lock()
	defer gpuCacheMu.Unlock()
	gpuCache = stats
	gpuCacheTime = time.Now()
}

// getNvidiaStats uses nvidia-smi to get NVIDIA GPU stats (for external GPUs)
func getNvidiaStats() (*GpuStats, error) {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=utilization.gpu,utilization.memory,temperature.gpu,name,memory.total,memory.used",
		"--format=csv,noheader,nounits")

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
	temp, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
	name := strings.TrimSpace(parts[3])
	vramTotal, _ := strconv.Atoi(strings.TrimSpace(parts[4]))
	vramUsed, _ := strconv.Atoi(strings.TrimSpace(parts[5]))

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

// SystemProfilerGPU represents GPU info from system_profiler JSON output
type systemProfilerGPU struct {
	SPDisplaysDataType []struct {
		Name              string `json:"_name"`
		Model             string `json:"sppci_model"`
		Vendor            string `json:"spdisplays_vendor"`
		Cores             string `json:"sppci_cores"`
		DeviceType        string `json:"sppci_device_type"`
		MetalFamily       string `json:"spdisplays_mtlgpufamilysupport"`
		Bus               string `json:"sppci_bus"`
		VRAMDynamic       string `json:"spdisplays_vram_dynamic,omitempty"`
		VRAMShared        string `json:"spdisplays_vram_shared,omitempty"`
		VRAMStatic        string `json:"spdisplays_vram,omitempty"`
	} `json:"SPDisplaysDataType"`
}

// getSystemProfilerGpuStats uses system_profiler to get GPU info
func getSystemProfilerGpuStats() (*GpuStats, error) {
	out, err := exec.Command("system_profiler", "SPDisplaysDataType", "-json").Output()
	if err != nil {
		return nil, err
	}

	var profilerData systemProfilerGPU
	if err := json.Unmarshal(out, &profilerData); err != nil {
		// Fall back to text parsing
		return parseSystemProfilerText()
	}

	if len(profilerData.SPDisplaysDataType) == 0 {
		return parseSystemProfilerText()
	}

	gpu := profilerData.SPDisplaysDataType[0]

	// Get model from sppci_model field
	model := gpu.Model
	if model == "" {
		// Try to extract from _name field (e.g., "kHW_AppleM1Item" -> "Apple M1")
		name := gpu.Name
		if strings.Contains(name, "AppleM") {
			// Extract chip name from kHW_AppleM1Item format
			name = strings.TrimPrefix(name, "kHW_")
			name = strings.TrimSuffix(name, "Item")
			// Convert AppleM1 to Apple M1
			name = strings.Replace(name, "AppleM", "Apple M", 1)
			model = name
		}
	}

	// Determine vendor from vendor field or model name
	vendor := "Unknown"
	vendorStr := strings.ToLower(gpu.Vendor)
	if strings.Contains(vendorStr, "apple") {
		vendor = "Apple"
	} else {
		modelLower := strings.ToLower(model)
		if strings.Contains(modelLower, "apple") || strings.Contains(modelLower, "m1") ||
			strings.Contains(modelLower, "m2") || strings.Contains(modelLower, "m3") ||
			strings.Contains(modelLower, "m4") {
			vendor = "Apple"
		} else if strings.Contains(modelLower, "intel") {
			vendor = "Intel"
		} else if strings.Contains(modelLower, "amd") || strings.Contains(modelLower, "radeon") {
			vendor = "AMD"
		}
	}

	// Parse VRAM (Apple Silicon uses unified memory, so VRAM fields may be empty)
	var vramMB *int
	vramStr := gpu.VRAMStatic
	if vramStr == "" {
		vramStr = gpu.VRAMDynamic
	}
	if vramStr == "" {
		vramStr = gpu.VRAMShared
	}
	if vramStr != "" {
		// Parse values like "8 GB" or "1536 MB"
		vramStr = strings.ToUpper(vramStr)
		if strings.Contains(vramStr, "GB") {
			numStr := strings.TrimSpace(strings.Split(vramStr, "GB")[0])
			if num, err := strconv.Atoi(numStr); err == nil {
				mb := num * 1024
				vramMB = &mb
			}
		} else if strings.Contains(vramStr, "MB") {
			numStr := strings.TrimSpace(strings.Split(vramStr, "MB")[0])
			if num, err := strconv.Atoi(numStr); err == nil {
				vramMB = &num
			}
		}
	}

	// Parse GPU cores (Apple Silicon specific)
	var gpuCores *int
	if gpu.Cores != "" {
		if cores, err := strconv.Atoi(gpu.Cores); err == nil {
			gpuCores = &cores
		}
	}

	// Initialize default values (-1 means unavailable)
	notAvailable := -1
	vramUsed := notAvailable
	temp := notAvailable
	util := notAvailable

	// For Apple Silicon, include cores info in model if available
	if gpuCores != nil && vendor == "Apple" && !strings.Contains(model, "core") {
		model = fmt.Sprintf("%s (%d-core GPU)", model, *gpuCores)
	}

	return &GpuStats{
		Vendor:         vendor,
		Model:          model,
		Vram:           vramMB,
		VramUsed:       &vramUsed,
		TemperatureGpu: &temp,
		UtilizationGpu: &util,
	}, nil
}

// parseSystemProfilerText parses the text output of system_profiler as fallback
func parseSystemProfilerText() (*GpuStats, error) {
	out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var model, vendor string
	var vramMB *int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chipset Model:") {
			model = strings.TrimSpace(strings.TrimPrefix(line, "Chipset Model:"))
		} else if strings.HasPrefix(line, "Vendor:") {
			vendor = strings.TrimSpace(strings.TrimPrefix(line, "Vendor:"))
		} else if strings.HasPrefix(line, "VRAM") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				vramStr := strings.ToUpper(strings.TrimSpace(parts[1]))
				if strings.Contains(vramStr, "GB") {
					numStr := strings.TrimSpace(strings.Split(vramStr, " ")[0])
					if num, err := strconv.Atoi(numStr); err == nil {
						mb := num * 1024
						vramMB = &mb
					}
				} else if strings.Contains(vramStr, "MB") {
					numStr := strings.TrimSpace(strings.Split(vramStr, " ")[0])
					if num, err := strconv.Atoi(numStr); err == nil {
						vramMB = &num
					}
				}
			}
		}
	}

	if model == "" {
		return nil, nil
	}

	// Normalize vendor
	if vendor == "" {
		modelLower := strings.ToLower(model)
		if strings.Contains(modelLower, "apple") || strings.Contains(modelLower, "m1") ||
			strings.Contains(modelLower, "m2") || strings.Contains(modelLower, "m3") {
			vendor = "Apple"
		} else if strings.Contains(modelLower, "intel") {
			vendor = "Intel"
		} else if strings.Contains(modelLower, "amd") || strings.Contains(modelLower, "radeon") {
			vendor = "AMD"
		}
	}

	// Initialize default values (-1 means unavailable)
	notAvailable := -1
	vramUsed := notAvailable
	temp := notAvailable
	util := notAvailable

	return &GpuStats{
		Vendor:         vendor,
		Model:          model,
		Vram:           vramMB,
		VramUsed:       &vramUsed,
		TemperatureGpu: &temp,
		UtilizationGpu: &util,
	}, nil
}
