// Package student - Authored by Usama (2023331088)
package student

import (
	"encoding/json"
	"net/http"

	"sust-cse-portal/internal/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GET /api/v1/students
func (h *Handler) GetAllStudents(w http.ResponseWriter, r *http.Request) {
	students := h.service.GetAllStudents()
	utils.Success(w, http.StatusOK, "Student directory retrieved", students)
}

// GET /api/v1/students/{reg}
func (h *Handler) GetStudentProfile(w http.ResponseWriter, r *http.Request) {
	reg := chi.URLParam(r, "reg")
	if reg == "" {
		utils.Error(w, http.StatusBadRequest, "Registration parameter is required", nil)
		return
	}

	st, found := h.service.GetStudentByRegistration(reg)
	if !found {
		utils.Error(w, http.StatusNotFound, "Student not found in batch roster", nil)
		return
	}

	utils.Success(w, http.StatusOK, "Student profile retrieved", st)
}

// GET /api/v1/students/analytics
func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics := h.service.GetBatchAnalytics()
	utils.Success(w, http.StatusOK, "Batch analytics retrieved", analytics)
}

// POST /api/v1/students/enroll
func (h *Handler) EnrollCourse(w http.ResponseWriter, r *http.Request) {
	var req CourseEnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid enrollment payload", nil)
		return
	}

	res := h.service.ValidateAndEnrollCourse(req.RegistrationNumber, req.CourseCode, req.Credits)
	utils.Success(w, http.StatusOK, "Enrollment validation evaluated", res)
}
