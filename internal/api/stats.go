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

// GetFullStats handles GET /rog/stats
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

// GetMemoryStats handles GET /rog/stats/memory
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

// GetCpuStats handles GET /rog/stats/cpu
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

// GetDiskStats handles GET /rog/stats/disk
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
