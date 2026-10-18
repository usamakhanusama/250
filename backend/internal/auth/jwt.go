package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrInvalidAlg   = errors.New("unexpected signing method")
)

// CustomClaims contains ONLY user_id and role as required by strict security rules
type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair holds both access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// GenerateTokenPair generates a 15-minute access token and a 7-day refresh token
func GenerateTokenPair(userID, role, accessSecret, refreshSecret string, accessExpiry, refreshExpiry time.Duration) (*TokenPair, string, error) {
	now := time.Now()
	accessExp := now.Add(accessExpiry)
	refreshExp := now.Add(refreshExpiry)

	// 1. Generate Access Token (15m default)
	accessClaims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessTokenObj.SignedString([]byte(accessSecret))
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// 2. Generate Refresh Token (7d default)
	refreshClaims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshTokenObj.SignedString([]byte(refreshSecret))
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Calculate SHA-256 hash of refresh token for safe database storage
	refreshTokenHash := HashToken(refreshTokenStr)

	return &TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		ExpiresAt:    accessExp,
	}, refreshTokenHash, nil
}

// ValidateToken parses and validates a JWT token with strict signing method verification
func ValidateToken(tokenStr, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Strict check: verify signing method inside keyfunc
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("%w: %v", ErrInvalidAlg, t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// HashToken generates a SHA-256 hash for secure storage of refresh tokens
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
