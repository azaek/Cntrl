package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/getlantern/systray"

	"go-pc-rem/internal/api"
	"go-pc-rem/internal/config"
)

// Version is set at build time
var Version = "dev"

var (
	server     *http.Server
	serverDone chan struct{}
)

// Tray menu items
var (
	mStatus    *systray.MenuItem
	mDashboard *systray.MenuItem
	mConfig    *systray.MenuItem
	mStartup   *systray.MenuItem
	mQuit      *systray.MenuItem
)

func main() {
	// Check for CLI commands
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "version":
			fmt.Printf("%s version %s\n", config.AppName, Version)
			return
		case "help", "-h", "--help":
			printUsage()
			return
		}
	}

	// Default: run tray app with embedded server
	systray.Run(onTrayReady, onTrayExit)
}

func printUsage() {
	fmt.Printf(`%s - Remote PC control and monitoring

Usage:
  %s              Run the application (system tray)
  %s version      Show version info
  %s help         Show this help

The application runs as a system tray app with an embedded HTTP server.
Right-click the tray icon to access settings and dashboard.

API available at: http://localhost:9990/rog/stats

`, config.AppName, config.AppName, config.AppName, config.AppName)
}

func onTrayReady() {
	// Set icon and title
	systray.SetIcon(getTrayIcon())
	systray.SetTitle("Go PC Remote")
	systray.SetTooltip("Go PC Remote")

	// Start the HTTP server
	startServer()

	// Status display
	mStatus = systray.AddMenuItem("● Running", "Server is running")
	mStatus.Disable()

	systray.AddSeparator()

	// Quick actions
	mDashboard = systray.AddMenuItem("Open Dashboard", "Open stats in browser")
	mConfig = systray.AddMenuItem("Open Config Folder", "Open configuration folder")

	systray.AddSeparator()

	// Startup toggle
	mStartup = systray.AddMenuItem("Run at Startup", "Start automatically when Windows starts")
	if isInStartup() {
		mStartup.Check()
	}

	systray.AddSeparator()

	mQuit = systray.AddMenuItem("Exit", "Stop server and exit")

	// Handle menu clicks
	go handleClicks()
}

func onTrayExit() {
	// Stop the HTTP server gracefully
	stopServer()
}

func handleClicks() {
	for {
		select {
		case <-mDashboard.ClickedCh:
			openDashboard()
		case <-mConfig.ClickedCh:
			openConfigFolder()
		case <-mStartup.ClickedCh:
			toggleStartup()
		case <-mQuit.ClickedCh:
			systray.Quit()
		}
	}
}

// ============== HTTP SERVER ==============

func startServer() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Could not load config, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}

	router := api.NewRouter(cfg)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverDone = make(chan struct{})

	go func() {
		log.Printf("Starting HTTP server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
		close(serverDone)
	}()
}

func stopServer() {
	if server != nil {
		log.Println("Stopping HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		<-serverDone
		log.Println("Server stopped")
	}
}

// ============== STARTUP MANAGEMENT ==============

func getStartupShortcutPath() string {
	appData := os.Getenv("APPDATA")
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", config.AppName+".lnk")
}

func isInStartup() bool {
	shortcutPath := getStartupShortcutPath()
	_, err := os.Stat(shortcutPath)
	return err == nil
}

func toggleStartup() {
	if isInStartup() {
		// Remove from startup
		os.Remove(getStartupShortcutPath())
		mStartup.Uncheck()
	} else {
		// Add to startup
		if err := createShortcut(); err != nil {
			showError(fmt.Sprintf("Failed to add to startup: %v", err))
			return
		}
		mStartup.Check()
	}
}

func createShortcut() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exePath, _ = filepath.Abs(exePath)

	shortcutPath := getStartupShortcutPath()

	// Use PowerShell to create shortcut
	script := fmt.Sprintf(`
		$WshShell = New-Object -ComObject WScript.Shell
		$Shortcut = $WshShell.CreateShortcut('%s')
		$Shortcut.TargetPath = '%s'
		$Shortcut.WorkingDirectory = '%s'
		$Shortcut.Save()
	`, shortcutPath, exePath, filepath.Dir(exePath))

	cmd := exec.Command("powershell", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// ============== UTILITIES ==============

func openDashboard() {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}
	url := fmt.Sprintf("http://localhost:%d/rog/stats", cfg.Server.Port)
	exec.Command("cmd", "/c", "start", url).Start()
}

func openConfigFolder() {
	configPath, err := config.GetConfigPath()
	if err != nil {
		showError(fmt.Sprintf("Could not get config path: %v", err))
		return
	}
	configDir := filepath.Dir(configPath)
	os.MkdirAll(configDir, 0755)
	exec.Command("explorer", configDir).Start()
}

func showError(message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxW := user32.NewProc("MessageBoxW")
	title, _ := syscall.UTF16PtrFromString("Go PC Remote")
	text, _ := syscall.UTF16PtrFromString(message)
	messageBoxW.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x10)
}

// getTrayIcon returns the tray icon. It looks for icon.ico in the executable
// directory first, falling back to an embedded green icon if not found.
func getTrayIcon() []byte {
	// Try to load custom icon.ico from executable directory
	exePath, err := os.Executable()
	if err == nil {
		iconPath := filepath.Join(filepath.Dir(exePath), "icon.ico")
		if data, err := os.ReadFile(iconPath); err == nil {
			return data
		}
	}

	// Proper 16x16 32-bit ICO file with green color (#67FF90)
	// ICO structure: Header (6) + Directory Entry (16) + BITMAPINFOHEADER (40) + Pixels (1024) + Mask (64)
	ico := []byte{
		// ICO Header (6 bytes)
		0x00, 0x00, // Reserved
		0x01, 0x00, // Type: 1 = ICO
		0x01, 0x00, // Number of images: 1

		// Directory Entry (16 bytes)
		0x10,       // Width: 16
		0x10,       // Height: 16
		0x00,       // Color palette: 0 (no palette for 32-bit)
		0x00,       // Reserved
		0x01, 0x00, // Color planes: 1
		0x20, 0x00, // Bits per pixel: 32
		0x68, 0x04, 0x00, 0x00, // Image size: 1128 bytes (40 + 1024 + 64)
		0x16, 0x00, 0x00, 0x00, // Offset: 22 bytes

		// BITMAPINFOHEADER (40 bytes)
		0x28, 0x00, 0x00, 0x00, // Header size: 40
		0x10, 0x00, 0x00, 0x00, // Width: 16
		0x20, 0x00, 0x00, 0x00, // Height: 32 (doubled for ICO)
		0x01, 0x00, // Planes: 1
		0x20, 0x00, // Bits per pixel: 32
		0x00, 0x00, 0x00, 0x00, // Compression: none
		0x00, 0x04, 0x00, 0x00, // Image size: 1024
		0x00, 0x00, 0x00, 0x00, // X pixels per meter
		0x00, 0x00, 0x00, 0x00, // Y pixels per meter
		0x00, 0x00, 0x00, 0x00, // Colors used
		0x00, 0x00, 0x00, 0x00, // Important colors
	}

	// Add pixel data (16x16 BGRA, bottom-up)
	green := []byte{0x90, 0xFF, 0x67, 0xFF} // BGRA: #67FF90 fully opaque
	for i := 0; i < 256; i++ {              // 16 * 16 = 256 pixels
		ico = append(ico, green...)
	}

	// Add AND mask (16x16 1-bit, all zeros = fully visible)
	for i := 0; i < 64; i++ {
		ico = append(ico, 0x00)
	}

	return ico
}
