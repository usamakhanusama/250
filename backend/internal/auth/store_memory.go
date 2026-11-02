package auth

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryStore implements Store with Go maps, protected by a mutex.
// It is used only when PostgreSQL is not running (local demo mode).
// Data is lost when the server restarts.
type MemoryStore struct {
	mu            sync.Mutex
	users         map[string]*UserRecord // key: user ID
	refreshTokens map[string]*RefreshRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:         make(map[string]*UserRecord),
		refreshTokens: make(map[string]*RefreshRecord),
	}
}

func (m *MemoryStore) CreateStudent(_ context.Context, email, passwordHash, regNo, fullName string) (*UserRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.users {
		if u.Student != nil && u.Student.RegNo == regNo {
			return nil, ErrRegNoExists
		}
		if u.Email == email {
			return nil, ErrEmailExists
		}
	}

	rec := &UserRecord{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         "student",
		CreatedAt:    time.Now(),
		Student: &StudentProfileBrief{
			StudentID:       uuid.NewString(),
			RegNo:           regNo,
			FullName:        fullName,
			CurrentSemester: "4th Semester",
		},
	}
	m.users[rec.ID] = rec
	copy := *rec
	return &copy, nil
}

func (m *MemoryStore) FindByLogin(_ context.Context, emailOrRegNo string) (*UserRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Email == emailOrRegNo || (u.Student != nil && u.Student.RegNo == emailOrRegNo) {
			copy := *u
			return &copy, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *MemoryStore) FindByID(_ context.Context, userID string) (*UserRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	copy := *u
	return &copy, nil
}

func (m *MemoryStore) SetLoginFailures(_ context.Context, userID string, attempts int, lockedUntil *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[userID]; ok {
		u.FailedAttempts = attempts
		u.LockedUntil = lockedUntil
	}
	return nil
}

func (m *MemoryStore) UpdatePasswordAndRevokeSessions(_ context.Context, userID, newHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return ErrUserNotFound
	}
	u.PasswordHash = newHash
	for _, rt := range m.refreshTokens {
		if rt.UserID == userID {
			rt.Revoked = true
		}
	}
	return nil
}

func (m *MemoryStore) SaveRefreshToken(_ context.Context, userID, tokenHash string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshTokens[tokenHash] = &RefreshRecord{UserID: userID, ExpiresAt: expiresAt}
	return nil
}

func (m *MemoryStore) FindRefreshToken(_ context.Context, tokenHash string) (*RefreshRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rt, ok := m.refreshTokens[tokenHash]
	if !ok {
		return nil, ErrInvalidRefreshToken
	}
	copy := *rt
	return &copy, nil
}

func (m *MemoryStore) RevokeRefreshToken(_ context.Context, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rt, ok := m.refreshTokens[tokenHash]; ok {
		rt.Revoked = true
	}
	return nil
}
