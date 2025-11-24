package handler

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID  int64                  `json:"user_id" binding:"required"`
	Type    string                 `json:"type" binding:"required"`
	Title   string                 `json:"title" binding:"required"`
	Message *string                `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

