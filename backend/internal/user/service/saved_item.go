package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/internal/user/repository"
)

// ============================================================================
// SavedItemService - Internal service for user-service (direct database access)
// ============================================================================

// SavedItemService defines the interface for saved items operations (internal)
type SavedItemService interface {
	GetSavedItems(ctx context.Context, userID int64) ([]*model.SavedItem, error)
	GetSavedItemsByType(ctx context.Context, userID int64, itemType model.SavedItemType) ([]*model.SavedItem, error)
	SaveItem(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) error
	UnsaveItem(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) error
	IsSaved(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) (bool, error)
}

type savedItemService struct {
	savedItemRepo repository.SavedItemRepository
}

// NewSavedItemService creates a new saved item service (for user-service internal use)
func NewSavedItemService(savedItemRepo repository.SavedItemRepository) SavedItemService {
	return &savedItemService{
		savedItemRepo: savedItemRepo,
	}
}

func (s *savedItemService) GetSavedItems(ctx context.Context, userID int64) ([]*model.SavedItem, error) {
	return s.savedItemRepo.FindByUserID(ctx, userID)
}

func (s *savedItemService) GetSavedItemsByType(ctx context.Context, userID int64, itemType model.SavedItemType) ([]*model.SavedItem, error) {
	return s.savedItemRepo.FindByUserIDAndType(ctx, userID, itemType)
}

func (s *savedItemService) SaveItem(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) error {
	// Check if already saved
	exists, err := s.savedItemRepo.Exists(ctx, userID, itemID, itemType)
	if err != nil {
		return fmt.Errorf("failed to check saved item: %w", err)
	}
	if exists {
		return nil // Already saved, no error
	}

	item := &model.SavedItem{
		UserID:   userID,
		ItemID:   itemID,
		ItemType: itemType,
	}
	return s.savedItemRepo.Create(ctx, item)
}

func (s *savedItemService) UnsaveItem(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) error {
	return s.savedItemRepo.Delete(ctx, userID, itemID, itemType)
}

func (s *savedItemService) IsSaved(ctx context.Context, userID int64, itemID int64, itemType model.SavedItemType) (bool, error) {
	return s.savedItemRepo.Exists(ctx, userID, itemID, itemType)
}

// ============================================================================
// SavedItemClient - HTTP client for other services to call user-service
// ============================================================================

// SavedItemClient defines interface for saved items operations via HTTP
// This client calls user-service's saved items API through Gateway
// It can be used by other services (opp-service, portfolio-service, etc.)
type SavedItemClient interface {
	// SaveItem saves an item (project, opportunity, user, or future event)
	SaveItem(ctx context.Context, userID int64, itemID int64, itemType string) error

	// UnsaveItem unsaves an item
	UnsaveItem(ctx context.Context, userID int64, itemID int64, itemType string) error

	// IsSaved checks if an item is saved by a user
	IsSaved(ctx context.Context, userID int64, itemID int64, itemType string) (bool, error)

	// SaveUser saves a user profile (convenience method)
	SaveUser(ctx context.Context, userID int64, savedUserID int64) error

	// UnsaveUser unsaves a user profile (convenience method)
	UnsaveUser(ctx context.Context, userID int64, savedUserID int64) error

	// IsUserSaved checks if a user is saved by another user (convenience method)
	IsUserSaved(ctx context.Context, userID int64, savedUserID int64) (bool, error)

	// SaveOpportunity saves an opportunity (convenience method)
	SaveOpportunity(ctx context.Context, userID int64, opportunityID int64) error

	// UnsaveOpportunity unsaves an opportunity (convenience method)
	UnsaveOpportunity(ctx context.Context, userID int64, opportunityID int64) error

	// IsOpportunitySaved checks if an opportunity is saved by a user (convenience method)
	IsOpportunitySaved(ctx context.Context, userID int64, opportunityID int64) (bool, error)

	// SaveEvent saves an event (convenience method)
	SaveEvent(ctx context.Context, userID int64, eventID int64) error

	// UnsaveEvent unsaves an event (convenience method)
	UnsaveEvent(ctx context.Context, userID int64, eventID int64) error

	// IsEventSaved checks if an event is saved by a user (convenience method)
	IsEventSaved(ctx context.Context, userID int64, eventID int64) (bool, error)
}

type httpSavedItemClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPSavedItemClient creates a new HTTP saved item client
// All internal calls go through Gateway for unified monitoring and rate limiting
// This client can be used by any service that needs to interact with saved items
func NewHTTPSavedItemClient() SavedItemClient {
	baseURL := os.Getenv("GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://gateway:8080" // Default gateway URL in Docker
	}

	return &httpSavedItemClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type saveItemRequest struct {
	ItemID   int64  `json:"item_id"`
	ItemType string `json:"item_type"`
}

type savedItemResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID        int64  `json:"id"`
		UserID    int64  `json:"user_id"`
		ItemID    int64  `json:"item_id"`
		ItemType  string `json:"item_type"`
		CreatedAt string `json:"created_at"`
	} `json:"data"`
}

type savedItemsListResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID        int64  `json:"id"`
		UserID    int64  `json:"user_id"`
		ItemID    int64  `json:"item_id"`
		ItemType  string `json:"item_type"`
		CreatedAt string `json:"created_at"`
	} `json:"data"`
}

func (c *httpSavedItemClient) SaveItem(ctx context.Context, userID int64, itemID int64, itemType string) error {
	reqBody := saveItemRequest{
		ItemID:   itemID,
		ItemType: itemType,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal save item request: %w", err)
	}

	// Call Gateway's user-service API route
	url := fmt.Sprintf("%s/api/v1/users/me/saved", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Set user ID header for Gateway to forward to user-service
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send save item request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("user-service returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *httpSavedItemClient) UnsaveItem(ctx context.Context, userID int64, itemID int64, itemType string) error {
	reqBody := saveItemRequest{
		ItemID:   itemID,
		ItemType: itemType,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal unsave item request: %w", err)
	}

	// Call Gateway's user-service API route
	url := fmt.Sprintf("%s/api/v1/users/me/saved", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Set user ID header for Gateway to forward to user-service
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send unsave item request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("user-service returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *httpSavedItemClient) IsSaved(ctx context.Context, userID int64, itemID int64, itemType string) (bool, error) {
	// Call Gateway's user-service API route to get saved items
	url := fmt.Sprintf("%s/api/v1/users/me/saved?type=%s", c.baseURL, itemType)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user ID header for Gateway to forward to user-service
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send get saved items request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("user-service returned status %d", resp.StatusCode)
	}

	var listResp savedItemsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if the specific item is in the list
	for _, item := range listResp.Data {
		if item.ItemID == itemID && item.ItemType == itemType {
			return true, nil
		}
	}

	return false, nil
}

// SaveUser saves a user profile (convenience method)
func (c *httpSavedItemClient) SaveUser(ctx context.Context, userID int64, savedUserID int64) error {
	return c.SaveItem(ctx, userID, savedUserID, string(model.SavedItemTypeUser))
}

// UnsaveUser unsaves a user profile (convenience method)
func (c *httpSavedItemClient) UnsaveUser(ctx context.Context, userID int64, savedUserID int64) error {
	return c.UnsaveItem(ctx, userID, savedUserID, string(model.SavedItemTypeUser))
}

// IsUserSaved checks if a user is saved by another user (convenience method)
func (c *httpSavedItemClient) IsUserSaved(ctx context.Context, userID int64, savedUserID int64) (bool, error) {
	return c.IsSaved(ctx, userID, savedUserID, string(model.SavedItemTypeUser))
}

// SaveOpportunity saves an opportunity (convenience method)
func (c *httpSavedItemClient) SaveOpportunity(ctx context.Context, userID int64, opportunityID int64) error {
	return c.SaveItem(ctx, userID, opportunityID, string(model.SavedItemTypeOpportunity))
}

// UnsaveOpportunity unsaves an opportunity (convenience method)
func (c *httpSavedItemClient) UnsaveOpportunity(ctx context.Context, userID int64, opportunityID int64) error {
	return c.UnsaveItem(ctx, userID, opportunityID, string(model.SavedItemTypeOpportunity))
}

// IsOpportunitySaved checks if an opportunity is saved by a user (convenience method)
func (c *httpSavedItemClient) IsOpportunitySaved(ctx context.Context, userID int64, opportunityID int64) (bool, error) {
	return c.IsSaved(ctx, userID, opportunityID, string(model.SavedItemTypeOpportunity))
}

// SaveEvent saves an event (convenience method)
func (c *httpSavedItemClient) SaveEvent(ctx context.Context, userID int64, eventID int64) error {
	return c.SaveItem(ctx, userID, eventID, string(model.SavedItemTypeEvent))
}

// UnsaveEvent unsaves an event (convenience method)
func (c *httpSavedItemClient) UnsaveEvent(ctx context.Context, userID int64, eventID int64) error {
	return c.UnsaveItem(ctx, userID, eventID, string(model.SavedItemTypeEvent))
}

// IsEventSaved checks if an event is saved by a user (convenience method)
func (c *httpSavedItemClient) IsEventSaved(ctx context.Context, userID int64, eventID int64) (bool, error) {
	return c.IsSaved(ctx, userID, eventID, string(model.SavedItemTypeEvent))
}
