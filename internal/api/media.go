package api

import (
	"encoding/json"
	"net/http"

	"github.com/azaek/cntrl/internal/config"
	"github.com/azaek/cntrl/internal/media"
)

type MediaHandler struct {
	cfg *config.Config
}

func NewMediaHandler(cfg *config.Config) *MediaHandler {
	return &MediaHandler{cfg: cfg}
}

// ControlRequest body
type ControlRequest struct {
	Action string `json:"action"` // play, pause, next, prev
}

func (h *MediaHandler) HandleControl(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableMedia {
		writeError(w, http.StatusForbidden, "Media feature is disabled")
		return
	}

	var req ControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := media.Control(req.Action); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (h *MediaHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Features.EnableMedia {
		writeError(w, http.StatusForbidden, "Media feature is disabled")
		return
	}

	status, err := media.GetStatus()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}
