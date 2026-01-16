package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aatist/backend/internal/notification/model"
	"github.com/jmoiron/sqlx"
)

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

func (r *postgresNotificationRepository) CreateBatch(ctx context.Context, notifications []*model.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	query := `INSERT INTO notifications (user_id, type, title, message, data, is_read, created_at)
		VALUES (:user_id, :type, :title, :message, :data, :is_read, :created_at)`

	now := time.Now()
	for _, n := range notifications {
		n.CreatedAt = now
	}

	_, err := r.db.NamedExecContext(ctx, query, notifications)
	if err != nil {
		return fmt.Errorf("failed to create batch notifications: %w", err)
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

func (r *postgresNotificationRepository) DeleteMultiple(ctx context.Context, notificationIDs []int64, userID int64) (int64, error) {
	if len(notificationIDs) == 0 {
		return 0, nil
	}

	query, args, err := sqlx.In(`DELETE FROM notifications WHERE id IN (?) AND user_id = ?`, notificationIDs, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to build delete query: %w", err)
	}

	// Rebind for PostgreSQL
	query = r.db.Rebind(query)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete notifications: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (r *postgresNotificationRepository) DeleteAll(ctx context.Context, userID int64) (int64, error) {
	query := `DELETE FROM notifications WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete all notifications: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

