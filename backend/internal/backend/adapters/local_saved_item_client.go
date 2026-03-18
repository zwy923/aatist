package adapters

import (
	"context"

	"github.com/aatist/backend/internal/user/model"
	userservice "github.com/aatist/backend/internal/user/service"
)

// LocalSavedItemClient implements opportunity's SavedItemClient via direct service call.
// Used when user and opportunity run in the same process (modular monolith).
type LocalSavedItemClient struct {
	savedItemSvc userservice.SavedItemService
}

// NewLocalSavedItemClient creates a new local saved item client
func NewLocalSavedItemClient(savedItemSvc userservice.SavedItemService) userservice.SavedItemClient {
	return &LocalSavedItemClient{savedItemSvc: savedItemSvc}
}

func toSavedItemType(s string) model.SavedItemType {
	return model.SavedItemType(s)
}

func (c *LocalSavedItemClient) SaveItem(ctx context.Context, userID int64, itemID int64, itemType string) error {
	return c.savedItemSvc.SaveItem(ctx, userID, itemID, toSavedItemType(itemType))
}

func (c *LocalSavedItemClient) UnsaveItem(ctx context.Context, userID int64, itemID int64, itemType string) error {
	return c.savedItemSvc.UnsaveItem(ctx, userID, itemID, toSavedItemType(itemType))
}

func (c *LocalSavedItemClient) IsSaved(ctx context.Context, userID int64, itemID int64, itemType string) (bool, error) {
	return c.savedItemSvc.IsSaved(ctx, userID, itemID, toSavedItemType(itemType))
}

func (c *LocalSavedItemClient) SaveUser(ctx context.Context, userID int64, savedUserID int64) error {
	return c.SaveItem(ctx, userID, savedUserID, string(model.SavedItemTypeUser))
}

func (c *LocalSavedItemClient) UnsaveUser(ctx context.Context, userID int64, savedUserID int64) error {
	return c.UnsaveItem(ctx, userID, savedUserID, string(model.SavedItemTypeUser))
}

func (c *LocalSavedItemClient) IsUserSaved(ctx context.Context, userID int64, savedUserID int64) (bool, error) {
	return c.IsSaved(ctx, userID, savedUserID, string(model.SavedItemTypeUser))
}

func (c *LocalSavedItemClient) SaveOpportunity(ctx context.Context, userID int64, opportunityID int64) error {
	return c.SaveItem(ctx, userID, opportunityID, string(model.SavedItemTypeOpportunity))
}

func (c *LocalSavedItemClient) UnsaveOpportunity(ctx context.Context, userID int64, opportunityID int64) error {
	return c.UnsaveItem(ctx, userID, opportunityID, string(model.SavedItemTypeOpportunity))
}

func (c *LocalSavedItemClient) IsOpportunitySaved(ctx context.Context, userID int64, opportunityID int64) (bool, error) {
	return c.IsSaved(ctx, userID, opportunityID, string(model.SavedItemTypeOpportunity))
}

func (c *LocalSavedItemClient) SaveEvent(ctx context.Context, userID int64, eventID int64) error {
	return c.SaveItem(ctx, userID, eventID, string(model.SavedItemTypeEvent))
}

func (c *LocalSavedItemClient) UnsaveEvent(ctx context.Context, userID int64, eventID int64) error {
	return c.UnsaveItem(ctx, userID, eventID, string(model.SavedItemTypeEvent))
}

func (c *LocalSavedItemClient) IsEventSaved(ctx context.Context, userID int64, eventID int64) (bool, error) {
	return c.IsSaved(ctx, userID, eventID, string(model.SavedItemTypeEvent))
}
