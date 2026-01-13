![Cover](https://github.com/azaek/cntrl/blob/main/github.png)

[![Go Report Card](https://goreportcard.com/badge/github.com/azaek/cntrl)](https://goreportcard.com/report/github.com/azaek/cntrl)
![License](https://img.shields.io/github/license/azaek/cntrl)
![Latest Release](https://img.shields.io/github/v/release/azaek/cntrl?include_prereleases)

Cntrl is a lightweight remote management bridge for Windows. It exposes your PC's hardware statistics and power controls through a simple, high-performance HTTP API, making it a first-class citizen in your homelab or remote monitoring dashboard.

## Features

-   **Remote Hardware Monitoring** - Instant access to CPU, Memory, GPU (NVIDIA/AMD/Intel), and Disk stats.
-   **Power Management** - Perform Shutdown, Restart, or Hibernate actions remotely.
-   **Silent & Passive** - Runs quietly in the system tray with a minimal memory footprint (~7 MB).
-   **Reactive Branding** - Dynamic tray icon provides instant visual feedback on server status and errors.
-   **Zero Dependencies** - Single-binary architecture with no external runtimes or background services.
-   **Ready for Startup** - Easy one-click toggle to launch with Windows.

### Build from Source 🛠️

#### 0. Setup

```powershell
# 1. Install dependencies
go mod tidy

# 2. Install GoReleaser (The automated build engine)
go install github.com/goreleaser/goreleaser/v2@latest

# 3. Install go-winres (for icons)
go install github.com/tc-hib/go-winres@latest
```

#### 1. Build for Production (via Git Tag)

GoReleaser uses your **Git Tag** as the version number.

```powershell
# 1. Tag your release
git tag -a v0.0.23-beta -m "First beta release"

# 2. Run GoReleaser (Builds and packages everything)
goreleaser release --clean
```

#### 2. Local Build (Testing)

If you want to build a binary quickly without tagging:

```powershell
goreleaser build --snapshot --clean --single-target
```

The binary will be generated in `dist/cntrl_windows_amd64_v1/`.

The app appears in your system tray. Right-click to access the menu.

## Tray Menu

| Option              | Description           |
| ------------------- | --------------------- |
| ● Running on port X | Status & Port info    |
| Open Dashboard      | Open stats in browser |
| Open Config         | Edit configuration    |
| Features            | Toggle sub-menu       |
| Run at Startup      | Toggle auto-start     |
| Exit                | Stop and exit         |

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

## Configuration ⚙️

The configuration is stored in `%APPDATA%\Cntrl\config.yaml`.

-   **Interactive Setup**: The installer will ask for your preferred **Port** and **Features** (Stats, Power actions) during installation.
-   **Manual Edit**: Right-click the tray icon and select **Open Config** to edit the YAML file directly.
-   **Dynamic Updates**: Toggling features via the tray menu takes effect immediately without restart.

## How It Works

-   **App running** = HTTP server available
-   **App closed** = Everything stops
-   **No hidden services** - What you see is what you get

## Requirements

-   Windows 10/11
-   Go 1.22+ (for building)
