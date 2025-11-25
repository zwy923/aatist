package handler

import (
	"fmt"
	"net/http"

	"github.com/aalto-talent-network/backend/internal/notification/model"
	"github.com/aalto-talent-network/backend/internal/notification/service"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/platform/middleware"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"github.com/aalto-talent-network/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationSvc service.NotificationService
	logger          *log.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationSvc service.NotificationService, logger *log.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationSvc: notificationSvc,
		logger:          logger,
	}
}

// respondError is a helper to respond with error
func (h *NotificationHandler) respondError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, response.Error(err))
}

// handleServiceError handles service errors
func (h *NotificationHandler) handleServiceError(c *gin.Context, err error) {
	if appErr, ok := err.(*errs.AppError); ok {
		h.respondError(c, appErr.StatusCode, appErr)
		return
	}
	h.respondError(c, http.StatusInternalServerError, errs.ErrInternalError)
}

// GetNotificationsHandler returns notifications for the authenticated user
func (h *NotificationHandler) GetNotificationsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized)
		return
	}

	limit := 50
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		fmt.Sscanf(offsetStr, "%d", &offset)
	}

	unreadOnly := c.Query("unread") == "true"
	var notifications interface{}
	var err2 error

	if unreadOnly {
		notifications, err2 = h.notificationSvc.GetUnreadNotifications(c.Request.Context(), userID, limit, offset)
	} else {
		notifications, err2 = h.notificationSvc.GetNotifications(c.Request.Context(), userID, limit, offset)
	}

	if err2 != nil {
		h.handleServiceError(c, err2)
		return
	}

	c.JSON(http.StatusOK, response.Success(notifications))
}

// GetUnreadCountHandler returns unread notification count
func (h *NotificationHandler) GetUnreadCountHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized)
		return
	}

	count, err := h.notificationSvc.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"count": count}))
}

// MarkNotificationAsReadHandler marks a notification as read
func (h *NotificationHandler) MarkNotificationAsReadHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized)
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput)
		return
	}

	if err := h.notificationSvc.MarkAsRead(c.Request.Context(), userID, req.ID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "notification marked as read"}))
}

// MarkAllNotificationsAsReadHandler marks all notifications as read
func (h *NotificationHandler) MarkAllNotificationsAsReadHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized)
		return
	}

	if err := h.notificationSvc.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "all notifications marked as read"}))
}

// DeleteNotificationHandler deletes a notification
func (h *NotificationHandler) DeleteNotificationHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized)
		return
	}

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput)
		return
	}

	if err := h.notificationSvc.DeleteNotification(c.Request.Context(), userID, req.ID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "notification deleted successfully"}))
}

// CreateNotificationHandler creates a new notification (internal API for other services)
func (h *NotificationHandler) CreateNotificationHandler(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput)
		return
	}

	notifType := model.NotificationType(req.Type)
	if err := h.notificationSvc.CreateNotification(c.Request.Context(), req.UserID, notifType, req.Title, req.Message, model.NotificationData(req.Data)); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(gin.H{"message": "notification created successfully"}))
}

// DeleteMultipleNotificationsHandler deletes multiple notifications
func (h *NotificationHandler) DeleteMultipleNotificationsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized)
		return
	}

	var req struct {
		IDs []int64 `json:"ids"`
		All bool    `json:"all"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput)
		return
	}

	var deleted int64
	if req.All {
		// Delete all notifications
		deleted, err = h.notificationSvc.DeleteAllNotifications(c.Request.Context(), userID)
	} else if len(req.IDs) > 0 {
		// Delete specific notifications
		deleted, err = h.notificationSvc.DeleteMultipleNotifications(c.Request.Context(), userID, req.IDs)
	} else {
		h.respondError(c, http.StatusBadRequest, errs.NewAppError(errs.ErrInvalidInput, 400, "either 'ids' array or 'all: true' is required"))
		return
	}

	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message": "notifications deleted successfully",
		"deleted": deleted,
	}))
}