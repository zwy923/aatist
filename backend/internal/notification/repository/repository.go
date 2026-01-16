package repository

import (
	"context"

	"github.com/aatist/backend/internal/notification/model"
)

// NotificationRepository defines the interface for notification operations
type NotificationRepository interface {
	// Create creates a new notification
	Create(ctx context.Context, notification *model.Notification) error

	// CreateBatch creates multiple notifications in a single transaction
	CreateBatch(ctx context.Context, notifications []*model.Notification) error

	// FindByUserID finds all notifications for a user
	FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error)

	// FindUnreadByUserID finds unread notifications for a user
	FindUnreadByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error)

	// MarkAsRead marks a notification as read
	MarkAsRead(ctx context.Context, notificationID int64, userID int64) error

	// MarkAllAsRead marks all notifications as read for a user
	MarkAllAsRead(ctx context.Context, userID int64) error

	// CountUnread counts unread notifications for a user
	CountUnread(ctx context.Context, userID int64) (int64, error)

	// Delete deletes a notification
	Delete(ctx context.Context, notificationID int64, userID int64) error

	// DeleteMultiple deletes multiple notifications by IDs
	DeleteMultiple(ctx context.Context, notificationIDs []int64, userID int64) (int64, error)

	// DeleteAll deletes all notifications for a user
	DeleteAll(ctx context.Context, userID int64) (int64, error)
}

