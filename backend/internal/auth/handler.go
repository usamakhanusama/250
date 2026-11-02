package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"sust-cse-portal/internal/utils"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service, validate: validator.New()}
}

// decodeAndValidate reads the JSON body into dst and runs the `validate:"..."` struct tags.
// DisallowUnknownFields rejects payloads with extra fields such as {"role":"admin"}.
func (h *Handler) decodeAndValidate(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return false
	}
	if err := h.validate.Struct(dst); err != nil {
		utils.Error(w, http.StatusUnprocessableEntity, "Validation error", formatValidationErrors(err))
		return false
	}
	return true
}

// POST /api/v1/auth/signup (public, rate limited)
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}

	res, err := h.service.Signup(r.Context(), req)
	switch {
	case err == nil:
		utils.Success(w, http.StatusCreated, "Account created successfully", res)
	case errors.Is(err, ErrRegNoExists), errors.Is(err, ErrEmailExists):
		utils.Error(w, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, ErrInvalidRegNo), errors.Is(err, ErrWeakPassword),
		errors.Is(err, ErrInvalidName), errors.Is(err, ErrPasswordsDiffer):
		utils.Error(w, http.StatusUnprocessableEntity, err.Error(), nil)
	default:
		internalError(w, "signup", err)
	}
}

// POST /api/v1/auth/login (public, rate limited)
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}

	res, err := h.service.Login(r.Context(), req)
	switch {
	case err == nil:
		utils.Success(w, http.StatusOK, "Login successful", res)
	case errors.Is(err, ErrAccountLocked):
		utils.Error(w, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, ErrInvalidCredentials):
		utils.Error(w, http.StatusUnauthorized, err.Error(), nil)
	default:
		internalError(w, "login", err)
	}
}

// POST /api/v1/auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	res, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, ErrInvalidRefreshToken.Error(), nil)
		return
	}
	utils.Success(w, http.StatusOK, "Token refreshed successfully", res)
}

// POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	_ = h.service.Logout(r.Context(), req.RefreshToken)
	utils.Success(w, http.StatusOK, "Logged out successfully", nil)
}

// GET /api/v1/auth/me (requires JWT)
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	user, err := h.service.Me(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User no longer exists", nil)
		return
	}
	utils.Success(w, http.StatusOK, "Authenticated profile retrieved", user)
}

// POST /api/v1/auth/change-password (requires JWT)
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)

	var req ChangePasswordRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}

	err := h.service.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	switch {
	case err == nil:
		utils.Success(w, http.StatusOK, "Password updated successfully. Please login again.", nil)
	case errors.Is(err, ErrWrongOldPassword), errors.Is(err, ErrWeakPassword):
		utils.Error(w, http.StatusBadRequest, err.Error(), nil)
	default:
		internalError(w, "change-password", err)
	}
}

// internalError logs the real error on the server but sends only a generic
// message to the client, so database details never leak to the browser.
func internalError(w http.ResponseWriter, action string, err error) {
	log.Printf("auth %s error: %v", action, err)
	utils.Error(w, http.StatusInternalServerError, "Something went wrong. Please try again.", nil)
}

func formatValidationErrors(err error) []string {
	var errs []string
	var valErr validator.ValidationErrors
	if errors.As(err, &valErr) {
		for _, e := range valErr {
			errs = append(errs, e.Field()+": failed on '"+e.Tag()+"' condition")
		}
	} else {
		errs = append(errs, err.Error())
	}
	return errs
}
