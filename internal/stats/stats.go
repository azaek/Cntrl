package stats

import (
	"os"
)

var (
	// Ensure we only look up hostname once to save time
	cachedHostnameGlobal string
)

func getHostname(override string) string {
	if override != "" {
		return override
	}
	if cachedHostnameGlobal == "" {
		cachedHostnameGlobal, _ = os.Hostname()
	}
	return cachedHostnameGlobal
}

// GetSystemInfo returns static system information
func GetSystemInfo(hostnameOverride string, gpuEnabled bool) (*SystemInfo, error) {
	// 1. CPU
	cpuInfoStatic := getCPUStaticInfo()
	cpuInfo := CpuInfo{
		Manufacturer:  cpuInfoStatic.Manufacturer,
		Brand:         cpuInfoStatic.Brand,
		Cores:         cpuInfoStatic.Cores,
		PhysicalCores: cpuInfoStatic.PhysicalCores,
		BaseSpeed:     cpuInfoStatic.Speed,
	}

	// 2. Memory
	memStats, err := GetMemoryStats()
	memInfo := MemoryInfo{}
	if err == nil {
		memInfo.Total = memStats.Total
	}

	// 3. GPU
	var gpuInfo *GpuInfo
	if gpuEnabled {
		if gpuStats, _ := GetGpuStats(); gpuStats != nil {
			gpuInfo = &GpuInfo{
				Vendor: gpuStats.Vendor,
				Model:  gpuStats.Model,
				Vram:   gpuStats.Vram,
			}
		}
	}

	// 4. Disks
	diskStatsList, err := GetDiskStats()
	var diskInfos []DiskInfo
	if err == nil {
		for _, ds := range diskStatsList {
			diskInfos = append(diskInfos, DiskInfo{
				Fs:    ds.Fs,
				Type:  ds.Type,
				Size:  ds.Size,
				Mount: ds.Mount,
			})
		}
	} else {
		diskInfos = []DiskInfo{}
	}

	return &SystemInfo{
		Hostname: getHostname(hostnameOverride),
		Platform: getPlatformName(),
		Cpu:      cpuInfo,
		Gpu:      gpuInfo,
		Memory:   memInfo,
		Disks:    diskInfos,
	}, nil
}

// GetSystemUsage returns dynamic system usage
func GetSystemUsage(gpuEnabled bool) (*SystemUsage, error) {
	// 1. Uptime
	uptime := getSystemUptime()

	// 2. CPU
	cpuStats, _ := GetCpuStats()
	cpuUsage := CpuUsage{
		CurrentLoad:  0,
		CurrentSpeed: 0,
		CurrentTemp:  0,
	}
	if cpuStats != nil {
		cpuUsage.CurrentLoad = cpuStats.CurrentLoad
		cpuUsage.CurrentSpeed = cpuStats.Speed
		cpuUsage.CurrentTemp = 0
	}

	// 3. Memory
	memStats, _ := GetMemoryStats()
	memUsage := MemoryUsage{}
	if memStats != nil {
		memUsage.Used = memStats.Used
		memUsage.Free = memStats.Free
		memUsage.UsedPercent = memStats.UsedPercent
	}

	// 4. GPU
	var gpuUsage *GpuUsage
	if gpuEnabled {
		if gpuStats, _ := GetGpuStats(); gpuStats != nil {
			gpuUsage = &GpuUsage{
				UtilizationGpu: gpuStats.UtilizationGpu,
				TemperatureGpu: gpuStats.TemperatureGpu,
				VramUsed:       gpuStats.VramUsed,
			}
		}
	}

	// 5. Disks
	diskStatsList, _ := GetDiskStats()
	var diskUsages []DiskUsage
	for _, ds := range diskStatsList {
		diskUsages = append(diskUsages, DiskUsage{
			Fs:          ds.Fs,
			Used:        ds.Used,
			Available:   ds.Available,
			UsedPercent: ds.UsedPercent,
		})
	}
	if diskUsages == nil {
		diskUsages = []DiskUsage{}
	}

	return &SystemUsage{
		Uptime: uptime,
		Cpu:    cpuUsage,
		Memory: memUsage,
		Gpu:    gpuUsage,
		Disks:  diskUsages,
	}, nil
}

// TODO: Need these helpers to be available:
// - getPlatformName() -> "win32" or "macos"
