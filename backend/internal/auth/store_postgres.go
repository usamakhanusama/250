package auth

import (
	"context"
	"errors"
	"time"

	"sust-cse-portal/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresStore implements Store using PostgreSQL.
// Every query uses $1, $2 placeholders, so user input is never pasted into SQL
// (this is what prevents SQL injection).
type PostgresStore struct {
	db *db.DB
}

func NewPostgresStore(database *db.DB) *PostgresStore {
	return &PostgresStore{db: database}
}

func (s *PostgresStore) CreateStudent(ctx context.Context, email, passwordHash, regNo, fullName string) (*UserRecord, error) {
	var emailTaken, regTaken bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1),
		       EXISTS(SELECT 1 FROM students WHERE reg_no = $2)
	`, email, regNo).Scan(&emailTaken, &regTaken)
	if err != nil {
		return nil, err
	}
	if regTaken {
		return nil, ErrRegNoExists
	}
	if emailTaken {
		return nil, ErrEmailExists
	}

	// users + students rows must be created together, so we use a transaction:
	// if the second INSERT fails, the first one is rolled back.
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rec := &UserRecord{Email: email, Role: "student", Student: &StudentProfileBrief{RegNo: regNo, FullName: fullName}}
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role) VALUES ($1, $2, 'student')
		RETURNING id, created_at
	`, email, passwordHash).Scan(&rec.ID, &rec.CreatedAt)
	if err != nil {
		return nil, mapUniqueViolation(err)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO students (user_id, reg_no, full_name, batch, session, current_semester)
		VALUES ($1, $2, $3, '2023', '2023-2024', '4th Semester')
		RETURNING id, current_semester
	`, rec.ID, regNo, fullName).Scan(&rec.Student.StudentID, &rec.Student.CurrentSemester)
	if err != nil {
		return nil, mapUniqueViolation(err)
	}

	return rec, tx.Commit(ctx)
}

const selectUserWithStudent = `
	SELECT u.id, u.email, u.password_hash, u.role, u.failed_login_attempts, u.locked_until, u.created_at,
	       s.id, s.reg_no, s.full_name, s.current_semester
	FROM users u
	LEFT JOIN students s ON s.user_id = u.id
`

func (s *PostgresStore) FindByLogin(ctx context.Context, emailOrRegNo string) (*UserRecord, error) {
	return s.scanUser(s.db.Pool.QueryRow(ctx, selectUserWithStudent+` WHERE u.email = $1 OR s.reg_no = $1 LIMIT 1`, emailOrRegNo))
}

func (s *PostgresStore) FindByID(ctx context.Context, userID string) (*UserRecord, error) {
	return s.scanUser(s.db.Pool.QueryRow(ctx, selectUserWithStudent+` WHERE u.id = $1`, userID))
}

func (s *PostgresStore) scanUser(row pgx.Row) (*UserRecord, error) {
	var rec UserRecord
	var stID, reg, name, sem *string // nullable because of LEFT JOIN
	err := row.Scan(&rec.ID, &rec.Email, &rec.PasswordHash, &rec.Role, &rec.FailedAttempts, &rec.LockedUntil, &rec.CreatedAt,
		&stID, &reg, &name, &sem)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if stID != nil {
		rec.Student = &StudentProfileBrief{StudentID: *stID, RegNo: *reg, FullName: *name, CurrentSemester: *sem}
	}
	return &rec, nil
}

func (s *PostgresStore) SetLoginFailures(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error {
	_, err := s.db.Pool.Exec(ctx, `UPDATE users SET failed_login_attempts = $1, locked_until = $2 WHERE id = $3`,
		attempts, lockedUntil, userID)
	return err
}

func (s *PostgresStore) UpdatePasswordAndRevokeSessions(ctx context.Context, userID, newHash string) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, newHash, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.Pool.Exec(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	return err
}

func (s *PostgresStore) FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshRecord, error) {
	var rt RefreshRecord
	err := s.db.Pool.QueryRow(ctx, `SELECT user_id, revoked, expires_at FROM refresh_tokens WHERE token_hash = $1`, tokenHash).
		Scan(&rt.UserID, &rt.Revoked, &rt.ExpiresAt)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	return &rt, nil
}

func (s *PostgresStore) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := s.db.Pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`, tokenHash)
	return err
}

// mapUniqueViolation turns a PostgreSQL "duplicate key" error (code 23505) into a friendly error.
// This covers the rare race where two people sign up with the same reg no at the same moment.
func mapUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrRegNoExists
	}
	return err
}
