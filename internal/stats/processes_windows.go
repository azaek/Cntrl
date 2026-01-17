//go:build windows

package stats

import (
	"sort"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows API constants
const (
	TH32CS_SNAPPROCESS         = 0x00000002
	PROCESS_QUERY_LIMITED_INFO = 0x1000
	PROCESS_VM_READ            = 0x0010
)

// PROCESS_MEMORY_COUNTERS structure
type PROCESS_MEMORY_COUNTERS struct {
	Cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

var (
	modPsapi                 = windows.NewLazySystemDLL("psapi.dll")
	procGetProcessMemoryInfo = modPsapi.NewProc("GetProcessMemoryInfo")
	procGetProcessTimes      = modKernel32.NewProc("GetProcessTimes")
)

func GetProcesses() ([]Process, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Process32First(snapshot, &entry); err != nil {
		return nil, err
	}

	// Aggregate by process name
	aggregated := make(map[string]*Process)

	for {
		// Extract name
		name := windows.UTF16ToString(entry.ExeFile[:])

		// Skip system idle process
		if entry.ProcessID != 0 {
			memory, cpuTime := getProcessStats(entry.ProcessID)

			if existing, ok := aggregated[name]; ok {
				// Add to existing
				existing.Count++
				existing.Memory += memory
				existing.CpuTime += cpuTime
			} else {
				// Create new entry
				aggregated[name] = &Process{
					Name:    name,
					Count:   1,
					Memory:  memory,
					CpuTime: cpuTime,
				}
			}
		}

		if err := windows.Process32Next(snapshot, &entry); err != nil {
			break
		}
	}

	// Convert map to slice
	results := make([]Process, 0, len(aggregated))
	for _, p := range aggregated {
		p.MemoryMB = float64(p.Memory) / (1024 * 1024)
		results = append(results, *p)
	}

	// Sort by memory (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Memory > results[j].Memory
	})

	// Return top 20
	if len(results) > 20 {
		results = results[:20]
	}

	return results, nil
}

// getProcessStats retrieves memory and CPU time for a process
func getProcessStats(pid uint32) (uint64, float64) {
	handle, err := windows.OpenProcess(PROCESS_QUERY_LIMITED_INFO|PROCESS_VM_READ, false, pid)
	if err != nil {
		return 0, 0
	}
	defer windows.CloseHandle(handle)

	// Get memory
	var memCounters PROCESS_MEMORY_COUNTERS
	memCounters.Cb = uint32(unsafe.Sizeof(memCounters))
	procGetProcessMemoryInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&memCounters)),
		uintptr(memCounters.Cb),
	)
	memory := uint64(memCounters.WorkingSetSize)

	// Get CPU time
	var creationTime, exitTime, kernelTime, userTime windows.Filetime
	ret, _, _ := procGetProcessTimes.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&creationTime)),
		uintptr(unsafe.Pointer(&exitTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)

	var cpuTime float64
	if ret != 0 {
		kernelNS := uint64(kernelTime.HighDateTime)<<32 | uint64(kernelTime.LowDateTime)
		userNS := uint64(userTime.HighDateTime)<<32 | uint64(userTime.LowDateTime)
		totalNS := kernelNS + userNS
		cpuTime = float64(totalNS) / 10000000.0
	}

	return memory, cpuTime
}
