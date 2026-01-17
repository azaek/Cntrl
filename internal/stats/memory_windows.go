//go:build windows

package stats

import (
	"sync"
	"time"
	"unsafe"
)

var (
	memoryCache     *MemoryStats
	memoryCacheMu   sync.RWMutex
	memoryCacheTime time.Time
	memoryCacheTTL  = 500 * time.Millisecond
)

// MEMORYSTATUSEX structure
type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

// GetMemoryStats retrieves current memory statistics using Windows syscall
func GetMemoryStats() (*MemoryStats, error) {
	// Check cache
	memoryCacheMu.RLock()
	if memoryCache != nil && time.Since(memoryCacheTime) < memoryCacheTTL {
		cached := *memoryCache
		memoryCacheMu.RUnlock()
		return &cached, nil
	}
	memoryCacheMu.RUnlock()

	var memStatus memoryStatusEx
	memStatus.Length = uint32(unsafe.Sizeof(memStatus))

	ret, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret == 0 {
		return nil, err
	}

	total := int64(memStatus.TotalPhys)
	free := int64(memStatus.AvailPhys)
	used := total - free

	stats := &MemoryStats{
		Total:       total,
		Used:        used,
		Free:        free,
		UsedPercent: float64(memStatus.MemoryLoad),
	}

	// Update cache
	memoryCacheMu.Lock()
	memoryCache = stats
	memoryCacheTime = time.Now()
	memoryCacheMu.Unlock()

	return stats, nil
}
