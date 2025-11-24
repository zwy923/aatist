package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"github.com/jmoiron/sqlx"
)

const userSelectColumns = `id, email, password_hash, name, nickname, avatar_url, role,
	student_id, school, faculty, major, availability, projects, skills, bio,
	is_verified_email, oauth_provider, last_login_at, failed_attempts, locked_until,
	created_at, updated_at`

var profileUpdatableColumns = []string{
	"name",
	"nickname",
	"student_id",
	"school",
	"faculty",
	"major",
	"availability",
	"projects",
	"skills",
	"bio",
}

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
	query := fmt.Sprintf("SELECT %s FROM users WHERE email = $1", userSelectColumns)

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
	query := fmt.Sprintf("SELECT %s FROM users WHERE id = $1", userSelectColumns)

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
	query := `INSERT INTO users (email, password_hash, name, role, student_id, school, faculty, is_verified_email, 
		oauth_provider, failed_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Role,
		user.StudentID,
		user.School,
		user.Faculty,
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
func (r *postgresRepository) UpdateProfile(ctx context.Context, update ProfileUpdate) (*model.User, error) {
	if len(update.Fields) == 0 {
		return r.FindByID(ctx, update.UserID)
	}

	setClauses := make([]string, 0, len(update.Fields)+1)
	args := make([]interface{}, 0, len(update.Fields)+2)
	argIdx := 1

	for _, column := range profileUpdatableColumns {
		value, ok := update.Fields[column]
		if !ok {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, argIdx))
		args = append(args, value)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.FindByID(ctx, update.UserID)
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++
	args = append(args, update.UserID)

	query := fmt.Sprintf(`UPDATE users SET %s WHERE id = $%d RETURNING %s`,
		strings.Join(setClauses, ", "),
		argIdx,
		userSelectColumns,
	)

	var updated model.User
	if err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&updated); err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return &updated, nil
}

func (r *postgresRepository) UpdateAvatarURL(ctx context.Context, userID int64, avatarURL string) (*model.User, error) {
	query := fmt.Sprintf(`UPDATE users SET avatar_url = $1, updated_at = $2 WHERE id = $3 RETURNING %s`, userSelectColumns)

	var updated model.User
	if err := r.db.QueryRowxContext(ctx, query, avatarURL, time.Now(), userID).StructScan(&updated); err != nil {
		return nil, fmt.Errorf("failed to update avatar url: %w", err)
	}

	return &updated, nil
}
