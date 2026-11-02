package auth

import (
	"context"
	"time"
)

// UserRecord is the full user row as stored in the database.
// It contains the bcrypt PasswordHash, so it must NEVER be sent to the client.
// Always convert it with toUserResponse() before writing a response.
type UserRecord struct {
	ID             string
	Email          string
	PasswordHash   string
	Role           string
	FailedAttempts int
	LockedUntil    *time.Time
	CreatedAt      time.Time
	Student        *StudentProfileBrief // nil for non-students
}

// RefreshRecord is one stored refresh token (only its SHA-256 hash is stored).
type RefreshRecord struct {
	UserID    string
	Revoked   bool
	ExpiresAt time.Time
}

// Store is the "repository" the auth Service talks to.
// Two implementations exist:
//   - PostgresStore (store_postgres.go): the real database
//   - MemoryStore   (store_memory.go):   used when PostgreSQL is not running (demo mode)
//
// Because the Service only knows this interface, the login/signup logic is
// written once and works with either storage.
type Store interface {
	CreateStudent(ctx context.Context, email, passwordHash, regNo, fullName string) (*UserRecord, error)
	FindByLogin(ctx context.Context, emailOrRegNo string) (*UserRecord, error)
	FindByID(ctx context.Context, userID string) (*UserRecord, error)
	SetLoginFailures(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error
	UpdatePasswordAndRevokeSessions(ctx context.Context, userID, newHash string) error

	SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshRecord, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
}
