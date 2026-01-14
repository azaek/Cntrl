package main

import (
	"context"
	"embed"
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
	"golang.org/x/sys/windows"

	"github.com/azaek/cntrl/internal/api"
	"github.com/azaek/cntrl/internal/config"
)

// Global mutex handle to keep it alive for the duration of the app
var mutexHandle windows.Handle

//go:embed assets/*
var assets embed.FS

// Version is set at build time
var Version = "dev"

var (
	server     *http.Server
	serverDone chan struct{}
)

// Tray menu items
var (
	mStatus    *systray.MenuItem
	mGitHub    *systray.MenuItem
	mConfig    *systray.MenuItem
	mFeatures  *systray.MenuItem
	mStats     *systray.MenuItem
	mShutdown  *systray.MenuItem
	mRestart   *systray.MenuItem
	mSleep     *systray.MenuItem
	mHibernate *systray.MenuItem
	mStartup   *systray.MenuItem
	mQuit      *systray.MenuItem

	appConfig *config.Config
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

	// Single instance check using Windows Mutex
	mutexNamePtr, _ := windows.UTF16PtrFromString("Local\\" + config.AppName + "_Lock")
	var err error
	mutexHandle, err = windows.CreateMutex(nil, false, mutexNamePtr)
	if err != nil {
		// Failed to create mutex
		return
	}
	if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
		showError(fmt.Sprintf("%s is already running.", config.AppName))
		return
	}
	defer windows.CloseHandle(mutexHandle)

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

API available at: http://localhost:9990/api/stats

`, config.AppName, config.AppName, config.AppName, config.AppName)
}

func onTrayReady() {
	// Load config first
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Could not load config, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}
	appConfig = cfg

	// Set icon and title
	updateTrayIcon()
	systray.SetTitle(config.AppName)
	systray.SetTooltip(config.AppName)

	// Start the HTTP server with the shared config
	startServer(appConfig)

	// Status display
	port := 9990
	if appConfig != nil {
		port = appConfig.Server.Port
	}
	mStatus = systray.AddMenuItem(fmt.Sprintf("● Running on port %d", port), "Server is running")
	mStatus.Disable()

	systray.AddSeparator()

	// Quick actions
	mGitHub = systray.AddMenuItem("View on GitHub", "Open repository in browser")
	mConfig = systray.AddMenuItem("Open Config", "Edit configuration file")

	systray.AddSeparator()

	// Feature Toggles
	mFeatures = systray.AddMenuItem("Features", "Enable or disable features")
	mStats = mFeatures.AddSubMenuItem("Enable Stats", "Expose system statistics")
	mShutdown = mFeatures.AddSubMenuItem("Enable Shutdown", "Allow remote shutdown")
	mRestart = mFeatures.AddSubMenuItem("Enable Restart", "Allow remote restart")
	mSleep = mFeatures.AddSubMenuItem("Enable Sleep", "Allow remote sleep")
	mHibernate = mFeatures.AddSubMenuItem("Enable Hibernate", "Allow remote hibernation")

	// Set initial check states
	if appConfig.Features.EnableStats {
		mStats.Check()
	}
	if appConfig.Features.EnableShutdown {
		mShutdown.Check()
	}
	if appConfig.Features.EnableRestart {
		mRestart.Check()
	}
	if appConfig.Features.EnableSleep {
		mSleep.Check()
	}
	if appConfig.Features.EnableHibernate {
		mHibernate.Check()
	}

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
		case <-mGitHub.ClickedCh:
			openGitHub()
		case <-mConfig.ClickedCh:
			openConfigFile()
		case <-mStats.ClickedCh:
			toggleFeature("stats")
		case <-mShutdown.ClickedCh:
			toggleFeature("shutdown")
		case <-mRestart.ClickedCh:
			toggleFeature("restart")
		case <-mSleep.ClickedCh:
			toggleFeature("sleep")
		case <-mHibernate.ClickedCh:
			toggleFeature("hibernate")
		case <-mStartup.ClickedCh:
			toggleStartup()
		case <-mQuit.ClickedCh:
			systray.Quit()
		}
	}
}

// ============== HTTP SERVER ==============

func startServer(cfg *config.Config) {
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
			updateTrayIconWithError()
			if mStatus != nil {
				mStatus.SetTitle(fmt.Sprintf("● Error: %v", err))
				mStatus.SetTooltip(fmt.Sprintf("Failed to start server: %v", err))
			}
		}
		close(serverDone)
	}()
}

func toggleFeature(feature string) {
	if appConfig == nil {
		return
	}

	switch feature {
	case "stats":
		appConfig.Features.EnableStats = !appConfig.Features.EnableStats
		if appConfig.Features.EnableStats {
			mStats.Check()
		} else {
			mStats.Uncheck()
		}
	case "shutdown":
		appConfig.Features.EnableShutdown = !appConfig.Features.EnableShutdown
		if appConfig.Features.EnableShutdown {
			mShutdown.Check()
		} else {
			mShutdown.Uncheck()
		}
	case "restart":
		appConfig.Features.EnableRestart = !appConfig.Features.EnableRestart
		if appConfig.Features.EnableRestart {
			mRestart.Check()
		} else {
			mRestart.Uncheck()
		}
	case "sleep":
		appConfig.Features.EnableSleep = !appConfig.Features.EnableSleep
		if appConfig.Features.EnableSleep {
			mSleep.Check()
		} else {
			mSleep.Uncheck()
		}
	case "hibernate":
		appConfig.Features.EnableHibernate = !appConfig.Features.EnableHibernate
		if appConfig.Features.EnableHibernate {
			mHibernate.Check()
		} else {
			mHibernate.Uncheck()
		}
	}

	// Save and restart (routing is static, needs restart to apply toggles)
	if err := config.Save(appConfig); err != nil {
		showError(fmt.Sprintf("Failed to save config: %v", err))
	}

	// We don't restart server here automatically to avoid session disruption,
	// but routing is static so it won't take effect until restart.
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
	updateTrayIcon()
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

func openGitHub() {
	exec.Command("cmd", "/c", "start", config.AppURL).Start()
}

func openConfigFile() {
	configPath, err := config.GetConfigPath()
	if err != nil {
		showError(fmt.Sprintf("Could not get config path: %v", err))
		return
	}
	// Create if not exists
	config.CreateDefaultConfig()
	exec.Command("cmd", "/c", "start", configPath).Start()
}

func showError(message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxW := user32.NewProc("MessageBoxW")
	title, _ := syscall.UTF16PtrFromString(config.AppName)
	text, _ := syscall.UTF16PtrFromString(message)
	messageBoxW.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x10)
}

// ============== TRAY ICON ==============

func updateTrayIcon() {
	iconName := "default.ico"
	if isInStartup() {
		iconName = "startup.ico"
	}

	data, err := assets.ReadFile("assets/" + iconName)
	if err == nil {
		systray.SetIcon(data)
	}
}

func updateTrayIconWithError() {
	data, err := assets.ReadFile("assets/error.ico")
	if err == nil {
		systray.SetIcon(data)
	}
}
