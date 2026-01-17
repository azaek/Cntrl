//go:build darwin

package stats

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	memoryCache     *MemoryStats
	memoryCacheMu   sync.RWMutex
	memoryCacheTime time.Time
	memoryCacheTTL  = 500 * time.Millisecond
)

// GetMemoryStats retrieves current memory statistics using sysctl and vm_stat
func GetMemoryStats() (*MemoryStats, error) {
	// Check cache
	memoryCacheMu.RLock()
	if memoryCache != nil && time.Since(memoryCacheTime) < memoryCacheTTL {
		cached := *memoryCache
		memoryCacheMu.RUnlock()
		return &cached, nil
	}
	memoryCacheMu.RUnlock()

	// Get total physical memory using sysctl
	totalOut, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return nil, err
	}
	total, err := strconv.ParseInt(strings.TrimSpace(string(totalOut)), 10, 64)
	if err != nil {
		return nil, err
	}

	// Get memory usage from vm_stat
	vmOut, err := exec.Command("vm_stat").Output()
	if err != nil {
		return nil, err
	}

	// Parse vm_stat output
	pageSize := int64(4096) // Default page size
	var pagesActive, pagesWired, pagesCompressed, pagesFree int64

	lines := strings.Split(string(vmOut), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mach Virtual Memory Statistics") {
			// Extract page size if mentioned
			if strings.Contains(line, "page size of") {
				parts := strings.Split(line, "page size of")
				if len(parts) == 2 {
					sizeStr := strings.TrimSpace(strings.Split(parts[1], " ")[1])
					if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
						pageSize = size
					}
				}
			}
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(strings.TrimSuffix(parts[1], "."))
		value, _ := strconv.ParseInt(valueStr, 10, 64)

		switch key {
		case "Pages active":
			pagesActive = value
		case "Pages wired down":
			pagesWired = value
		case "Pages occupied by compressor":
			pagesCompressed = value
		case "Pages free":
			pagesFree = value
		}
	}

	// Calculate memory usage
	// Used = Active + Wired + Compressed
	used := (pagesActive + pagesWired + pagesCompressed) * pageSize
	free := pagesFree * pageSize

	// If our calculation seems off, fall back to simpler calculation
	if used <= 0 || used > total {
		used = total - free
	}

	usedPercent := 0.0
	if total > 0 {
		usedPercent = float64(used) / float64(total) * 100
	}

	stats := &MemoryStats{
		Total:       total,
		Used:        used,
		Free:        free,
		UsedPercent: usedPercent,
	}

	// Update cache
	memoryCacheMu.Lock()
	memoryCache = stats
	memoryCacheTime = time.Now()
	memoryCacheMu.Unlock()

	return stats, nil
}
