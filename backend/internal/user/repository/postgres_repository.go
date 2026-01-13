package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aatist/backend/internal/user/model"
	"github.com/aatist/backend/pkg/errs"
	"github.com/jmoiron/sqlx"
)

const userSelectColumns = `id, email, password_hash, name, avatar_url, role,
	bio, profile_visibility, portfolio_visibility,
	is_verified_email, role_verified, oauth_provider, oauth_subject, last_login_at, failed_attempts, locked_until,
	created_at, updated_at,
	student_id, school, faculty, major, weekly_hours, weekly_availability, skills, courses,
	organization_name, organization_bio, contact_title, is_affiliated_with_school, org_size`

var profileUpdatableColumns = []string{
	// Common fields
	"name",
	"bio",
	"profile_visibility",
	// Student/Alumni fields
	"student_id",
	"school",
	"faculty",
	"major",
	"weekly_hours",
	"weekly_availability",
	"skills",
	"courses",
	"portfolio_visibility",
	// Organization fields
	"organization_name",
	"organization_bio",
	"contact_title",
	"is_affiliated_with_school",
	"org_size",
}

// postgresRepository implements UserRepository using PostgreSQL
type postgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL user repository
func NewPostgresRepository(db *sqlx.DB) UserRepository {
	return &postgresRepository{db: db}
}

// FindByEmail finds a user by email (case-insensitive)
// Note: email should be normalized to lowercase before calling this method
func (r *postgresRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	// Use LOWER() for case-insensitive comparison (matches the unique index)
	query := fmt.Sprintf("SELECT %s FROM users WHERE LOWER(email) = LOWER($1)", userSelectColumns)

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

// ExistsByEmail checks if an email is already registered (case-insensitive)
func (r *postgresRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER($1))`

	err := r.db.GetContext(ctx, &exists, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// UpdatePassword updates user's password hash
func (r *postgresRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errs.ErrUserNotFound
	}

	return nil
}

// CreateUser creates a new user
func (r *postgresRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (email, password_hash, name, role, bio, profile_visibility, portfolio_visibility,
		is_verified_email, role_verified, oauth_provider, oauth_subject, failed_attempts,
		student_id, school, faculty, major, weekly_hours, weekly_availability, skills, courses,
		organization_name, organization_bio, contact_title, is_affiliated_with_school, org_size,
		created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
		RETURNING id`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Role,
		user.Bio,
		user.ProfileVisibility.String(),
		user.PortfolioVisibility.String(),
		user.IsVerifiedEmail,
		user.RoleVerified,
		user.OAuthProvider,
		user.OAuthSubject,
		user.FailedAttempts,
		// Student/Alumni fields
		user.StudentID,
		user.School,
		user.Faculty,
		user.Major,
		user.WeeklyHours,
		user.WeeklyAvailability,
		user.Skills,
		user.Courses,
		// Organization fields
		user.OrganizationName,
		user.OrganizationBio,
		user.ContactTitle,
		user.IsAffiliatedWithSchool,
		user.OrgSize,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		// Check for unique constraint violation (case-insensitive email index)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
			strings.Contains(err.Error(), "idx_users_email_ci") {
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

func (r *postgresRepository) SearchSkills(ctx context.Context, query string, limit int) ([]model.SkillMetadata, error) {
	var skills []model.SkillMetadata
	sqlQuery := `SELECT id, name, category, created_at FROM skills WHERE name ILIKE $1 ORDER BY name ASC LIMIT $2`
	err := r.db.SelectContext(ctx, &skills, sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search skills: %w", err)
	}
	return skills, nil
}

func (r *postgresRepository) SearchCourses(ctx context.Context, query string, limit int) ([]model.CourseMetadata, error) {
	var courses []model.CourseMetadata
	sqlQuery := `SELECT id, code, name, school, created_at FROM courses 
		WHERE name ILIKE $1 OR code ILIKE $1 ORDER BY code ASC LIMIT $2`
	err := r.db.SelectContext(ctx, &courses, sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search courses: %w", err)
	}
	return courses, nil
}

func (r *postgresRepository) SearchTags(ctx context.Context, tagType string, query string, limit int) ([]model.TagMetadata, error) {
	var tags []model.TagMetadata
	sqlQuery := `SELECT id, name, type, created_at FROM tags 
		WHERE type = $1 AND name ILIKE $2 ORDER BY name ASC LIMIT $3`
	err := r.db.SelectContext(ctx, &tags, sqlQuery, tagType, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search tags: %w", err)
	}
	return tags, nil
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

// SavedItemRepository implementation

type postgresSavedItemRepository struct {
	db *sqlx.DB
}

func NewPostgresSavedItemRepository(db *sqlx.DB) SavedItemRepository {
	return &postgresSavedItemRepository{db: db}
}

func (r *postgresSavedItemRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.SavedItem, error) {
	var items []*model.SavedItem
	query := `SELECT id, user_id, item_id, item_type, created_at
		FROM saved_items WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &items, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find saved items by user id: %w", err)
	}

	return items, nil
}

func (r *postgresSavedItemRepository) FindByUserIDAndType(ctx context.Context, userID int64, itemType model.SavedItemType) ([]*model.SavedItem, error) {
	var items []*model.SavedItem
	query := `SELECT id, user_id, item_id, item_type, created_at
		FROM saved_items WHERE user_id = $1 AND item_type = $2 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &items, query, userID, itemType)
	if err != nil {
		return nil, fmt.Errorf("failed to find saved items by user id and type: %w", err)
	}

	return items, nil
}

func (r *postgresSavedItemRepository) Create(ctx context.Context, item *model.SavedItem) error {
	query := `INSERT INTO saved_items (user_id, item_id, item_type, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	item.CreatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		item.UserID,
		item.ItemID,
		item.ItemType,
		item.CreatedAt,
	).Scan(&item.ID)

	if err != nil {
		return fmt.Errorf("failed to create saved item: %w", err)
	}

	return nil
}

func (r *postgresSavedItemRepository) Delete(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) error {
	query := `DELETE FROM saved_items WHERE user_id = $1 AND item_id = $2 AND item_type = $3`

	result, err := r.db.ExecContext(ctx, query, userID, itemID, itemType)
	if err != nil {
		return fmt.Errorf("failed to delete saved item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("saved item not found or unauthorized")
	}

	return nil
}

func (r *postgresSavedItemRepository) Exists(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM saved_items WHERE user_id = $1 AND item_id = $2 AND item_type = $3)`

	err := r.db.GetContext(ctx, &exists, query, userID, itemID, itemType)
	if err != nil {
		return false, fmt.Errorf("failed to check saved item existence: %w", err)
	}

	return exists, nil
}
