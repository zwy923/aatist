package repository

import (
	"context"

	"github.com/aatist/backend/internal/community/model"
)

// PostListSort controls ordering for list queries.
type PostListSort string

const (
	PostListSortNewest PostListSort = "newest"
	PostListSortOldest PostListSort = "oldest"
)

// PostListFilter represents filters for listing posts.
type PostListFilter struct {
	Category *model.PostCategory
	Limit    int
	Offset   int
	Sort     PostListSort
}

// PostSearchFilter represents filters for searching posts.
type PostSearchFilter struct {
	Query  string
	Limit  int
	Offset int
}

// PostRepository defines operations on discussion posts.
type PostRepository interface {
	Create(ctx context.Context, post *model.DiscussionPost) error
	Update(ctx context.Context, post *model.DiscussionPost) error
	Delete(ctx context.Context, id int64, userID int64, force bool) error
	FindByID(ctx context.Context, id int64) (*model.DiscussionPost, error)
	List(ctx context.Context, filter PostListFilter) ([]*model.DiscussionPost, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.DiscussionPost, error)
	Search(ctx context.Context, filter PostSearchFilter) ([]*model.DiscussionPost, error)
	SearchIDs(ctx context.Context, filter PostSearchFilter) ([]int64, error)
	GetPostsByIDs(ctx context.Context, ids []int64) ([]*model.DiscussionPost, error)
	UpdateEngagementCounts(ctx context.Context, postID int64, likes, comments int64) error
}

// CommentRepository defines operations on post comments.
type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	Update(ctx context.Context, comment *model.Comment, userID int64) error
	Delete(ctx context.Context, id int64, userID int64) error
	ListByPostID(ctx context.Context, postID int64, limit, offset int) ([]*model.Comment, error)
	FindByID(ctx context.Context, id int64) (*model.Comment, error)
}

// LikeRepository defines operations on post likes.
type LikeRepository interface {
	Create(ctx context.Context, postID int64, userID int64) error
	Delete(ctx context.Context, postID int64, userID int64) (bool, error)
	Exists(ctx context.Context, postID int64, userID int64) (bool, error)
	CountByPostID(ctx context.Context, postID int64) (int64, error)
}
