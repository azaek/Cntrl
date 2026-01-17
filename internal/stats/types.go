package stats

// MemoryUsage represents dynamic memory usage
type MemoryUsage struct {
	Used        int64   `json:"used"`        // Used RAM in bytes
	Free        int64   `json:"free"`        // Free RAM in bytes
	UsedPercent float64 `json:"usedPercent"` // Usage percentage
}

// MemoryInfo represents static memory information
type MemoryInfo struct {
	Total int64 `json:"total"` // Total RAM in bytes
}

// CpuUsage represents dynamic CPU usage
type CpuUsage struct {
	CurrentLoad  float64 `json:"currentLoad"`  // Current CPU usage %
	CurrentTemp  float64 `json:"currentTemp"`  // Current CPU temperature in °C
	CurrentSpeed float64 `json:"currentSpeed"` // Current CPU speed in GHz
}

// CpuInfo represents static CPU information
type CpuInfo struct {
	Manufacturer  string  `json:"manufacturer"`  // e.g., "Intel"
	Brand         string  `json:"brand"`         // e.g., "Core™ i9-14900K"
	Cores         int     `json:"cores"`         // Total logical cores
	PhysicalCores int     `json:"physicalCores"` // Physical cores
	BaseSpeed     float64 `json:"baseSpeed"`     // Base speed in GHz
}

// GpuUsage represents dynamic GPU usage
type GpuUsage struct {
	UtilizationGpu *int `json:"utilizationGpu,omitempty"` // GPU usage %
	TemperatureGpu *int `json:"temperatureGpu,omitempty"` // Temperature in °C
	VramUsed       *int `json:"vramUsed,omitempty"`       // VRAM usage %
}

// GpuInfo represents static GPU information
type GpuInfo struct {
	Vendor string `json:"vendor"`         // e.g., "NVIDIA", "AMD", "Intel"
	Model  string `json:"model"`          // e.g., "NVIDIA GeForce RTX 4070 Ti SUPER"
	Vram   *int   `json:"vram,omitempty"` // VRAM in MB
}

// DiskUsage represents dynamic disk usage
type DiskUsage struct {
	Fs          string  `json:"fs"`          // e.g., "C:"
	Used        int64   `json:"used"`        // Used space in bytes
	Available   int64   `json:"available"`   // Available space in bytes
	UsedPercent float64 `json:"usedPercent"` // Usage percentage
}

// DiskInfo represents static disk information
type DiskInfo struct {
	Fs    string `json:"fs"`    // e.g., "C:"
	Type  string `json:"type"`  // e.g., "NTFS"
	Size  int64  `json:"size"`  // Total size in bytes
	Mount string `json:"mount"` // e.g., "C:"
}

// SystemInfo represents static system information
type SystemInfo struct {
	Hostname string     `json:"hostname"`
	Platform string     `json:"platform"`
	Cpu      CpuInfo    `json:"cpu"`
	Gpu      *GpuInfo   `json:"gpu"` // null if no GPU detected
	Memory   MemoryInfo `json:"memory"`
	Disks    []DiskInfo `json:"disks"`
}

// SystemUsage represents dynamic system usage
type SystemUsage struct {
	Uptime int64       `json:"uptime"` // System uptime in seconds
	Cpu    CpuUsage    `json:"cpu"`
	Memory MemoryUsage `json:"memory"`
	Gpu    *GpuUsage   `json:"gpu"` // null if no GPU detected
	Disks  []DiskUsage `json:"disks"`
}

// --- Legacy Types (Deprecated) ---

// MemoryStats represents system memory information
// Deprecated: Use MemoryInfo and MemoryUsage instead
type MemoryStats struct {
	Total       int64   `json:"total"`       // Total RAM in bytes
	Used        int64   `json:"used"`        // Used RAM in bytes
	Free        int64   `json:"free"`        // Free RAM in bytes
	UsedPercent float64 `json:"usedPercent"` // Usage percentage
}

// CpuStats represents CPU information
// Deprecated: Use CpuInfo and CpuUsage instead
type CpuStats struct {
	Manufacturer  string  `json:"manufacturer"`  // e.g., "Intel"
	Brand         string  `json:"brand"`         // e.g., "Core™ i9-14900K"
	Cores         int     `json:"cores"`         // Total logical cores
	PhysicalCores int     `json:"physicalCores"` // Physical cores
	Speed         float64 `json:"speed"`         // Base speed in GHz
	CurrentLoad   float64 `json:"currentLoad"`   // Current CPU usage %
}

// GpuStats represents GPU information
// Deprecated: Use GpuInfo and GpuUsage instead
type GpuStats struct {
	Vendor         string `json:"vendor"`                   // e.g., "NVIDIA", "AMD", "Intel"
	Model          string `json:"model"`                    // e.g., "NVIDIA GeForce RTX 4070 Ti SUPER"
	Vram           *int   `json:"vram,omitempty"`           // VRAM in MB
	VramUsed       *int   `json:"vramUsed,omitempty"`       // VRAM usage %
	TemperatureGpu *int   `json:"temperatureGpu,omitempty"` // Temperature in °C
	UtilizationGpu *int   `json:"utilizationGpu,omitempty"` // GPU usage %
}

// DiskStats represents disk/volume information
// Deprecated: Use DiskInfo and DiskUsage instead
type DiskStats struct {
	Fs          string  `json:"fs"`          // e.g., "C:"
	Type        string  `json:"type"`        // e.g., "NTFS"
	Size        int64   `json:"size"`        // Total size in bytes
	Used        int64   `json:"used"`        // Used space in bytes
	Available   int64   `json:"available"`   // Available space in bytes
	UsedPercent float64 `json:"usedPercent"` // Usage percentage
	Mount       string  `json:"mount"`       // e.g., "C:"
}

// SystemStats represents the full system statistics
// Deprecated: Use SystemInfo and SystemUsage instead
type SystemStats struct {
	Memory   MemoryStats `json:"memory"`
	Cpu      CpuStats    `json:"cpu"`
	Gpu      *GpuStats   `json:"gpu"` // null if no GPU detected
	Disks    []DiskStats `json:"disks"`
	Uptime   int64       `json:"uptime"`   // System uptime in seconds
	Hostname string      `json:"hostname"` // e.g., "ROG-GT502"
	Platform string      `json:"platform"` // Always "win32" for compatibility
}

// Process represents a running system process (aggregated by name)
type Process struct {
	Name     string  `json:"name"`
	Count    int     `json:"count"`     // Number of instances with this name
	Memory   uint64  `json:"memory"`    // Total memory usage in bytes
	MemoryMB float64 `json:"memory_mb"` // Total memory in MB (human readable)
	CpuTime  float64 `json:"cpu_time"`  // Total CPU time used in seconds
}

// MediaStatus represents media playback capability/status
// Status is optional as it's hard to retrieve on some platforms logic-free
type MediaStatus struct {
	Playing      bool   `json:"playing"`      // Is currently playing
	Title        string `json:"title"`        // Track title
	Artist       string `json:"artist"`       // Artist name
	SupportsCtrl bool   `json:"supportsCtrl"` // If controls are available
}
