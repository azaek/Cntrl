//go:build darwin

package stats

// On macOS, we use CLI tools (sysctl, vm_stat, etc.) instead of syscalls,
// so no global proc definitions are needed.
