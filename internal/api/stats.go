package api

import (
	"net/http"

	"github.com/azaek/cntrl/internal/config"
	"github.com/azaek/cntrl/internal/stats"
)

// StatsHandler handles stats-related endpoints
type StatsHandler struct {
	cfg *config.Config
}

// NewStatsHandler creates a new StatsHandler
func NewStatsHandler(cfg *config.Config) *StatsHandler {
	// Configure disk cache TTL from config
	stats.SetDiskCacheTTL(cfg.Stats.DiskCacheSeconds)

	return &StatsHandler{cfg: cfg}
}

// GetFullStats handles GET /api/stats
// Deprecated: Use GetSystemInfo (/api/system) and GetSystemUsage (/api/usage) instead
func (h *StatsHandler) GetFullStats(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableStats {
		writeError(w, http.StatusForbidden, "Stats feature is disabled")
		return
	}
	sysStats, err := stats.GetSystemStats(h.cfg.Display.Hostname, h.cfg.Stats.GpuEnabled)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sysStats)
}

// GetSystemInfo handles GET /api/system (Static info)
func (h *StatsHandler) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableSystem {
		writeError(w, http.StatusForbidden, "System info feature is disabled")
		return
	}
	info, err := stats.GetSystemInfo(h.cfg.Display.Hostname, h.cfg.Stats.GpuEnabled)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, info)
}

// GetSystemUsage handles GET /api/usage (Dynamic usage)
func (h *StatsHandler) GetSystemUsage(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableUsage {
		writeError(w, http.StatusForbidden, "Usage data feature is disabled")
		return
	}
	usage, err := stats.GetSystemUsage(h.cfg.Stats.GpuEnabled)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, usage)
}

// GetMemoryStats handles GET /api/stats/memory
func (h *StatsHandler) GetMemoryStats(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableStats {
		writeError(w, http.StatusForbidden, "Stats feature is disabled")
		return
	}
	memStats, err := stats.GetMemoryStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, memStats)
}

// GetCpuStats handles GET /api/stats/cpu
func (h *StatsHandler) GetCpuStats(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableStats {
		writeError(w, http.StatusForbidden, "Stats feature is disabled")
		return
	}
	cpuStats, err := stats.GetCpuStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cpuStats)
}

// GetDiskStats handles GET /api/stats/disk
func (h *StatsHandler) GetDiskStats(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableStats {
		writeError(w, http.StatusForbidden, "Stats feature is disabled")
		return
	}
	diskStats, err := stats.GetDiskStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, diskStats)
}

// GetProcessList handles GET /api/processes
func (h *StatsHandler) GetProcessList(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableProcesses {
		writeError(w, http.StatusForbidden, "Processes feature is disabled")
		return
	}

	procs, err := stats.GetProcesses()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, procs)
}
