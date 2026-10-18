// Package result - Authored by Usama (2023331088)
package result

import (
	"net/http"
	"strconv"

	"sust-cse-portal/internal/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GET /api/v1/results/transcript/{reg}
func (h *Handler) GetTranscript(w http.ResponseWriter, r *http.Request) {
	reg := chi.URLParam(r, "reg")
	if reg == "" {
		reg = "2023331020" // default Afif
	}

	transcript, found := h.service.GetTranscriptByRegistration(reg)
	if !found {
		utils.Error(w, http.StatusNotFound, "Transcript records not found", nil)
		return
	}

	utils.Success(w, http.StatusOK, "Academic transcript records retrieved", transcript)
}

// GET /api/v1/results/simulate?reg=2023331020&target=3.80&credits=22.5
func (h *Handler) SimulateTargetGPA(w http.ResponseWriter, r *http.Request) {
	reg := r.URL.Query().Get("reg")
	if reg == "" {
		reg = "2023331020"
	}

	targetStr := r.URL.Query().Get("target")
	target := 3.80
	if targetStr != "" {
		if parsed, err := strconv.ParseFloat(targetStr, 64); err == nil {
			target = parsed
		}
	}

	creditsStr := r.URL.Query().Get("credits")
	credits := 22.5
	if creditsStr != "" {
		if parsed, err := strconv.ParseFloat(creditsStr, 64); err == nil {
			credits = parsed
		}
	}

	sim := h.service.SimulateTargetGPA(reg, target, credits)
	utils.Success(w, http.StatusOK, "GPA projection simulated", sim)
}
