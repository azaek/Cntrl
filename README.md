# Go PC Remote

A lightweight system tray app for Windows that exposes system stats and power control via HTTP.

## Features

-   **System Tray App** - Runs in background, easy to access
-   **System Stats** - CPU, Memory, GPU (NVIDIA/AMD/Intel), Disks
-   **Power Control** - Shutdown, Restart, Hibernate
-   **Run at Startup** - One-click toggle
-   **Single Binary** - ~7 MB, no dependencies

## Quick Start

```powershell
# Build
go build -ldflags="-s -w -H windowsgui" -o go-pc-rem.exe ./cmd/go-pc-rem

# Run
.\go-pc-rem.exe
```

The app appears in your system tray. Right-click to access the menu.

## Tray Menu

| Option             | Description           |
| ------------------ | --------------------- |
| ● Running          | Status indicator      |
| Open Dashboard     | Open stats in browser |
| Open Config Folder | Edit configuration    |
| Run at Startup     | Toggle auto-start     |
| Exit               | Stop and exit         |

## API Endpoints

Base URL: `http://localhost:9990`

| Endpoint                | Description       |
| ----------------------- | ----------------- |
| `GET /rog/status`       | Health check      |
| `GET /rog/stats`        | Full system stats |
| `GET /rog/stats/memory` | Memory only       |
| `GET /rog/stats/cpu`    | CPU only          |
| `GET /rog/stats/disk`   | Disks only        |
| `POST /rog/pw/shutdown` | Shutdown PC       |
| `POST /rog/pw/restart`  | Restart PC        |
| `POST /rog/pw/hb`       | Hibernate PC      |

## Configuration

Location: `%APPDATA%\go-pc-rem\config.yaml`

```yaml
server:
    port: 9990
    host: "0.0.0.0"

features:
    enable_shutdown: true
    enable_restart: true
    enable_hibernate: true
    enable_stats: true

stats:
    gpu_enabled: true
    disk_cache_seconds: 30
```

## How It Works

-   **App running** = HTTP server available
-   **App closed** = Everything stops
-   **No hidden services** - What you see is what you get

## Requirements

-   Windows 10/11
-   Go 1.22+ (for building)
