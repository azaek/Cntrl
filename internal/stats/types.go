package stats

// MemoryStats represents system memory information
type MemoryStats struct {
	Total       int64   `json:"total"`       // Total RAM in bytes
	Used        int64   `json:"used"`        // Used RAM in bytes
	Free        int64   `json:"free"`        // Free RAM in bytes
	UsedPercent float64 `json:"usedPercent"` // Usage percentage
}

// CpuStats represents CPU information
type CpuStats struct {
	Manufacturer  string  `json:"manufacturer"`  // e.g., "Intel"
	Brand         string  `json:"brand"`         // e.g., "Core™ i9-14900K"
	Cores         int     `json:"cores"`         // Total logical cores
	PhysicalCores int     `json:"physicalCores"` // Physical cores
	Speed         float64 `json:"speed"`         // Base speed in GHz
	CurrentLoad   float64 `json:"currentLoad"`   // Current CPU usage %
}

// GpuStats represents GPU information
type GpuStats struct {
	Vendor         string `json:"vendor"`                   // e.g., "NVIDIA", "AMD", "Intel"
	Model          string `json:"model"`                    // e.g., "NVIDIA GeForce RTX 4070 Ti SUPER"
	Vram           *int   `json:"vram,omitempty"`           // VRAM in MB
	VramUsed       *int   `json:"vramUsed,omitempty"`       // VRAM usage %
	TemperatureGpu *int   `json:"temperatureGpu,omitempty"` // Temperature in °C
	UtilizationGpu *int   `json:"utilizationGpu,omitempty"` // GPU usage %
}

// DiskStats represents disk/volume information
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
type SystemStats struct {
	Memory   MemoryStats `json:"memory"`
	Cpu      CpuStats    `json:"cpu"`
	Gpu      *GpuStats   `json:"gpu"` // null if no GPU detected
	Disks    []DiskStats `json:"disks"`
	Uptime   int64       `json:"uptime"`   // System uptime in seconds
	Hostname string      `json:"hostname"` // e.g., "ROG-GT502"
	Platform string      `json:"platform"` // Always "win32" for compatibility
}
