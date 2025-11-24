package model

import "time"

// FileType represents the type of file
type FileType string

const (
	FileTypeAvatar      FileType = "avatar"
	FileTypeProjectCover FileType = "project_cover"
	FileTypePostImage   FileType = "post_image"
	FileTypeResume      FileType = "resume"
	FileTypeAIOutput    FileType = "ai_output"
	FileTypeOther       FileType = "other"
)

// File represents a file record in the database
type File struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Type        FileType  `json:"type" db:"type"`
	ObjectKey   string    `json:"object_key" db:"object_key"`
	URL         string    `json:"url" db:"url"`
	Filename    string    `json:"filename" db:"filename"`
	ContentType string    `json:"content_type" db:"content_type"`
	Size        int64     `json:"size" db:"size"`
	Metadata    string    `json:"metadata" db:"metadata"` // JSON string for additional metadata
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

