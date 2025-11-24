package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// HTTPUserServiceClient implements UserServiceClient via HTTP calls to user-service
type HTTPUserServiceClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPUserServiceClient creates a new HTTP user service client
// All internal calls go through Gateway for unified monitoring and rate limiting
// If baseURL is empty, it will try to read from GATEWAY_URL environment variable
// If still empty, defaults to "http://gateway:8080"
func NewHTTPUserServiceClient(baseURL string) UserServiceClient {
	if baseURL == "" {
		baseURL = os.Getenv("GATEWAY_URL")
		if baseURL == "" {
			baseURL = "http://gateway:8080" // Default gateway URL in Docker
		}
	}

	return &HTTPUserServiceClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type userProfileResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID                int64  `json:"id"`
		ProfileVisibility string `json:"profile_visibility"`
		Email             string `json:"email"`
	} `json:"data"`
}

func (c *HTTPUserServiceClient) CheckProfileVisibility(ctx context.Context, userID int64, viewerEmail *string) (bool, error) {
	// Call Gateway's internal API route
	url := fmt.Sprintf("%s/api/v1/internal/user/users/%d", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Forward viewer email if available (for aalto_only check)
	if viewerEmail != nil {
		req.Header.Set("X-User-Email", *viewerEmail)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call user-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return false, nil // Profile is not accessible
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("user-service returned status %d", resp.StatusCode)
	}

	var userResp userProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	// If we got here, the profile is accessible
	return true, nil
}

