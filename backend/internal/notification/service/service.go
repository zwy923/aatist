package service

import (
	"context"

	"github.com/aatist/backend/internal/notification/model"
	"github.com/aatist/backend/internal/notification/repository"
)

// NotificationService defines the interface for notification operations
type NotificationService interface {
	CreateNotification(ctx context.Context, userID int64, notifType model.NotificationType, title string, message *string, data model.NotificationData) error
	CreateNotifications(ctx context.Context, notifications []*model.Notification) error
	GetNotifications(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error)
	GetUnreadNotifications(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error)
	MarkAsRead(ctx context.Context, userID int64, notificationID int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	DeleteNotification(ctx context.Context, userID int64, notificationID int64) error
	DeleteMultipleNotifications(ctx context.Context, userID int64, notificationIDs []int64) (int64, error)
	DeleteAllNotifications(ctx context.Context, userID int64) (int64, error)
}

type notificationService struct {
	notifRepo repository.NotificationRepository
}

func NewNotificationService(notifRepo repository.NotificationRepository) NotificationService {
	return &notificationService{
		notifRepo: notifRepo,
	}
}

func (s *notificationService) CreateNotification(ctx context.Context, userID int64, notifType model.NotificationType, title string, message *string, data model.NotificationData) error {
	if data == nil {
		data = make(model.NotificationData)
	}

	notification := &model.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
		Data:    data,
		IsRead:  false,
	}

	return s.notifRepo.Create(ctx, notification)
}

func (s *notificationService) CreateNotifications(ctx context.Context, notifications []*model.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	// Ideally validate notifications here
	return s.notifRepo.CreateBatch(ctx, notifications)
}

func (s *notificationService) GetNotifications(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.notifRepo.FindByUserID(ctx, userID, limit, offset)
}

func (s *notificationService) GetUnreadNotifications(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.notifRepo.FindUnreadByUserID(ctx, userID, limit, offset)
}

func (s *notificationService) MarkAsRead(ctx context.Context, userID int64, notificationID int64) error {
	return s.notifRepo.MarkAsRead(ctx, notificationID, userID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID int64) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	return s.notifRepo.CountUnread(ctx, userID)
}

func (s *notificationService) DeleteNotification(ctx context.Context, userID int64, notificationID int64) error {
	return s.notifRepo.Delete(ctx, notificationID, userID)
}

func (s *notificationService) DeleteMultipleNotifications(ctx context.Context, userID int64, notificationIDs []int64) (int64, error) {
	if len(notificationIDs) == 0 {
		return 0, nil
	}
	if len(notificationIDs) > 100 {
		notificationIDs = notificationIDs[:100] // Limit to 100 per request
	}
	return s.notifRepo.DeleteMultiple(ctx, notificationIDs, userID)
}

func (s *notificationService) DeleteAllNotifications(ctx context.Context, userID int64) (int64, error) {
	return s.notifRepo.DeleteAll(ctx, userID)
}

