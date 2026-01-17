//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/azaek/cntrl/internal/config"
)

var pidFilePath string

func init() {
	// Set up PID file path in user's Application Support
	homeDir, _ := os.UserHomeDir()
	pidFilePath = filepath.Join(homeDir, "Library", "Application Support", config.AppName, "cntrl.pid")
}

// checkSingleInstance ensures only one instance of the app is running using a PID file
func checkSingleInstance() bool {
	// Ensure directory exists
	dir := filepath.Dir(pidFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return true // Can't create dir, allow running
	}

	// Check if PID file exists and if process is still running
	if data, err := os.ReadFile(pidFilePath); err == nil {
		var pid int
		if _, err := fmt.Sscanf(string(data), "%d", &pid); err == nil {
			// Check if process is still running
			process, err := os.FindProcess(pid)
			if err == nil {
				// On Unix, FindProcess always succeeds, need to send signal 0 to check
				if err := process.Signal(os.Signal(nil)); err == nil {
					showError(fmt.Sprintf("%s is already running.", config.AppName))
					return false
				}
			}
		}
	}

	// Write our PID
	if err := os.WriteFile(pidFilePath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		return true // Can't write PID, allow running
	}

	return true
}

// cleanupSingleInstance removes the PID file
func cleanupSingleInstance() {
	if pidFilePath != "" {
		os.Remove(pidFilePath)
	}
}

// hideConsoleWindow is a no-op on macOS
func hideConsoleWindow() {
	// Not needed on macOS
}

// showError displays an error dialog using osascript
func showError(message string) {
	script := fmt.Sprintf(`display dialog "%s" with title "%s" buttons {"OK"} default button "OK" with icon stop`, message, config.AppName)
	exec.Command("osascript", "-e", script).Run()
}

// getStartupPath returns the LaunchAgent plist path
func getStartupPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Library", "LaunchAgents", "com.azaek.cntrl.plist")
}

// isInStartup checks if the LaunchAgent is installed
func isInStartup() bool {
	_, err := os.Stat(getStartupPath())
	return err == nil
}

// addToStartup creates a LaunchAgent plist for auto-start
func addToStartup() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exePath, _ = filepath.Abs(exePath)

	plistPath := getStartupPath()

	// Ensure LaunchAgents directory exists
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return err
	}

	// Create LaunchAgent plist
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.azaek.cntrl</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>
`, exePath)

	return os.WriteFile(plistPath, []byte(plist), 0644)
}

// removeFromStartup removes the LaunchAgent plist
func removeFromStartup() error {
	return os.Remove(getStartupPath())
}

// openURL opens a URL in the default browser
func openURL(url string) {
	exec.Command("open", url).Start()
}

// openFile opens a file with the default application
func openFile(path string) {
	exec.Command("open", path).Start()
}

// getTrayTitle returns the title to show in the system tray
// On macOS, return empty string to only show icon (no text beside it)
func getTrayTitle() string {
	return ""
}
