package api

import (
	"net/http"
	"time"

	"github.com/azaek/cntrl/internal/config"
	"github.com/azaek/cntrl/internal/power"
)

// PowerHandler handles power-related endpoints
type PowerHandler struct {
	cfg *config.Config
}

// NewPowerHandler creates a new PowerHandler
func NewPowerHandler(cfg *config.Config) *PowerHandler {
	return &PowerHandler{cfg: cfg}
}

// Shutdown handles POST /api/pw/shutdown
func (h *PowerHandler) Shutdown(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableShutdown {
		writeError(w, http.StatusForbidden, "Shutdown feature is disabled")
		return
	}

	// Send response first so the client knows it was received
	writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})

	go func() {
		time.Sleep(500 * time.Millisecond)
		power.Shutdown()
	}()
}

// Restart handles POST /api/pw/restart
func (h *PowerHandler) Restart(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableRestart {
		writeError(w, http.StatusForbidden, "Restart feature is disabled")
		return
	}

	// Send response first
	writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})

	// Trigger restart in background
	go func() {
		time.Sleep(500 * time.Millisecond)
		power.Restart()
	}()
}

// Hibernate handles POST /api/pw/hb
func (h *PowerHandler) Hibernate(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableHibernate {
		writeError(w, http.StatusForbidden, "Hibernate feature is disabled")
		return
	}

	// Send response first
	writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})

	// Trigger hibernate in background
	go func() {
		time.Sleep(500 * time.Millisecond)
		power.Hibernate()
	}()
}
