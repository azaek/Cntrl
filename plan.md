# PC Controller Go Service - Implementation Plan

## Overview

Build a **lightweight, self-contained Windows service** in Go that exposes system stats and power control endpoints. The service should be distributable as a single executable with built-in service management CLI.

---

## Requirements

### Functional

-   Expose HTTP API for system stats (CPU, GPU, Memory, Disks)
-   Expose HTTP API for power commands (shutdown, restart, hibernate)
-   Run as a Windows service (survives reboots)
-   CLI for service management (install, uninstall, start, stop)
-   YAML configuration file for settings

### Non-Functional

-   Single executable (~10MB)
-   Low memory footprint (~15MB RAM)
-   Zero external dependencies at runtime
-   Works on any Windows 10/11 PC

---

## CLI Interface

```bash
pc-controller.exe install     # Install as Windows service (requires admin)
pc-controller.exe uninstall   # Remove the service (requires admin)
pc-controller.exe start       # Start the service
pc-controller.exe stop        # Stop the service
pc-controller.exe status      # Check if service is running
pc-controller.exe run         # Run in foreground (for debugging)
pc-controller.exe config      # Show config file location
pc-controller.exe version     # Show version info
```

---

## Project Structure

```
pc-controller-go/
├── cmd/
│   └── pc-controller/
│       └── main.go              # Entry point + CLI handling
├── internal/
│   ├── api/
│   │   ├── router.go            # HTTP router setup
│   │   ├── stats.go             # GET /rog/stats handler
│   │   ├── power.go             # POST /rog/pw/* handlers
│   │   └── status.go            # GET /rog/status handler
│   ├── stats/
│   │   ├── memory.go            # Memory stats (syscall)
│   │   ├── cpu.go               # CPU stats (registry + PDH)
│   │   ├── gpu.go               # GPU stats (nvidia-smi)
│   │   ├── disk.go              # Disk stats (GetDiskFreeSpace)
│   │   └── system.go            # Combined system stats
│   ├── power/
│   │   └── commands.go          # Shutdown/Restart/Hibernate
│   ├── service/
│   │   ├── service.go           # Windows service wrapper
│   │   └── manager.go           # Install/Uninstall/Start/Stop
│   └── config/
│       └── config.go            # YAML config loading
├── config.example.yaml          # Example configuration
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Configuration File

**Location:** Same directory as executable, or `%APPDATA%\pc-controller\config.yaml`

```yaml
# PC Controller Configuration

# Server settings
server:
    port: 8080 # HTTP server port
    host: "0.0.0.0" # Bind address (0.0.0.0 = all interfaces)

# Display settings
display:
    hostname: "" # Override hostname (empty = auto-detect)

# Security settings
security:
    auth_token: "" # Optional Bearer token for API auth
    allowed_origins: # CORS origins (empty = allow all)
        - "http://192.168.1.*"

# Feature toggles
features:
    enable_shutdown: true
    enable_restart: true
    enable_hibernate: true
    enable_stats: true

# Stats collection settings
stats:
    gpu_enabled: true # Try to read NVIDIA GPU stats
    disk_cache_seconds: 30 # Cache disk stats for N seconds
```

---

## API Specification

### Base URL

```
http://{PC_IP}:{PORT}
```

### Endpoints

| Method | Endpoint            | Description       |
| ------ | ------------------- | ----------------- |
| GET    | `/rog/status`       | Health check      |
| GET    | `/rog/stats`        | Full system stats |
| GET    | `/rog/stats/memory` | Memory stats only |
| GET    | `/rog/stats/cpu`    | CPU stats only    |
| GET    | `/rog/stats/disk`   | Disk stats only   |
| POST   | `/rog/pw/shutdown`  | Shutdown PC       |
| POST   | `/rog/pw/restart`   | Restart PC        |
| POST   | `/rog/pw/hb`        | Hibernate PC      |

---

## Data Structures (Go)

These **must match exactly** what the Raspberry Pi dashboard expects:

```go
// ===== Memory Stats =====
type MemoryStats struct {
    Total       int64   `json:"total"`       // Total RAM in bytes
    Used        int64   `json:"used"`        // Used RAM in bytes
    Free        int64   `json:"free"`        // Free RAM in bytes
    UsedPercent float64 `json:"usedPercent"` // Usage percentage (e.g., 33.44)
}

// ===== CPU Stats =====
type CpuStats struct {
    Manufacturer  string  `json:"manufacturer"`  // e.g., "Intel"
    Brand         string  `json:"brand"`         // e.g., "Core™ i9-14900K"
    Cores         int     `json:"cores"`         // Total logical cores
    PhysicalCores int     `json:"physicalCores"` // Physical cores
    Speed         float64 `json:"speed"`         // Base speed in GHz
    CurrentLoad   float64 `json:"currentLoad"`   // Current CPU usage %
}

// ===== GPU Stats =====
type GpuStats struct {
    Vendor         string `json:"vendor"`                   // e.g., "NVIDIA"
    Model          string `json:"model"`                    // e.g., "NVIDIA GeForce RTX 4070 Ti SUPER"
    Vram           *int   `json:"vram,omitempty"`           // VRAM in MB
    VramUsed       *int   `json:"vramUsed,omitempty"`       // VRAM usage %
    TemperatureGpu *int   `json:"temperatureGpu,omitempty"` // Temperature in °C
    UtilizationGpu *int   `json:"utilizationGpu,omitempty"` // GPU usage %
}

// ===== Disk Stats =====
type DiskStats struct {
    Fs          string  `json:"fs"`          // e.g., "C:"
    Type        string  `json:"type"`        // e.g., "NTFS"
    Size        int64   `json:"size"`        // Total size in bytes
    Used        int64   `json:"used"`        // Used space in bytes
    Available   int64   `json:"available"`   // Available space in bytes
    UsedPercent float64 `json:"usedPercent"` // Usage percentage
    Mount       string  `json:"mount"`       // e.g., "C:"
}

// ===== Full System Stats =====
type SystemStats struct {
    Memory   MemoryStats  `json:"memory"`
    Cpu      CpuStats     `json:"cpu"`
    Gpu      *GpuStats    `json:"gpu"`    // null if no GPU detected
    Disks    []DiskStats  `json:"disks"`
    Uptime   int64        `json:"uptime"`   // System uptime in seconds
    Hostname string       `json:"hostname"` // e.g., "ROG-GT502"
    Platform string       `json:"platform"` // Always "win32" for compatibility
}
```

---

## Example API Response

### `GET /rog/stats`

```json
{
    "memory": {
        "total": 68422488064,
        "used": 22878642176,
        "free": 45543845888,
        "usedPercent": 33.44
    },
    "cpu": {
        "manufacturer": "Intel",
        "brand": "Core™ i9-14900K",
        "cores": 32,
        "physicalCores": 24,
        "speed": 3.2,
        "currentLoad": 6.5
    },
    "gpu": {
        "vendor": "NVIDIA",
        "model": "NVIDIA GeForce RTX 4070 Ti SUPER",
        "vram": 16376,
        "vramUsed": 15,
        "temperatureGpu": 35,
        "utilizationGpu": 10
    },
    "disks": [
        {
            "fs": "C:",
            "type": "NTFS",
            "size": 999248883712,
            "used": 838152736768,
            "available": 161096146944,
            "usedPercent": 83.88,
            "mount": "C:"
        },
        {
            "fs": "D:",
            "type": "NTFS",
            "size": 1000203087872,
            "used": 586261716992,
            "available": 413941370880,
            "usedPercent": 58.61,
            "mount": "D:"
        }
    ],
    "uptime": 3600,
    "hostname": "ROG-GT502",
    "platform": "win32"
}
```

### `GET /rog/status`

```json
{
    "status": "ok"
}
```

### `POST /rog/pw/shutdown`

```json
{
    "status": "ok"
}
```

---

## Implementation Details

### Memory Stats

Use Windows `GlobalMemoryStatusEx` syscall:

```go
import "golang.org/x/sys/windows"

func GetMemoryStats() (*MemoryStats, error) {
    var memStatus windows.MemoryStatusEx
    memStatus.Length = uint32(unsafe.Sizeof(memStatus))
    windows.GlobalMemoryStatusEx(&memStatus)

    return &MemoryStats{
        Total:       int64(memStatus.TotalPhys),
        Free:        int64(memStatus.AvailPhys),
        Used:        int64(memStatus.TotalPhys - memStatus.AvailPhys),
        UsedPercent: float64(memStatus.MemoryLoad),
    }, nil
}
```

### CPU Stats

-   **Brand/Model**: Read from registry `HKLM\HARDWARE\DESCRIPTION\System\CentralProcessor\0`
-   **Cores**: Use `runtime.NumCPU()` for logical cores
-   **Current Load**: Use PDH (Performance Data Helper) counter `\Processor(_Total)\% Processor Time`

### GPU Stats

Execute `nvidia-smi` with windowsHide flag:

```go
cmd := exec.Command("nvidia-smi",
    "--query-gpu=utilization.gpu,utilization.memory,temperature.gpu,name,memory.total",
    "--format=csv,noheader,nounits")
cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
```

### Disk Stats

Use `GetDiskFreeSpaceEx` for each drive letter:

```go
import "golang.org/x/sys/windows"

func GetDiskStats() ([]DiskStats, error) {
    drives, _ := windows.GetLogicalDriveStrings(...)
    for _, drive := range drives {
        var freeBytesAvailable, totalBytes, totalFreeBytes uint64
        windows.GetDiskFreeSpaceEx(drive, &freeBytesAvailable, &totalBytes, &totalFreeBytes)
        // Build DiskStats...
    }
}
```

### Power Commands

```go
func Shutdown() error {
    return exec.Command("shutdown", "/s", "/t", "0").Run()
}

func Restart() error {
    return exec.Command("shutdown", "/r", "/t", "0").Run()
}

func Hibernate() error {
    return exec.Command("shutdown", "/h").Run()
}
```

### Windows Service

Use `golang.org/x/sys/windows/svc`:

```go
import "golang.org/x/sys/windows/svc"

type pcControllerService struct{}

func (s *pcControllerService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
    changes <- svc.Status{State: svc.StartPending}

    // Start HTTP server in goroutine
    go startServer()

    changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

    // Wait for stop signal
    for c := range r {
        switch c.Cmd {
        case svc.Stop, svc.Shutdown:
            changes <- svc.Status{State: svc.StopPending}
            return false, 0
        }
    }
    return false, 0
}
```

---

## Dependencies (go.mod)

```go
module github.com/yourusername/pc-controller

go 1.22

require (
    github.com/go-chi/chi/v5 v5.0.11    // HTTP router
    github.com/go-chi/cors v1.2.1       // CORS middleware
    golang.org/x/sys v0.16.0            // Windows syscalls
    gopkg.in/yaml.v3 v3.0.1             // YAML config parsing
)
```

---

## Build Instructions

### Prerequisites

-   Go 1.22+
-   Windows SDK (for syscalls)

### Build Commands

```bash
# Development build
go build -o pc-controller.exe ./cmd/pc-controller

# Production build (smaller binary, no debug symbols)
go build -ldflags="-s -w" -o pc-controller.exe ./cmd/pc-controller

# Cross-compile from Linux/Mac (for Windows)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o pc-controller.exe ./cmd/pc-controller
```

### Makefile

```makefile
.PHONY: build clean install

VERSION := 1.0.0
LDFLAGS := -ldflags="-s -w -X main.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o pc-controller.exe ./cmd/pc-controller

clean:
	rm -f pc-controller.exe

install: build
	./pc-controller.exe install

uninstall:
	./pc-controller.exe uninstall
```

---

## Installation Flow

1. **Download** `pc-controller.exe` to any folder
2. **Run as Admin**: `pc-controller.exe install`
    - Creates Windows service "PC Controller"
    - Auto-creates `config.yaml` with defaults
3. **Start**: `pc-controller.exe start`
4. **Verify**: Open `http://localhost:8080/rog/status` in browser

---

## Optional: Installer (Inno Setup)

Create `installer.iss` for proper Windows installer:

```iss
[Setup]
AppName=PC Controller
AppVersion=1.0.0
DefaultDirName={pf}\PC Controller
OutputBaseFilename=pc-controller-setup
PrivilegesRequired=admin

[Files]
Source: "pc-controller.exe"; DestDir: "{app}"
Source: "config.example.yaml"; DestDir: "{app}"; DestName: "config.yaml"

[Run]
Filename: "{app}\pc-controller.exe"; Parameters: "install"; Flags: runhidden

[UninstallRun]
Filename: "{app}\pc-controller.exe"; Parameters: "uninstall"; Flags: runhidden
```

---

## Testing Checklist

-   [ ] `pc-controller.exe run` starts HTTP server
-   [ ] `GET /rog/status` returns `{"status":"ok"}`
-   [ ] `GET /rog/stats` returns full system info
-   [ ] `GET /rog/stats/memory` returns memory only
-   [ ] `GET /rog/stats/cpu` returns CPU only
-   [ ] `GET /rog/stats/disk` returns disks only
-   [ ] `POST /rog/pw/shutdown` triggers shutdown
-   [ ] `POST /rog/pw/restart` triggers restart
-   [ ] `POST /rog/pw/hb` triggers hibernate
-   [ ] `pc-controller.exe install` creates service (as admin)
-   [ ] `pc-controller.exe start` starts service
-   [ ] `pc-controller.exe stop` stops service
-   [ ] `pc-controller.exe uninstall` removes service
-   [ ] Service auto-starts after reboot
-   [ ] Config file changes take effect after restart

---

## Dashboard Compatibility

The Raspberry Pi dashboard expects these exact endpoints and response shapes. Do not modify:

| Dashboard Call | Service Endpoint        | Notes                     |
| -------------- | ----------------------- | ------------------------- |
| Health check   | `GET /rog/status`       | Returns `{"status":"ok"}` |
| System stats   | `GET /rog/stats`        | Full SystemStats JSON     |
| Shutdown       | `POST /rog/pw/shutdown` | Returns `{"status":"ok"}` |
| Restart        | `POST /rog/pw/restart`  | Returns `{"status":"ok"}` |
| Hibernate      | `POST /rog/pw/hb`       | Returns `{"status":"ok"}` |

The `platform` field in stats should always be `"win32"` for consistency with the Node.js version.
