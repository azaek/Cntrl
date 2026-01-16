//go:build windows

package stats

import (
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// LOGICAL_PROCESSOR_RELATIONSHIP constants
const (
	RelationProcessorCore = 0
)

// SYSTEM_LOGICAL_PROCESSOR_INFORMATION structure
type systemLogicalProcessorInformation struct {
	ProcessorMask uintptr
	Relationship  int32
	_             [16]byte // Union padding
}

// FILETIME structure for GetSystemTimes
type fileTime struct {
	LowDateTime  uint32
	HighDateTime uint32
}

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

		// Read from registry
		key, err := registry.OpenKey(registry.LOCAL_MACHINE,
			`HARDWARE\DESCRIPTION\System\CentralProcessor\0`,
			registry.QUERY_VALUE)
		if err == nil {
			defer key.Close()

			// Get processor name
			if name, _, err := key.GetStringValue("ProcessorNameString"); err == nil {
				cpuInfoCache.Brand = strings.TrimSpace(name)
				// Extract manufacturer from brand
				if strings.Contains(strings.ToLower(name), "intel") {
					cpuInfoCache.Manufacturer = "Intel"
				} else if strings.Contains(strings.ToLower(name), "amd") {
					cpuInfoCache.Manufacturer = "AMD"
				} else {
					cpuInfoCache.Manufacturer = "Unknown"
				}
			}

			// Get speed in MHz
			if mhz, _, err := key.GetIntegerValue("~MHz"); err == nil {
				cpuInfoCache.Speed = float64(mhz) / 1000.0 // Convert to GHz
			}
		}

		// Fallback values
		if cpuInfoCache.Brand == "" {
			cpuInfoCache.Brand = "Unknown Processor"
		}
		if cpuInfoCache.Manufacturer == "" {
			cpuInfoCache.Manufacturer = "Unknown"
		}
	})

	return cpuInfoCache
}

// getPhysicalCores uses GetLogicalProcessorInformation to count physical cores
func getPhysicalCores() int {
	var bufferSize uint32 = 0

	// First call to get required buffer size
	ret, _, _ := procGetLogicalProcessorInfo.Call(
		0,
		uintptr(unsafe.Pointer(&bufferSize)),
	)
	if ret != 0 {
		return runtime.NumCPU() // Fallback
	}

	if bufferSize == 0 {
		return runtime.NumCPU()
	}

	// Allocate buffer
	buffer := make([]byte, bufferSize)

	ret, _, err := procGetLogicalProcessorInfo.Call(
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&bufferSize)),
	)
	if ret == 0 {
		// Check if error is meaningful
		if err != nil && err != syscall.Errno(0) {
			return runtime.NumCPU() // Fallback
		}
		return runtime.NumCPU()
	}

	// Count processor cores
	physicalCores := 0
	structSize := unsafe.Sizeof(systemLogicalProcessorInformation{})
	numStructs := int(bufferSize) / int(structSize)

	for i := 0; i < numStructs; i++ {
		offset := uintptr(i) * structSize
		info := (*systemLogicalProcessorInformation)(unsafe.Pointer(&buffer[offset]))
		if info.Relationship == RelationProcessorCore {
			physicalCores++
		}
	}

	if physicalCores == 0 {
		return runtime.NumCPU()
	}

	return physicalCores
}

// getSystemTimes wraps the Windows API call
func getSystemTimes(idleTime, kernelTime, userTime *fileTime) error {
	ret, _, err := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(idleTime)),
		uintptr(unsafe.Pointer(kernelTime)),
		uintptr(unsafe.Pointer(userTime)),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// calculateCPULoad calculates CPU load percentage by sampling over 100ms
func calculateCPULoad() float64 {
	var idleTime1, kernelTime1, userTime1 fileTime
	var idleTime2, kernelTime2, userTime2 fileTime

	// Get first sample
	if err := getSystemTimes(&idleTime1, &kernelTime1, &userTime1); err != nil {
		return 0
	}

	// Wait a short time
	time.Sleep(100 * time.Millisecond)

	// Get second sample
	if err := getSystemTimes(&idleTime2, &kernelTime2, &userTime2); err != nil {
		return 0
	}

	// Convert to uint64 for calculation
	idle1 := uint64(idleTime1.HighDateTime)<<32 | uint64(idleTime1.LowDateTime)
	idle2 := uint64(idleTime2.HighDateTime)<<32 | uint64(idleTime2.LowDateTime)
	kernel1 := uint64(kernelTime1.HighDateTime)<<32 | uint64(kernelTime1.LowDateTime)
	kernel2 := uint64(kernelTime2.HighDateTime)<<32 | uint64(kernelTime2.LowDateTime)
	user1 := uint64(userTime1.HighDateTime)<<32 | uint64(userTime1.LowDateTime)
	user2 := uint64(userTime2.HighDateTime)<<32 | uint64(userTime2.LowDateTime)

	idleDiff := float64(idle2 - idle1)
	totalDiff := float64((kernel2 - kernel1) + (user2 - user1))

	if totalDiff == 0 {
		return 0
	}

	// CPU load = 100 - idle percentage
	load := 100.0 * (1.0 - idleDiff/totalDiff)
	if load < 0 {
		load = 0
	}
	if load > 100 {
		load = 100
	}

	return load
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
