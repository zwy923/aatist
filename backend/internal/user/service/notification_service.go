package service

import (
	"context"
	"fmt"

	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/internal/user/repository"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, userID int64, notifType model.NotificationType, title string, message *string, data model.NotificationData) error
	GetNotifications(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error)
	GetUnreadNotifications(ctx context.Context, userID int64, limit, offset int) ([]*model.Notification, error)
	MarkAsRead(ctx context.Context, userID int64, notificationID int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	DeleteNotification(ctx context.Context, userID int64, notificationID int64) error
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

// Helper function to notify when profile is saved
func NotifyProfileSaved(notifSvc NotificationService, ctx context.Context, savedUserID int64, saverUserID int64, saverName string) error {
	title := "Your profile was saved"
	message := fmt.Sprintf("%s saved your profile", saverName)
	data := model.NotificationData{
		"saver_user_id": saverUserID,
		"saver_name":    saverName,
	}
	return notifSvc.CreateNotification(ctx, savedUserID, model.NotificationTypeProfileSaved, title, &message, data)
}
