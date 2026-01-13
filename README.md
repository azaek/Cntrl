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
# Install dependencies and generate go.sum
go mod tidy
```

#### 1. Build the Binary (skip if already generated)

```powershell
# 1. Install go-winres (required for embedding the EXE icon)
go install github.com/tc-hib/go-winres@latest

# 2. Process resources (creates syso files)
go-winres make

# 3. Build (outputs to bin/Cntrl.exe)
go build -ldflags="-s -w -H windowsgui" -o bin/Cntrl.exe ./cmd/cntrl
```

#### 2. Create the Installer (Optional)

Requires [Inno Setup 6](https://jrsoftware.org/isdl.php).

-   **Via GUI**: Right-click `Cntrl.iss` and select **Compile**.
-   **Via CLI**: `iscc Cntrl.iss`

The installer will be generated in the **`dist/`** directory.

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
