![Cover](https://github.com/azaek/cntrl/blob/main/github.png)

[![Go Report Card](https://goreportcard.com/badge/github.com/azaek/cntrl)](https://goreportcard.com/report/github.com/azaek/cntrl)
![License](https://img.shields.io/github/license/azaek/cntrl)
![Latest Release](https://img.shields.io/github/v/release/azaek/cntrl?include_prereleases)
![Ai-assited](https://img.shields.io/badge/AI--assisted-262626)

Cntrl is a lightweight remote management bridge for Windows and macOS. It exposes your PC's hardware statistics and power controls through a simple, high-performance HTTP API, making it a first-class citizen in your homelab or remote monitoring dashboard.

## Features

-   **Remote Hardware Monitoring** - Instant access to CPU, Memory, GPU (NVIDIA/AMD/Intel), and Disk stats.
-   **Power Management** - Perform Shutdown, Restart, Sleep, or Hibernate actions remotely.
-   **Silent & Passive** - Runs quietly in the system tray with a minimal memory footprint (~7 MB).
-   **Reactive Branding** - Dynamic tray icon provides instant visual feedback on server status and errors.
-   **Zero Dependencies** - Single-binary architecture with no external runtimes or background services.
-   **Ready for Startup** - Easy one-click toggle to launch with Windows/macOS.
-   **Cross-Platform** - Native support for Windows and macOS (Intel & Apple Silicon).

## Documentation 📖

Check out our full documentation site at **[cntrl.azaek.xyz](https://cntrl.azaek.xyz/)**.

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

GoReleaser uses your **Git Tag** as the version number. It will generate both a **Portable EXE** and an **Interactive Installer**.

```powershell
# 1. Tag your release
git tag -a v0.0.23-beta -m "First beta release"

# 2. Run GoReleaser (Builds and packages everything)
# This will generate:
# - cntrl_<version>_<arch>_installer.exe
# - cntrl_<version>_<arch>_portable.zip
goreleaser release --clean
```

#### 2. Local Build (Testing)

If you want to build a binary quickly without tagging:

```powershell
# Using GoReleaser (recommended)
goreleaser build --snapshot --clean --single-target

# Or via Standard Go (Manual)
go-winres make --in winres/winres.json --out cmd/cntrl/
go build -ldflags="-s -w -H windowsgui -X main.Version=0.0.23-beta" -o bin/Cntrl.exe ./cmd/cntrl
```

The binary will be generated in `dist/` (for GoReleaser) or `bin/` (for Manual).

### Manual Installer (Optional)

Requires [Inno Setup 6](https://jrsoftware.org/isdl.php).

-   **Via GUI**: Right-click `Cntrl.iss` and select **Compile**.
-   **Via CLI**: `iscc Cntrl.iss`

### macOS Build 🍎

macOS builds require CGO (for system tray) and must be built on a Mac.

#### Build App Bundle

```bash
# Build Cntrl.app for Apple Silicon
./scripts/build-macos.sh 0.1.0 arm64

# Build Cntrl.app for Intel Mac
./scripts/build-macos.sh 0.1.0 amd64
```

The `.app` bundle will be created at `macos/Cntrl.app`.

#### Create DMG Installer

```bash
# Create DMG (run after building the app bundle)
./scripts/create-dmg.sh 0.1.0
```

The DMG will be created at `dist/Cntrl_0.1.0_macOS.dmg`.

#### Manual Build (without scripts)

```bash
# Build binary
CGO_ENABLED=1 go build -ldflags="-s -w -X main.Version=0.1.0" -o macos/Cntrl.app/Contents/MacOS/Cntrl ./cmd/cntrl

# Test
open macos/Cntrl.app

# Install
cp -r macos/Cntrl.app /Applications/
```

## Versioning 🏁

Cntrl uses **Git Tags** for official releases. For manual builds, update:

-   **`Makefile`**: `VERSION` variable.
-   **`Cntrl.iss`**: `#define MyAppVersion`.
-   **`winres/winres.json`**: All version strings.

The app appears in your system tray. Right-click to access the menu.

## Tray Menu

| Option              | Description            |
| ------------------- | ---------------------- |
| ● Running on port X | Status & Port info     |
| View on GitHub      | Open project on GitHub |
| Open Config         | Edit configuration     |
| Features            | Toggle sub-menu        |
| Run at Startup      | Toggle auto-start      |
| Exit                | Stop and exit          |

## API Endpoints

Base URL: `http://localhost:9990`

| Endpoint                 | Description       |
| ------------------------ | ----------------- |
| `GET /api/status`        | Health check      |
| `GET /api/stats`         | Full system stats |
| `GET /api/stats/memory`  | Memory only       |
| `GET /api/stats/cpu`     | CPU only          |
| `GET /api/stats/disk`    | Disks only        |
| `POST /api/pw/shutdown`  | Shutdown PC       |
| `POST /api/pw/restart`   | Restart PC        |
| `POST /api/pw/sleep`     | Sleep PC          |
| `POST /api/pw/hibernate` | Hibernate PC      |

## Configuration ⚙️

The configuration is stored in your user's config directory:
- **Windows**: `%APPDATA%\Cntrl\config.yaml`
- **macOS**: `~/Library/Application Support/Cntrl/config.yaml`

-   **Interactive Setup**: The installer will ask for your preferred **Port** and **Features** (Stats, Power actions) during installation.
-   **Manual Edit**: Right-click the tray icon and select **Open Config** to edit the YAML file directly.
-   **Dynamic Updates**: Toggling features via the tray menu takes effect immediately without restart.

## How It Works

-   **App running** = HTTP server available
-   **App closed** = Everything stops
-   **No hidden services** - What you see is what you get

## Requirements

-   Windows 10/11 or macOS 10.13+
-   Go 1.22+ (for building)
-   Xcode Command Line Tools (for macOS builds)
