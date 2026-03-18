package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// FileServiceClient defines interface for uploading files via HTTP
type FileServiceClient interface {
	UploadFile(ctx context.Context, userID int64, role, email, fileType string, reader io.Reader, size int64, contentType, filename string) (*FileUploadResponse, error)
}

type httpFileServiceClient struct {
	baseURL string
	client  *http.Client
}

// FileUploadResponse represents the response from file-service
type FileUploadResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID          int64  `json:"id"`
		UserID      int64  `json:"user_id"`
		Type        string `json:"type"`
		URL         string `json:"url"`
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
		Size        int64  `json:"size"`
	} `json:"data"`
}

// NewHTTPFileServiceClient creates a new HTTP file service client
// All internal calls go through Gateway for unified monitoring and rate limiting
func NewHTTPFileServiceClient() FileServiceClient {
	// Use Gateway URL instead of direct service URL
	baseURL := os.Getenv("GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://gateway:8080" // Default gateway URL in Docker
	}

	return &httpFileServiceClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for file uploads
		},
	}
}

func (c *httpFileServiceClient) UploadFile(ctx context.Context, userID int64, role, email, fileType string, reader io.Reader, size int64, contentType, filename string) (*FileUploadResponse, error) {
	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, reader); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// Add type parameter
	if err := writer.WriteField("type", fileType); err != nil {
		return nil, fmt.Errorf("failed to write type field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Call Gateway's internal API route
	url := fmt.Sprintf("%s/api/v1/internal/file/upload", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
	req.Header.Set("X-User-Role", role)
	req.Header.Set("X-User-Email", email)
	if token := os.Getenv("INTERNAL_API_TOKEN"); token != "" {
		req.Header.Set("X-Internal-Call", "true")
		req.Header.Set("X-Internal-Token", token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send file upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("file service returned status %d", resp.StatusCode)
	}

	var fileResp FileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &fileResp, nil
}
