package api

import (
	"net/http"

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

// Shutdown handles POST /rog/pw/shutdown
func (h *PowerHandler) Shutdown(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableShutdown {
		writeError(w, http.StatusForbidden, "Shutdown feature is disabled")
		return
	}
	if err := power.Shutdown(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}

// Restart handles POST /rog/pw/restart
func (h *PowerHandler) Restart(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableRestart {
		writeError(w, http.StatusForbidden, "Restart feature is disabled")
		return
	}
	if err := power.Restart(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}

// Hibernate handles POST /rog/pw/hb
func (h *PowerHandler) Hibernate(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableHibernate {
		writeError(w, http.StatusForbidden, "Hibernate feature is disabled")
		return
	}
	if err := power.Hibernate(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}
