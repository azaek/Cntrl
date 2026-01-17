//go:build darwin

package power

import "os/exec"

// Shutdown initiates a system shutdown using AppleScript
func Shutdown() error {
	return exec.Command("osascript", "-e", `tell app "System Events" to shut down`).Run()
}

// Restart initiates a system restart using AppleScript
func Restart() error {
	return exec.Command("osascript", "-e", `tell app "System Events" to restart`).Run()
}

// Sleep puts the system into sleep mode using pmset
func Sleep() error {
	return exec.Command("pmset", "sleepnow").Run()
}

// Hibernate puts the system into deep sleep (macOS equivalent of hibernate)
// Note: macOS doesn't have true hibernate like Windows, this triggers deep sleep
func Hibernate() error {
	// On macOS, we can use pmset to trigger sleep
	// For deeper hibernation, the system handles it automatically based on power settings
	return exec.Command("pmset", "sleepnow").Run()
}
