package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// NotificationClient defines interface for creating notifications via HTTP
type NotificationClient interface {
	CreateNotification(ctx context.Context, userID int64, notifType string, title string, message *string, data map[string]interface{}) error
}

type httpNotificationClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPNotificationClient creates a new HTTP notification client
// All internal calls go through Gateway for unified monitoring and rate limiting
func NewHTTPNotificationClient() NotificationClient {
	// Use Gateway URL instead of direct service URL
	baseURL := os.Getenv("GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://gateway:8080" // Default gateway URL in Docker
	}

	return &httpNotificationClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

type createNotificationRequest struct {
	UserID  int64                  `json:"user_id"`
	Type    string                 `json:"type"`
	Title   string                 `json:"title"`
	Message *string                `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

func (c *httpNotificationClient) CreateNotification(ctx context.Context, userID int64, notifType string, title string, message *string, data map[string]interface{}) error {
	reqBody := createNotificationRequest{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
		Data:    data,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %w", err)
	}

	// Call Gateway's internal API route
	url := fmt.Sprintf("%s/api/v1/internal/notification/notifications", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Gateway will automatically set internal call headers when proxying
	// We don't need to set them here since Gateway handles it

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}

	return nil
}

// NotifyProfileSaved sends a notification when a profile is saved
func NotifyProfileSaved(client NotificationClient, ctx context.Context, savedUserID int64, saverUserID int64, saverName string) error {
	title := "Your profile was saved"
	message := fmt.Sprintf("%s saved your profile", saverName)
	data := map[string]interface{}{
		"saver_user_id": saverUserID,
		"saver_name":    saverName,
	}
	return client.CreateNotification(ctx, savedUserID, "profile_saved", title, &message, data)
}
