//go:build windows

package stats

import (
	"golang.org/x/sys/windows"
)

var (
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	// Common procedures
	procGetTickCount64          = modKernel32.NewProc("GetTickCount64")
	procGetLogicalProcessorInfo = modKernel32.NewProc("GetLogicalProcessorInformation")
	procGetSystemTimes          = modKernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatusEx    = modKernel32.NewProc("GlobalMemoryStatusEx")
	procGetDriveType            = modKernel32.NewProc("GetDriveTypeW")
	procGetVolumeInfo           = modKernel32.NewProc("GetVolumeInformationW")
	procGetLogicalDrives        = modKernel32.NewProc("GetLogicalDrives")
)
