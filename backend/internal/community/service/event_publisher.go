package service

import (
	"context"
	"time"
)

const (
	EventPostCreated   = "community.post.created"
	EventPostLiked     = "community.post.liked"
	EventPostCommented = "community.post.commented"
)

// EventPublisher abstracts message queue publishing so services can remain decoupled.
type EventPublisher interface {
	PublishCommunityEvent(ctx context.Context, eventType string, payload interface{}) error
}

// PostCreatedEvent represents payload for community.post.created.
type PostCreatedEvent struct {
	PostID    int64     `json:"post_id"`
	AuthorID  int64     `json:"author_id"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	Tags      []string  `json:"tags"`
}

// PostLikedEvent payload for community.post.liked.
type PostLikedEvent struct {
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
	LikerID  int64 `json:"liker_id"`
}

// PostCommentedEvent payload for community.post.commented.
type PostCommentedEvent struct {
	PostID            int64   `json:"post_id"`
	AuthorID          int64   `json:"author_id"`
	CommentID         int64   `json:"comment_id"`
	CommentAuthorID   int64   `json:"comment_author_id"`
	MentionedUserIDs  []int64 `json:"mentioned_user_ids,omitempty"`
	ParentCommentID   *int64  `json:"parent_comment_id,omitempty"`
	CommentSnippet    string  `json:"comment_snippet"`
	AdditionalContext string  `json:"additional_context,omitempty"`
}
