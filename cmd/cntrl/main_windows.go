//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/azaek/cntrl/internal/config"
)

// mutexHandle is the Windows mutex for single instance check
var mutexHandle windows.Handle

// checkSingleInstance ensures only one instance of the app is running
func checkSingleInstance() bool {
	mutexNamePtr, _ := windows.UTF16PtrFromString("Local\\" + config.AppName + "_Lock")
	var err error
	mutexHandle, err = windows.CreateMutex(nil, false, mutexNamePtr)
	if err != nil {
		return false
	}
	if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
		showError(fmt.Sprintf("%s is already running.", config.AppName))
		return false
	}
	return true
}

// cleanupSingleInstance releases the single instance mutex
func cleanupSingleInstance() {
	if mutexHandle != 0 {
		windows.CloseHandle(mutexHandle)
	}
}

// hideConsoleWindow hides console window on Windows (no-op, handled by -H windowsgui)
func hideConsoleWindow() {
	// Handled by ldflags -H windowsgui
}

// showError displays an error message box
func showError(message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxW := user32.NewProc("MessageBoxW")
	title, _ := syscall.UTF16PtrFromString(config.AppName)
	text, _ := syscall.UTF16PtrFromString(message)
	messageBoxW.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x10)
}

// getStartupPath returns the startup shortcut path for Windows
func getStartupPath() string {
	appData := os.Getenv("APPDATA")
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", config.AppName+".lnk")
}

// isInStartup checks if the app is configured to run at startup
func isInStartup() bool {
	shortcutPath := getStartupPath()
	_, err := os.Stat(shortcutPath)
	return err == nil
}

// addToStartup adds the app to Windows startup via shortcut
func addToStartup() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exePath, _ = filepath.Abs(exePath)

	shortcutPath := getStartupPath()

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

// removeFromStartup removes the app from Windows startup
func removeFromStartup() error {
	return os.Remove(getStartupPath())
}

// openURL opens a URL in the default browser
func openURL(url string) {
	exec.Command("cmd", "/c", "start", url).Start()
}

// openFile opens a file with the default application
func openFile(path string) {
	exec.Command("cmd", "/c", "start", path).Start()
}

// getTrayTitle returns the title to show in the system tray
// On Windows, this can show text (though typically just tooltip is used)
func getTrayTitle() string {
	return config.AppName
}
