package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sust-cse-portal/internal/config"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid registration number/email or password")
	ErrAccountLocked       = errors.New("account is temporarily locked due to too many failed login attempts")
	ErrEmailExists         = errors.New("this email is already registered")
	ErrRegNoExists         = errors.New("this registration number is already registered")
	ErrInvalidRefreshToken = errors.New("invalid or revoked refresh token")
	ErrWrongOldPassword    = errors.New("current password is incorrect")
)

// dummyHash is a real bcrypt hash used when the user does not exist.
// We still run bcrypt on it so "wrong user" and "wrong password" take the
// same time; otherwise an attacker could time responses to find valid reg numbers.
var dummyHash, _ = HashPassword("dummy-password-for-timing-1")

type Service struct {
	store Store
	cfg   *config.Config
}

func NewService(store Store, cfg *config.Config) *Service {
	return &Service{store: store, cfg: cfg}
}

// Signup: public student self-registration (only 2023331XXX reg numbers allowed).
// On success the student is logged in immediately (tokens are returned).
func (s *Service) Signup(ctx context.Context, req SignupRequest) (*LoginResponse, error) {
	normalizeSignup(&req)

	if err := ValidateRegNo(req.RegNo); err != nil {
		return nil, err
	}
	if len(req.FullName) < 3 || len(req.FullName) > 100 {
		return nil, ErrInvalidName
	}
	if req.Password != req.ConfirmPassword {
		return nil, ErrPasswordsDiffer
	}
	if err := ValidatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// Only the bcrypt hash is stored, never the plain password.
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.store.CreateStudent(ctx, req.Email, hash, req.RegNo, req.FullName)
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, user)
}

// Login checks the password with bcrypt and locks the account after too many failures.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.store.FindByLogin(ctx, req.Login)
	if errors.Is(err, ErrUserNotFound) {
		CheckPassword(dummyHash, req.Password) // same timing as a real check
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("looking up user: %w", err)
	}

	now := time.Now()

	// 1. Account lockout (per-account brute-force protection)
	if user.LockedUntil != nil && user.LockedUntil.After(now) {
		return nil, ErrAccountLocked
	}

	// 2. Compare password with stored bcrypt hash
	if !CheckPassword(user.PasswordHash, req.Password) {
		attempts := user.FailedAttempts + 1
		var lockUntil *time.Time
		if attempts >= s.cfg.MaxLoginAttempts {
			t := now.Add(s.cfg.LockoutDuration)
			lockUntil = &t
		}
		_ = s.store.SetLoginFailures(ctx, user.ID, attempts, lockUntil)
		if lockUntil != nil {
			return nil, ErrAccountLocked
		}
		// Same message as "user not found", so attackers can't tell which reg numbers exist.
		return nil, ErrInvalidCredentials
	}

	// 3. Success: reset the failure counter and issue JWTs
	_ = s.store.SetLoginFailures(ctx, user.ID, 0, nil)
	return s.issueTokens(ctx, user)
}

// Refresh exchanges a valid refresh token for a new pair (token rotation:
// the old refresh token is revoked so it can't be reused if stolen).
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	if _, err := ValidateToken(refreshToken, s.cfg.JWTRefreshSecret); err != nil {
		return nil, ErrInvalidRefreshToken
	}

	tokenHash := HashToken(refreshToken)
	rt, err := s.store.FindRefreshToken(ctx, tokenHash)
	if err != nil || rt.Revoked || rt.ExpiresAt.Before(time.Now()) {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.store.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	_ = s.store.RevokeRefreshToken(ctx, tokenHash)
	return s.issueTokens(ctx, user)
}

// Logout revokes the refresh token so it can no longer be used.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.store.RevokeRefreshToken(ctx, HashToken(refreshToken))
}

// Me returns the safe profile of the logged-in user.
func (s *Service) Me(ctx context.Context, userID string) (*UserResponse, error) {
	user, err := s.store.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := toUserResponse(user)
	return &res, nil
}

// ChangePassword updates the password and logs the user out everywhere.
func (s *Service) ChangePassword(ctx context.Context, userID, oldPass, newPass string) error {
	user, err := s.store.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}
	if !CheckPassword(user.PasswordHash, oldPass) {
		return ErrWrongOldPassword
	}
	if err := ValidatePasswordStrength(newPass); err != nil {
		return err
	}
	newHash, err := HashPassword(newPass)
	if err != nil {
		return err
	}
	return s.store.UpdatePasswordAndRevokeSessions(ctx, userID, newHash)
}

// issueTokens creates an access + refresh JWT and stores the refresh token's SHA-256 hash.
func (s *Service) issueTokens(ctx context.Context, user *UserRecord) (*LoginResponse, error) {
	regNo := ""
	if user.Student != nil {
		regNo = user.Student.RegNo
	}

	pair, refreshHash, err := GenerateTokenPair(user.ID, user.Role, regNo,
		s.cfg.JWTSecret, s.cfg.JWTRefreshSecret, s.cfg.JWTAccessExpiry, s.cfg.JWTRefreshExpiry)
	if err != nil {
		return nil, err
	}

	if err := s.store.SaveRefreshToken(ctx, user.ID, refreshHash, time.Now().Add(s.cfg.JWTRefreshExpiry)); err != nil {
		return nil, fmt.Errorf("saving refresh token: %w", err)
	}

	return &LoginResponse{
		User:         toUserResponse(user),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt,
	}, nil
}

// toUserResponse copies only the safe fields. PasswordHash is intentionally dropped here.
func toUserResponse(u *UserRecord) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		Student:   u.Student,
		CreatedAt: u.CreatedAt,
	}
}

// SeedDemoStudents adds the two project members to an empty MemoryStore so the
// demo works without PostgreSQL. Passwords are bcrypt-hashed here, on the server.
func SeedDemoStudents(ctx context.Context, store Store) {
	demo := []struct{ reg, name, email string }{
		{"2023331020", "Afif Zaman Khan", "afif@student.sust.edu"},
		{"2023331088", "Usama", "usama@student.sust.edu"},
	}
	for _, d := range demo {
		hash, err := HashPassword("Password@123")
		if err != nil {
			continue
		}
		_, _ = store.CreateStudent(ctx, d.email, hash, d.reg, d.name)
	}
}
