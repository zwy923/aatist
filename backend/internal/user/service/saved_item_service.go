package service

import (
	"context"
	"fmt"

	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/internal/user/repository"
)

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
