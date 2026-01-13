package api

import (
	"net/http"

	"go-pc-rem/internal/config"
	"go-pc-rem/internal/power"
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
	if err := power.Shutdown(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}

// Restart handles POST /rog/pw/restart
func (h *PowerHandler) Restart(w http.ResponseWriter, r *http.Request) {
	if err := power.Restart(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}

// Hibernate handles POST /rog/pw/hb
func (h *PowerHandler) Hibernate(w http.ResponseWriter, r *http.Request) {
	if err := power.Hibernate(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}
