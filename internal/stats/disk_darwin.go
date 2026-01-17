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
	diskCache     []DiskStats
	diskCacheMu   sync.RWMutex
	diskCacheTime time.Time
	diskCacheTTL  = 30 * time.Second // Default cache TTL
)

// SetDiskCacheTTL sets the disk stats cache duration
func SetDiskCacheTTL(seconds int) {
	diskCacheMu.Lock()
	defer diskCacheMu.Unlock()
	diskCacheTTL = time.Duration(seconds) * time.Second
}

// GetDiskStats retrieves disk statistics for all mounted volumes
func GetDiskStats() ([]DiskStats, error) {
	// Check cache
	diskCacheMu.RLock()
	if diskCache != nil && time.Since(diskCacheTime) < diskCacheTTL {
		cached := make([]DiskStats, len(diskCache))
		copy(cached, diskCache)
		diskCacheMu.RUnlock()
		return cached, nil
	}
	diskCacheMu.RUnlock()

	// Use df command to get disk info
	// -k for 1K blocks, -P for POSIX format
	out, err := exec.Command("df", "-k", "-P").Output()
	if err != nil {
		return nil, err
	}

	var stats []DiskStats
	lines := strings.Split(string(out), "\n")

	for i, line := range lines {
		// Skip header
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		filesystem := fields[0]
		mountPoint := fields[5]

		// Filter to only include physical disks and main volumes
		// Skip network mounts, devfs, map, etc.
		if !strings.HasPrefix(filesystem, "/dev/") {
			continue
		}

		// Skip snapshot and recovery volumes
		if strings.Contains(mountPoint, "/.") || strings.Contains(mountPoint, "/System/Volumes/") {
			// But keep Data volume
			if !strings.Contains(mountPoint, "/System/Volumes/Data") {
				continue
			}
		}

		// Parse sizes (in 1K blocks)
		sizeBlocks, _ := strconv.ParseInt(fields[1], 10, 64)
		usedBlocks, _ := strconv.ParseInt(fields[2], 10, 64)
		availBlocks, _ := strconv.ParseInt(fields[3], 10, 64)

		// Convert to bytes
		size := sizeBlocks * 1024
		used := usedBlocks * 1024
		available := availBlocks * 1024

		usedPercent := 0.0
		if size > 0 {
			usedPercent = float64(used) / float64(size) * 100
		}

		// Get filesystem type using diskutil (optional, may slow down)
		fsType := "Unknown"
		if strings.Contains(filesystem, "disk") {
			diskutilOut, err := exec.Command("diskutil", "info", filesystem).Output()
			if err == nil {
				for _, line := range strings.Split(string(diskutilOut), "\n") {
					if strings.Contains(line, "Type (Bundle):") || strings.Contains(line, "File System Personality:") {
						parts := strings.SplitN(line, ":", 2)
						if len(parts) == 2 {
							fsType = strings.TrimSpace(parts[1])
							break
						}
					}
				}
			}
		}

		stats = append(stats, DiskStats{
			Fs:          filesystem,
			Type:        fsType,
			Size:        size,
			Used:        used,
			Available:   available,
			UsedPercent: usedPercent,
			Mount:       mountPoint,
		})
	}

	// Update cache
	diskCacheMu.Lock()
	diskCache = stats
	diskCacheTime = time.Now()
	diskCacheMu.Unlock()

	return stats, nil
}
