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
	student_id, school, faculty, major, weekly_hours, emotional_status, weekly_availability,
	skills, bio, profile_visibility,
	is_verified_email, oauth_provider, oauth_subject, last_login_at, failed_attempts, locked_until,
	created_at, updated_at`

var profileUpdatableColumns = []string{
	"name",
	"nickname",
	"student_id",
	"school",
	"faculty",
	"major",
	"weekly_hours",
	"emotional_status",
	"weekly_availability",
	"skills",
	"bio",
	"profile_visibility",
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
	query := `INSERT INTO users (email, password_hash, name, role, student_id, school, faculty, profile_visibility, 
		is_verified_email, oauth_provider, oauth_subject, failed_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
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
		user.ProfileVisibility.String(), // Convert to string for DB
		user.IsVerifiedEmail,
		user.OAuthProvider,
		user.OAuthSubject,
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

// ProjectRepository implementation

type postgresProjectRepository struct {
	db *sqlx.DB
}

func NewPostgresProjectRepository(db *sqlx.DB) ProjectRepository {
	return &postgresProjectRepository{db: db}
}

func (r *postgresProjectRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.PortfolioProject, error) {
	var projects []*model.PortfolioProject
	query := `SELECT id, user_id, title, description, year, tags, cover_image_url, project_link, created_at, updated_at
		FROM projects WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &projects, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by user id: %w", err)
	}

	return projects, nil
}

func (r *postgresProjectRepository) FindByID(ctx context.Context, id int64) (*model.PortfolioProject, error) {
	var project model.PortfolioProject
	query := `SELECT id, user_id, title, description, year, tags, cover_image_url, project_link, created_at, updated_at
		FROM projects WHERE id = $1`

	err := r.db.GetContext(ctx, &project, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUserNotFound // Reuse error for consistency
		}
		return nil, fmt.Errorf("failed to find project by id: %w", err)
	}

	return &project, nil
}

func (r *postgresProjectRepository) Create(ctx context.Context, project *model.PortfolioProject) error {
	query := `INSERT INTO projects (user_id, title, description, year, tags, cover_image_url, project_link, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		project.UserID,
		project.Title,
		project.Description,
		project.Year,
		project.Tags,
		project.CoverImageURL,
		project.ProjectLink,
		project.CreatedAt,
		project.UpdatedAt,
	).Scan(&project.ID)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

func (r *postgresProjectRepository) Update(ctx context.Context, project *model.PortfolioProject) error {
	query := `UPDATE projects 
		SET title = $1, description = $2, year = $3, tags = $4, cover_image_url = $5, project_link = $6, updated_at = $7
		WHERE id = $8 AND user_id = $9
		RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		project.Title,
		project.Description,
		project.Year,
		project.Tags,
		project.CoverImageURL,
		project.ProjectLink,
		time.Now(),
		project.ID,
		project.UserID,
	).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrUserNotFound
		}
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

func (r *postgresProjectRepository) Delete(ctx context.Context, id int64, userID int64) error {
	query := `DELETE FROM projects WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or unauthorized")
	}

	return nil
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
		return fmt.Errorf("notification not found or unauthorized")
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

// NotificationRepository implementation

type postgresNotificationRepository struct {
	db *sqlx.DB
}

func NewPostgresNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &postgresNotificationRepository{db: db}
}

func (r *postgresNotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	query := `INSERT INTO notifications (user_id, type, title, message, data, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	notification.CreatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Data,
		notification.IsRead,
		notification.CreatedAt,
	).Scan(&notification.ID)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (r *postgresNotificationRepository) FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error) {
	var notifications []*model.Notification
	query := `SELECT id, user_id, type, title, message, data, is_read, created_at
		FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	err := r.db.SelectContext(ctx, &notifications, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find notifications by user id: %w", err)
	}

	return notifications, nil
}

func (r *postgresNotificationRepository) FindUnreadByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error) {
	var notifications []*model.Notification
	query := `SELECT id, user_id, type, title, message, data, is_read, created_at
		FROM notifications WHERE user_id = $1 AND is_read = false ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	err := r.db.SelectContext(ctx, &notifications, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find unread notifications by user id: %w", err)
	}

	return notifications, nil
}

func (r *postgresNotificationRepository) MarkAsRead(ctx context.Context, notificationID int64, userID int64) error {
	query := `UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or unauthorized")
	}

	return nil
}

func (r *postgresNotificationRepository) MarkAllAsRead(ctx context.Context, userID int64) error {
	query := `UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

func (r *postgresNotificationRepository) CountUnread(ctx context.Context, userID int64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`

	err := r.db.GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

func (r *postgresNotificationRepository) Delete(ctx context.Context, notificationID int64, userID int64) error {
	query := `DELETE FROM notifications WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or unauthorized")
	}

	return nil
}
