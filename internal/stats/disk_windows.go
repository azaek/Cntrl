//go:build windows

package stats

import (
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	diskCache     []DiskStats
	diskCacheMu   sync.RWMutex
	diskCacheTime time.Time
	diskCacheTTL  = 30 * time.Second // Default cache TTL
)

// Drive types
const (
	DRIVE_UNKNOWN     = 0
	DRIVE_NO_ROOT_DIR = 1
	DRIVE_REMOVABLE   = 2
	DRIVE_FIXED       = 3
	DRIVE_REMOTE      = 4
	DRIVE_CDROM       = 5
	DRIVE_RAMDISK     = 6
)

// SetDiskCacheTTL sets the disk stats cache duration
func SetDiskCacheTTL(seconds int) {
	diskCacheMu.Lock()
	defer diskCacheMu.Unlock()
	diskCacheTTL = time.Duration(seconds) * time.Second
}

// GetDiskStats retrieves disk statistics for all fixed drives
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

	// Get logical drives bitmask
	ret, _, _ := procGetLogicalDrives.Call()
	if ret == 0 {
		return nil, windows.GetLastError()
	}

	var stats []DiskStats
	driveMask := uint32(ret)

	for i := 0; i < 26; i++ {
		if driveMask&(1<<i) == 0 {
			continue
		}

		driveLetter := string(rune('A' + i))
		drivePath := driveLetter + ":\\"

		// Check drive type - only include fixed drives
		driveType, _, _ := procGetDriveType.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(drivePath))))
		if driveType != DRIVE_FIXED {
			continue
		}

		// Get volume info for filesystem type
		var volumeName [256]uint16
		var fsName [256]uint16
		var serialNumber, maxComponentLen, fsFlags uint32

		ret, _, _ := procGetVolumeInfo.Call(
			uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(drivePath))),
			uintptr(unsafe.Pointer(&volumeName[0])),
			uintptr(len(volumeName)),
			uintptr(unsafe.Pointer(&serialNumber)),
			uintptr(unsafe.Pointer(&maxComponentLen)),
			uintptr(unsafe.Pointer(&fsFlags)),
			uintptr(unsafe.Pointer(&fsName[0])),
			uintptr(len(fsName)),
		)

		fsType := "Unknown"
		if ret != 0 {
			fsType = windows.UTF16ToString(fsName[:])
		}

		// Get disk space
		var freeBytesAvailable, totalBytes, totalFreeBytes uint64
		err := windows.GetDiskFreeSpaceEx(
			windows.StringToUTF16Ptr(drivePath),
			&freeBytesAvailable,
			&totalBytes,
			&totalFreeBytes,
		)
		if err != nil {
			continue
		}

		used := int64(totalBytes - totalFreeBytes)
		usedPercent := 0.0
		if totalBytes > 0 {
			usedPercent = float64(used) / float64(totalBytes) * 100
		}

		mount := driveLetter + ":"
		stats = append(stats, DiskStats{
			Fs:          mount,
			Type:        fsType,
			Size:        int64(totalBytes),
			Used:        used,
			Available:   int64(freeBytesAvailable),
			UsedPercent: usedPercent,
			Mount:       mount,
		})
	}

	// Update cache
	diskCacheMu.Lock()
	diskCache = stats
	diskCacheTime = time.Now()
	diskCacheMu.Unlock()

	return stats, nil
}
