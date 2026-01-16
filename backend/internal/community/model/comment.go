package model

import "time"

// Comment represents a comment on a discussion post.
type Comment struct {
	ID        int64      `db:"id" json:"id"`
	PostID    int64      `db:"post_id" json:"post_id"`
	UserID    int64      `db:"user_id" json:"user_id"`
	ParentID  *int64     `db:"parent_id" json:"parent_id,omitempty"`
	Content   string     `db:"content" json:"content"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	Replies   []*Comment `db:"-" json:"replies,omitempty"`

	// Enriched fields
	AuthorName   string `db:"author_name" json:"author_name"`
	AuthorAvatar string `db:"author_avatar" json:"author_avatar"`
}
