package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"sust-cse-portal/internal/utils"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

// Register handles user registration (Admin only)
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.Error(w, http.StatusUnprocessableEntity, "Validation error", formatValidationErrors(err))
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		if err == ErrEmailExists {
			utils.Error(w, http.StatusConflict, err.Error(), nil)
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.Success(w, http.StatusCreated, "User registered successfully", user)
}

// Login authenticates users with brute-force lockout
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.Error(w, http.StatusUnprocessableEntity, "Validation error", formatValidationErrors(err))
		return
	}

	res, err := h.service.Login(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "locked") {
			utils.Error(w, http.StatusForbidden, err.Error(), nil)
			return
		}
		utils.Error(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	utils.Success(w, http.StatusOK, "Login successful", res)
}

// Refresh handles token rotation
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.Error(w, http.StatusUnprocessableEntity, "Validation error", formatValidationErrors(err))
		return
	}

	res, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	utils.Success(w, http.StatusOK, "Token refreshed successfully", res)
}

// Logout revokes refresh token
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	_ = h.service.Logout(r.Context(), req.RefreshToken)
	utils.Success(w, http.StatusOK, "Logged out successfully", nil)
}

// ChangePassword allows authenticated users to update their password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.Error(w, http.StatusUnprocessableEntity, "Validation error", formatValidationErrors(err))
		return
	}

	if err := h.service.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.Success(w, http.StatusOK, "Password updated successfully. Please login again.", nil)
}

// Me returns the currently authenticated user's ID and Role
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	role, _ := r.Context().Value("role").(string)

	utils.Success(w, http.StatusOK, "Authenticated profile retrieved", map[string]interface{}{
		"user_id": userID,
		"role":    role,
	})
}

func formatValidationErrors(err error) []string {
	var errs []string
	if valErr, ok := err.(validator.ValidationErrors); ok {
		for _, e := range valErr {
			errs = append(errs, e.Field()+": failed on '"+e.Tag()+"' condition")
		}
	} else {
		errs = append(errs, err.Error())
	}
	return errs
}
