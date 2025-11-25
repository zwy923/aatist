package model

import (
	"time"
)

// SavedItemType represents the type of saved item.
type SavedItemType string

const (
	SavedItemTypeProject     SavedItemType = "project"
	SavedItemTypeOpportunity SavedItemType = "opportunity"
	SavedItemTypeUser        SavedItemType = "user"
	SavedItemTypeEvent       SavedItemType = "event"
)

// SavedItem represents a saved item in the saved_items table.
type SavedItem struct {
	ID        int64         `db:"id" json:"id"`
	UserID    int64         `db:"user_id" json:"user_id"`
	ItemID    int64         `db:"item_id" json:"item_id"`
	ItemType  SavedItemType `db:"item_type" json:"item_type"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
}
