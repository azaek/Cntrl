//go:build !darwin || cgo

package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getlantern/systray"

	"github.com/azaek/cntrl/internal/api"
	"github.com/azaek/cntrl/internal/config"
)

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
	mStatus       *systray.MenuItem
	mGitHub       *systray.MenuItem
	mConfig       *systray.MenuItem
	mReloadConfig *systray.MenuItem
	mFeatures     *systray.MenuItem
	mSystem       *systray.MenuItem
	mUsage        *systray.MenuItem
	mStatsLegacy  *systray.MenuItem
	mShutdown     *systray.MenuItem
	mRestart      *systray.MenuItem
	mSleep        *systray.MenuItem
	mHibernate    *systray.MenuItem
	mMedia        *systray.MenuItem
	mProcesses    *systray.MenuItem
	mStartup      *systray.MenuItem
	mQuit         *systray.MenuItem

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

	// Single instance check (platform-specific)
	if !checkSingleInstance() {
		return
	}
	defer cleanupSingleInstance()

	// Hide console window if needed (platform-specific)
	hideConsoleWindow()

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
	systray.SetTitle(getTrayTitle()) // Empty on macOS, app name on Windows
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
	mReloadConfig = systray.AddMenuItem("Reload Config", "Apply config changes (restarts server)")

	systray.AddSeparator()

	// Feature Toggles
	mFeatures = systray.AddMenuItem("Features", "Enable or disable features")
	mSystem = mFeatures.AddSubMenuItem("Enable System Info", "Static hardware info (/api/system)")
	mUsage = mFeatures.AddSubMenuItem("Enable Usage Data", "Dynamic usage metrics (/api/usage)")
	mStatsLegacy = mFeatures.AddSubMenuItem("Enable Stats (Legacy)", "Combined endpoint (/api/stats)")
	mShutdown = mFeatures.AddSubMenuItem("Enable Shutdown", "⚠️ Critical Action! Allows remote shutdown")
	mRestart = mFeatures.AddSubMenuItem("Enable Restart", "⚠️ Critical Action! Allows remote restart")
	mSleep = mFeatures.AddSubMenuItem("Enable Sleep", "Allow remote sleep")
	mHibernate = mFeatures.AddSubMenuItem("Enable Hibernate", "Allow remote hibernation")
	mMedia = mFeatures.AddSubMenuItem("Enable Media (Experimental)", "Media playback control")
	mProcesses = mFeatures.AddSubMenuItem("Enable Processes (Experimental)", "Process list endpoint")

	// Set initial check states
	if appConfig.Features.EnableSystem {
		mSystem.Check()
	}
	if appConfig.Features.EnableUsage {
		mUsage.Check()
	}
	if appConfig.Features.EnableStats {
		mStatsLegacy.Check()
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
	if appConfig.Features.EnableMedia {
		mMedia.Check()
	}
	if appConfig.Features.EnableProcesses {
		mProcesses.Check()
	}

	systray.AddSeparator()

	// Startup toggle
	mStartup = systray.AddMenuItem("Run at Startup", "Start automatically when system starts")
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
		case <-mReloadConfig.ClickedCh:
			reloadConfig()
		case <-mSystem.ClickedCh:
			toggleFeature("system")
		case <-mUsage.ClickedCh:
			toggleFeature("usage")
		case <-mStatsLegacy.ClickedCh:
			toggleFeature("stats")
		case <-mShutdown.ClickedCh:
			toggleFeature("shutdown")
		case <-mRestart.ClickedCh:
			toggleFeature("restart")
		case <-mSleep.ClickedCh:
			toggleFeature("sleep")
		case <-mHibernate.ClickedCh:
			toggleFeature("hibernate")
		case <-mMedia.ClickedCh:
			toggleFeature("media")
		case <-mProcesses.ClickedCh:
			toggleFeature("processes")
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

func reloadConfig() {
	log.Println("Reloading configuration...")

	// Show loading/error icon during reload
	updateTrayIconWithError()
	mStatus.SetTitle("⟳ Reloading...")

	// Stop current server
	stopServer()

	// Reload config from disk
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to reload config: %v", err)
		showError(fmt.Sprintf("Failed to reload config: %v", err))
		// Restart with old config
		startServer(appConfig)
		updateTrayIcon()
		mStatus.SetTitle(fmt.Sprintf("● Running on port %d", appConfig.Server.Port))
		return
	}

	// Update global config
	appConfig = cfg

	// Update menu check states
	updateMenuCheckStates()

	// Start server with new config
	startServer(appConfig)

	// Restore normal icon
	updateTrayIcon()
	mStatus.SetTitle(fmt.Sprintf("● Running on port %d", appConfig.Server.Port))

	log.Printf("Configuration reloaded. Server now running on port %d", appConfig.Server.Port)
}

func updateMenuCheckStates() {
	// System/Usage toggles
	if appConfig.Features.EnableSystem {
		mSystem.Check()
	} else {
		mSystem.Uncheck()
	}
	if appConfig.Features.EnableUsage {
		mUsage.Check()
	} else {
		mUsage.Uncheck()
	}
	if appConfig.Features.EnableStats {
		mStatsLegacy.Check()
	} else {
		mStatsLegacy.Uncheck()
	}

	// Power toggles
	if appConfig.Features.EnableShutdown {
		mShutdown.Check()
	} else {
		mShutdown.Uncheck()
	}
	if appConfig.Features.EnableRestart {
		mRestart.Check()
	} else {
		mRestart.Uncheck()
	}
	if appConfig.Features.EnableSleep {
		mSleep.Check()
	} else {
		mSleep.Uncheck()
	}
	if appConfig.Features.EnableHibernate {
		mHibernate.Check()
	} else {
		mHibernate.Uncheck()
	}

	// Experimental toggles
	if appConfig.Features.EnableMedia {
		mMedia.Check()
	} else {
		mMedia.Uncheck()
	}
	if appConfig.Features.EnableProcesses {
		mProcesses.Check()
	} else {
		mProcesses.Uncheck()
	}
}

func toggleFeature(feature string) {
	if appConfig == nil {
		return
	}

	switch feature {
	case "system":
		appConfig.Features.EnableSystem = !appConfig.Features.EnableSystem
		if appConfig.Features.EnableSystem {
			mSystem.Check()
		} else {
			mSystem.Uncheck()
		}
	case "usage":
		appConfig.Features.EnableUsage = !appConfig.Features.EnableUsage
		if appConfig.Features.EnableUsage {
			mUsage.Check()
		} else {
			mUsage.Uncheck()
		}
	case "stats":
		appConfig.Features.EnableStats = !appConfig.Features.EnableStats
		if appConfig.Features.EnableStats {
			mStatsLegacy.Check()
		} else {
			mStatsLegacy.Uncheck()
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
	case "media":
		appConfig.Features.EnableMedia = !appConfig.Features.EnableMedia
		if appConfig.Features.EnableMedia {
			mMedia.Check()
		} else {
			mMedia.Uncheck()
		}
	case "processes":
		appConfig.Features.EnableProcesses = !appConfig.Features.EnableProcesses
		if appConfig.Features.EnableProcesses {
			mProcesses.Check()
		} else {
			mProcesses.Uncheck()
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

func toggleStartup() {
	if isInStartup() {
		// Remove from startup
		removeFromStartup()
		mStartup.Uncheck()
	} else {
		// Add to startup
		if err := addToStartup(); err != nil {
			showError(fmt.Sprintf("Failed to add to startup: %v", err))
			return
		}
		mStartup.Check()
	}
	updateTrayIcon()
}

// ============== UTILITIES ==============

func openGitHub() {
	openURL(config.AppURL)
}

func openConfigFile() {
	configPath, err := config.GetConfigPath()
	if err != nil {
		showError(fmt.Sprintf("Could not get config path: %v", err))
		return
	}
	// Create if not exists
	config.CreateDefaultConfig()
	openFile(configPath)
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
