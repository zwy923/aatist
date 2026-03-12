package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// User-related notifications
	NotificationTypeProfileSaved NotificationType = "profile_saved"

	// Social notifications
	NotificationTypeFollow       NotificationType = "follow"
	NotificationTypeProjectSaved NotificationType = "project_saved"
	NotificationTypeComment      NotificationType = "comment"

	// Opportunity notifications
	NotificationTypeOpportunityMatch NotificationType = "opportunity_match"

	// Project notifications
	NotificationTypeProjectInvite NotificationType = "project_invite"

	// AI notifications
	NotificationTypeAISummaryFinished NotificationType = "ai_summary_finished"

	// System notifications
	NotificationTypeSystem       NotificationType = "system"
	NotificationTypeWeeklyDigest NotificationType = "weekly_digest"

	// Message notifications
	NotificationTypeMessage NotificationType = "message"
)

// NotificationData is a JSONB field for additional notification data
type NotificationData map[string]interface{}

func (nd NotificationData) Value() (driver.Value, error) {
	if len(nd) == 0 {
		return []byte("{}"), nil
	}
	b, err := json.Marshal(nd)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (nd *NotificationData) Scan(value interface{}) error {
	if value == nil {
		*nd = make(NotificationData)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for NotificationData: %T", value)
	}

	if len(bytes) == 0 {
		*nd = make(NotificationData)
		return nil
	}

	var temp map[string]interface{}
	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}
	*nd = temp
	return nil
}

// Notification represents a notification in the system
// This is a generic domain model that can be used by any service
type Notification struct {
	ID        int64            `db:"id" json:"id"`
	UserID    int64            `db:"user_id" json:"user_id"`
	Type      NotificationType `db:"type" json:"type"`
	Title     string           `db:"title" json:"title"`
	Message   *string          `db:"message" json:"message,omitempty"`
	Data      NotificationData `db:"data" json:"data,omitempty"`
	IsRead    bool             `db:"is_read" json:"is_read"`
	CreatedAt time.Time        `db:"created_at" json:"created_at"`
}
