package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"github.com/jmoiron/sqlx"
)

// postgresRepository implements UserRepository using PostgreSQL
type postgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL user repository
func NewPostgresRepository(db *sqlx.DB) UserRepository {
	return &postgresRepository{db: db}
}

// FindByEmail finds a user by email
func (r *postgresRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	query := `SELECT id, email, password_hash, name, role, is_verified_email, 
		oauth_provider, last_login_at, failed_attempts, locked_until, 
		created_at, updated_at 
		FROM users WHERE email = $1`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return &user, nil
}

// FindByID finds a user by ID
func (r *postgresRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	query := `SELECT id, email, password_hash, name, role, is_verified_email, 
		oauth_provider, last_login_at, failed_attempts, locked_until, 
		created_at, updated_at 
		FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (r *postgresRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (email, password_hash, name, role, is_verified_email, 
		oauth_provider, failed_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Role,
		user.IsVerifiedEmail,
		user.OAuthProvider,
		user.FailedAttempts,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return errs.ErrEmailExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateLoginInfo updates login-related information
func (r *postgresRepository) UpdateLoginInfo(ctx context.Context, userID int64, lastLogin *time.Time, failedAttempts int, lockedUntil *time.Time) error {
	query := `UPDATE users 
		SET last_login_at = $1, failed_attempts = $2, locked_until = $3, updated_at = $4
		WHERE id = $5`

	_, err := r.db.ExecContext(ctx, query, lastLogin, failedAttempts, lockedUntil, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update login info: %w", err)
	}

	return nil
}

// SetEmailVerified sets email verification status
func (r *postgresRepository) SetEmailVerified(ctx context.Context, userID int64) error {
	query := `UPDATE users SET is_verified_email = true, updated_at = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to set email verified: %w", err)
	}

	return nil
}

// SetFailedAttempts sets failed login attempts count
func (r *postgresRepository) SetFailedAttempts(ctx context.Context, userID int64, attempts int) error {
	query := `UPDATE users SET failed_attempts = $1, updated_at = $2 WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, attempts, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to set failed attempts: %w", err)
	}

	return nil
}

// LockAccount locks an account until the specified time
func (r *postgresRepository) LockAccount(ctx context.Context, userID int64, until *time.Time) error {
	query := `UPDATE users SET locked_until = $1, updated_at = $2 WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, until, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to lock account: %w", err)
	}

	return nil
}

