---
trigger: always_on
---

# Development Guidelines

This document establishes the rules, patterns, and workflows for the continued development of **Cntrl** (go-pc-rem).

## 1. Project Overview

**Cntrl** is a cross-platform system service and tray application for remote PC/Mac monitoring and control. It exposes a REST API for clients (like the RPi dashboard) to consume.

- **Go Module:** `github.com/azaek/cntrl`
- **Target OS:** Windows, macOS (Darwin)
- **Documentation Site:** `https://cntrl.azaek.xyz/`

## 2. Technology Stack

### Backend (Service)

- **Language:** Go 1.22+
- **Router:** `github.com/go-chi/chi/v5`
- **System Tray:** `github.com/getlantern/systray`
- **Config:** `gopkg.in/yaml.v3`
- **Key Libraries:** `golang.org/x/sys` (Syscalls)

### Documentation / Frontend

- **Framework:** Next.js 16 (App Router)
- **Styling:** Tailwind CSS 4
- **Content:** Fumadocs (MDX)
- **Runtime:** Node.js (Latest LTS recommended)

### Build Tools

- **Build System:** `make`
- **Installer (Windows):** Inno Setup 6 (`ISCC.exe`)
- **Bundle (macOS):** `Cntrl.app` structure

## 3. Architecture & Structure

```
.
├── cmd/                # Main entry points
│   └── cntrl/         # Main application (Service + Tray + CLI)
├── internal/           # Private application code
│   ├── api/           # API handlers and router setup
│   ├── config/        # Configuration loading and types
│   ├── power/         # Power management (commands_windows.go, commands_darwin.go)
│   └── stats/         # System monitoring (cpu_windows.go, cpu_darwin.go, etc.)
├── docs/               # Documentation site (Next.js)
│   ├── content/       # MDX documentation pages
│   └── src/           # Doc site source code
├── winres/             # Windows resources (Icons, manifest)
├── macos/              # macOS paths and bundle resources
├── Cntrl.iss           # Inno Setup installer script
└── Makefile            # Build orchestration
```

## 4. Coding Standards

### Cross-Platform Go

- **File Naming:** Use `_windows.go` and `_darwin.go` suffixes for OS-specific implementations.
- **Build Tags:** Ensure strict separation of OS-dependent logic types (e.g., `syscall` imports).
- **Common Interfaces:** Define interfaces in `types.go` or the package root (e.g., `internal/stats/types.go`) and implement them in the OS-specific files.

### General Go

- **Formatting:** standard `gofmt`.
- **Error Handling:** Return errors explicitly; use wrapping where appropriate.
- **Configuration:** Inject `*config.Config` into handlers/services.
- **REST API:**
    - Use `chi` for routing.
    - Prefix API routes with `/api`.
    - Return JSON responses.

### TypeScript / React (Docs)

- **Style:** Functional components, Hooks.
- **Styling:** Tailwind CSS utility classes.
- **Content:** Use MDX for documentation pages.

## 5. Development Workflow

### Build Commands

| Command          | Description                                           |
| ---------------- | ----------------------------------------------------- |
| `make build`     | Build the binary (creates `bin/Cntrl.exe` on Windows) |
| `make dev`       | Run the app in dev mode                               |
| `make tray`      | Build and launch the tray app                         |
| `make installer` | Generate the Windows installer                        |

> **Note:** The current Makefile is optimized for Windows development. macOS developers may need to run `go build -o bin/Cntrl ./cmd/cntrl` directly or adapt the Makefile.

### Documentation

- Navigate to `docs/` directory.
- Run `npm run dev` to start the docs site locally on port 3001.

## 6. Release Process

1.  **Version Bump:** Update `VERSION` variable in `Makefile`.
2.  **Build:**
    - **Windows:** Run `make build` and `make installer`.
    - **macOS:** Ensure `Cntrl.app` structure is valid.
3.  **Test:** verify functionality on target platforms.
4.  **Commit:** Commit changes with message `chore: release vX.Y.Z`.
