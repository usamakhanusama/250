package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sust-cse-portal/internal/config"
	"sust-cse-portal/internal/db"

	"github.com/jackc/pgx/v5"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked     = errors.New("account is locked due to too many failed login attempts")
	ErrEmailExists       = errors.New("email is already registered")
	ErrInvalidRefreshToken = errors.New("invalid or revoked refresh token")
)

type Service struct {
	db  *db.DB
	cfg *config.Config
}

func NewService(database *db.DB, cfg *config.Config) *Service {
	return &Service{
		db:  database,
		cfg: cfg,
	}
}

// Register creates a new user account (Admin only)
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*UserResponse, error) {
	// Check if email already exists
	var exists bool
	err := s.db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed checking email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed hashing password: %w", err)
	}

	var user UserResponse
	err = s.db.Pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, role, created_at
	`, req.Email, hashedPassword, req.Role).Scan(&user.ID, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed creating user: %w", err)
	}

	return &user, nil
}

// Login authenticates a user with brute-force protection
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	var (
		userID        string
		passwordHash  string
		role          string
		failedAttempts int
		lockedUntil   *time.Time
		createdAt     time.Time
	)

	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, password_hash, role, failed_login_attempts, locked_until, created_at
		FROM users
		WHERE email = $1
	`, req.Email).Scan(&userID, &passwordHash, &role, &failedAttempts, &lockedUntil, &createdAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	now := time.Now()

	// 1. Check if account is currently locked
	if lockedUntil != nil && lockedUntil.After(now) {
		remaining := time.Until(*lockedUntil).Round(time.Minute)
		return nil, fmt.Errorf("%w: please try again after %v", ErrAccountLocked, remaining)
	}

	// 2. Validate password
	if !CheckPassword(passwordHash, req.Password) {
		failedAttempts++
		if failedAttempts >= s.cfg.MaxLoginAttempts {
			newLock := now.Add(s.cfg.LockoutDuration)
			_, _ = s.db.Pool.Exec(ctx, `
				UPDATE users
				SET failed_login_attempts = $1, locked_until = $2
				WHERE id = $3
			`, failedAttempts, newLock, userID)
			return nil, fmt.Errorf("%w: account locked for %v", ErrAccountLocked, s.cfg.LockoutDuration)
		}

		_, _ = s.db.Pool.Exec(ctx, `
			UPDATE users
			SET failed_login_attempts = $1
			WHERE id = $2
		`, failedAttempts, userID)

		remainingAttempts := s.cfg.MaxLoginAttempts - failedAttempts
		return nil, fmt.Errorf("%w (%d attempt(s) remaining before account lockout)", ErrInvalidCredentials, remainingAttempts)
	}

	// 3. Reset failed attempts on successful login
	_, _ = s.db.Pool.Exec(ctx, `
		UPDATE users
		SET failed_login_attempts = 0, locked_until = NULL
		WHERE id = $1
	`, userID)

	// 4. Generate JWT Token Pair
	tokenPair, tokenHash, err := GenerateTokenPair(
		userID,
		role,
		s.cfg.JWTSecret,
		s.cfg.JWTRefreshSecret,
		s.cfg.JWTAccessExpiry,
		s.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed generating tokens: %w", err)
	}

	// 5. Save hashed refresh token in database
	expiresAt := now.Add(s.cfg.JWTRefreshExpiry)
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed saving refresh token: %w", err)
	}

	userRes := UserResponse{
		ID:        userID,
		Email:     req.Email,
		Role:      role,
		CreatedAt: createdAt,
	}

	if role == "student" {
		var st StudentProfileBrief
		_ = s.db.Pool.QueryRow(ctx, "SELECT id, reg_no, full_name, current_semester FROM students WHERE user_id = $1", userID).Scan(&st.StudentID, &st.RegNo, &st.FullName, &st.CurrentSemester)
		if st.StudentID != "" {
			userRes.Student = &st
		}
	} else if role == "teacher" {
		var tp TeacherProfileBrief
		_ = s.db.Pool.QueryRow(ctx, "SELECT id, teacher_code, full_name FROM teachers WHERE user_id = $1", userID).Scan(&tp.TeacherID, &tp.TeacherCode, &tp.FullName)
		if tp.TeacherID != "" {
			userRes.Teacher = &tp
		}
	}

	return &LoginResponse{
		User:         userRes,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}, nil
}

// Refresh handles token rotation using valid refresh tokens
func (s *Service) Refresh(ctx context.Context, refreshTokenStr string) (*LoginResponse, error) {
	// Validate refresh token signature & claims
	_, err := ValidateToken(refreshTokenStr, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	tokenHash := HashToken(refreshTokenStr)

	// Verify token exists in database and is not revoked
	var (
		tokenID   string
		userID    string
		revoked   bool
		expiresAt time.Time
		role      string
		email     string
		createdAt time.Time
	)

	err = s.db.Pool.QueryRow(ctx, `
		SELECT rt.id, rt.user_id, rt.revoked, rt.expires_at, u.role, u.email, u.created_at
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token_hash = $1
	`, tokenHash).Scan(&tokenID, &userID, &revoked, &expiresAt, &role, &email, &createdAt)

	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if revoked || expiresAt.Before(time.Now()) {
		return nil, ErrInvalidRefreshToken
	}

	// Revoke the used refresh token (Token Rotation)
	_, _ = s.db.Pool.Exec(ctx, "UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1", tokenID)

	// Issue new token pair
	newTokens, newTokenHash, err := GenerateTokenPair(
		userID,
		role,
		s.cfg.JWTSecret,
		s.cfg.JWTRefreshSecret,
		s.cfg.JWTAccessExpiry,
		s.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed generating new tokens: %w", err)
	}

	// Save new refresh token hash
	newExpiresAt := time.Now().Add(s.cfg.JWTRefreshExpiry)
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, newTokenHash, newExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed storing new refresh token: %w", err)
	}

	userRes := UserResponse{
		ID:        userID,
		Email:     email,
		Role:      role,
		CreatedAt: createdAt,
	}

	if role == "student" {
		var st StudentProfileBrief
		_ = s.db.Pool.QueryRow(ctx, "SELECT id, reg_no, full_name, current_semester FROM students WHERE user_id = $1", userID).Scan(&st.StudentID, &st.RegNo, &st.FullName, &st.CurrentSemester)
		if st.StudentID != "" {
			userRes.Student = &st
		}
	} else if role == "teacher" {
		var tp TeacherProfileBrief
		_ = s.db.Pool.QueryRow(ctx, "SELECT id, teacher_code, full_name FROM teachers WHERE user_id = $1", userID).Scan(&tp.TeacherID, &tp.TeacherCode, &tp.FullName)
		if tp.TeacherID != "" {
			userRes.Teacher = &tp
		}
	}

	return &LoginResponse{
		User:         userRes,
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
		ExpiresAt:    newTokens.ExpiresAt,
	}, nil
}

// Logout revokes the given refresh token
func (s *Service) Logout(ctx context.Context, refreshTokenStr string) error {
	tokenHash := HashToken(refreshTokenStr)
	_, err := s.db.Pool.Exec(ctx, "UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1", tokenHash)
	return err
}

// ChangePassword updates user password and revokes all active refresh tokens
func (s *Service) ChangePassword(ctx context.Context, userID, oldPass, newPass string) error {
	var currentHash string
	err := s.db.Pool.QueryRow(ctx, "SELECT password_hash FROM users WHERE id = $1", userID).Scan(&currentHash)
	if err != nil {
		return ErrUserNotFound
	}

	if !CheckPassword(currentHash, oldPass) {
		return errors.New("current password is incorrect")
	}

	newHash, err := HashPassword(newPass)
	if err != nil {
		return err
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update password
	_, err = tx.Exec(ctx, "UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", newHash, userID)
	if err != nil {
		return err
	}

	// Revoke all existing sessions/refresh tokens for security
	_, err = tx.Exec(ctx, "UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
