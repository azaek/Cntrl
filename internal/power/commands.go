package power

import (
	"os/exec"
	"syscall"
)

// Shutdown initiates a system shutdown
func Shutdown() error {
	cmd := exec.Command("shutdown", "/s", "/t", "0")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// Restart initiates a system restart
func Restart() error {
	cmd := exec.Command("shutdown", "/r", "/t", "0")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// Hibernate puts the system into hibernation
func Hibernate() error {
	cmd := exec.Command("shutdown", "/h")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}
