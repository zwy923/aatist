package adapters

import (
	"context"

	"github.com/aatist/backend/internal/notification/model"
	"github.com/aatist/backend/internal/notification/service"
)

// LocalNotificationClient implements user's NotificationClient via direct service call.
// Used when user and notification run in the same process (modular monolith).
type LocalNotificationClient struct {
	notifSvc service.NotificationService
}

// NewLocalNotificationClient creates a new local notification client
func NewLocalNotificationClient(notifSvc service.NotificationService) *LocalNotificationClient {
	return &LocalNotificationClient{notifSvc: notifSvc}
}

// CreateNotification creates a notification for a user
func (c *LocalNotificationClient) CreateNotification(ctx context.Context, userID int64, notifType string, title string, message *string, data map[string]interface{}) error {
	var nd model.NotificationData
	if data != nil {
		nd = model.NotificationData(data)
	}
	return c.notifSvc.CreateNotification(ctx, userID, model.NotificationType(notifType), title, message, nd)
}
