package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/aatist/backend/internal/community/model"
	"github.com/aatist/backend/pkg/errs"
	"github.com/jmoiron/sqlx"
)

// Postgres repositories implementation.
type (
	postgresPostRepository struct {
		db *sqlx.DB
	}
	postgresCommentRepository struct {
		db *sqlx.DB
	}
	postgresLikeRepository struct {
		db *sqlx.DB
	}
)

// Factory helpers.
func NewPostgresPostRepository(db *sqlx.DB) PostRepository {
	return &postgresPostRepository{db: db}
}

func NewPostgresCommentRepository(db *sqlx.DB) CommentRepository {
	return &postgresCommentRepository{db: db}
}

func NewPostgresLikeRepository(db *sqlx.DB) LikeRepository {
	return &postgresLikeRepository{db: db}
}

// PostRepository implementation.
func (r *postgresPostRepository) Create(ctx context.Context, post *model.DiscussionPost) error {
	query := `INSERT INTO discussion_posts (user_id, title, content, category, tags)
			  VALUES ($1, $2, $3, $4, $5)
			  RETURNING id, created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		post.UserID,
		post.Title,
		post.Content,
		post.Category,
		post.Tags,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (r *postgresPostRepository) Update(ctx context.Context, post *model.DiscussionPost) error {
	query := `UPDATE discussion_posts
			  SET title = $1,
				  content = $2,
				  category = $3,
				  tags = $4,
				  updated_at = NOW()
			  WHERE id = $5 AND user_id = $6
			  RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		post.Title,
		post.Content,
		post.Category,
		post.Tags,
		post.ID,
		post.UserID,
	).Scan(&post.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrNotFound
		}
		return fmt.Errorf("failed to update post: %w", err)
	}
	return nil
}

func (r *postgresPostRepository) Delete(ctx context.Context, id int64, userID int64, force bool) error {
	var (
		query string
		args  []interface{}
	)

	if force {
		query = `DELETE FROM discussion_posts WHERE id = $1`
		args = []interface{}{id}
	} else {
		query = `DELETE FROM discussion_posts WHERE id = $1 AND user_id = $2`
		args = []interface{}{id, userID}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch rows affected: %w", err)
	}
	if rows == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (r *postgresPostRepository) FindByID(ctx context.Context, id int64) (*model.DiscussionPost, error) {
	query := `SELECT p.id, p.user_id, p.title, p.content, p.category, p.tags, p.like_count, p.comment_count, p.created_at, p.updated_at,
	                 u.name AS author_name, u.avatar_url AS author_avatar, u.faculty AS author_faculty
			  FROM discussion_posts p
			  JOIN users u ON p.user_id = u.id
			  WHERE p.id = $1`
	var post model.DiscussionPost
	err := r.db.GetContext(ctx, &post, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find post: %w", err)
	}
	return &post, nil
}

func (r *postgresPostRepository) List(ctx context.Context, filter PostListFilter) ([]*model.DiscussionPost, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var (
		args    []interface{}
		builder strings.Builder
	)
	builder.WriteString(`SELECT p.id, p.user_id, p.title, p.content, p.category, p.tags, p.like_count, p.comment_count, p.created_at, p.updated_at,
	                            u.name AS author_name, u.avatar_url AS author_avatar, u.faculty AS author_faculty
	                     FROM discussion_posts p
	                     JOIN users u ON p.user_id = u.id`)

	if filter.Category != nil {
		args = append(args, *filter.Category)
		builder.WriteString(fmt.Sprintf(" WHERE p.category = $%d", len(args)))
	}

	builder.WriteString(" ORDER BY ")
	if filter.Sort == PostListSortOldest {
		builder.WriteString("p.created_at ASC")
	} else {
		builder.WriteString("p.created_at DESC")
	}

	args = append(args, limit)
	builder.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))
	args = append(args, offset)
	builder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))

	var posts []*model.DiscussionPost
	if err := r.db.SelectContext(ctx, &posts, builder.String(), args...); err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}
	return posts, nil
}

func (r *postgresPostRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.DiscussionPost, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT p.id, p.user_id, p.title, p.content, p.category, p.tags, p.like_count, p.comment_count, p.created_at, p.updated_at,
	                 u.name AS author_name, u.avatar_url AS author_avatar, u.faculty AS author_faculty
			  FROM discussion_posts p
			  JOIN users u ON p.user_id = u.id
			  WHERE p.user_id = $1
			  ORDER BY p.created_at DESC
			  LIMIT $2 OFFSET $3`

	var posts []*model.DiscussionPost
	if err := r.db.SelectContext(ctx, &posts, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list user posts: %w", err)
	}
	return posts, nil
}

func (r *postgresPostRepository) Search(ctx context.Context, filter PostSearchFilter) ([]*model.DiscussionPost, error) {
	query := strings.TrimSpace(filter.Query)
	if query == "" {
		return r.List(ctx, PostListFilter{Limit: filter.Limit, Offset: filter.Offset})
	}

	limit := filter.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	// Use TSVector for content and trigram similarity for title.
	sqlQuery := `
		SELECT p.id, p.user_id, p.title, p.content, p.category, p.tags, p.like_count, p.comment_count, p.created_at, p.updated_at,
		       u.name AS author_name, u.avatar_url AS author_avatar, u.faculty AS author_faculty
		FROM discussion_posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.tsv @@ plainto_tsquery('english', $1)
		   OR p.title % $1
		ORDER BY (
			ts_rank_cd(p.tsv, plainto_tsquery('english', $1))
			+ similarity(p.title, $1)
		) DESC, p.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryxContext(ctx, sqlQuery, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}
	defer rows.Close()

	var posts []*model.DiscussionPost
	for rows.Next() {
		var post model.DiscussionPost
		if err := rows.StructScan(&post); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *postgresPostRepository) SearchIDs(ctx context.Context, filter PostSearchFilter) ([]int64, error) {
	query := strings.TrimSpace(filter.Query)
	if query == "" {
		posts, err := r.List(ctx, PostListFilter{Limit: filter.Limit, Offset: filter.Offset})
		if err != nil {
			return nil, err
		}
		ids := make([]int64, 0, len(posts))
		for _, p := range posts {
			ids = append(ids, p.ID)
		}
		return ids, nil
	}

	limit := filter.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	sqlQuery := `
		SELECT id
		FROM discussion_posts
		WHERE tsv @@ plainto_tsquery('english', $1)
		   OR title % $1
		ORDER BY (
			ts_rank_cd(tsv, plainto_tsquery('english', $1))
			+ similarity(title, $1)
		) DESC, created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryxContext(ctx, sqlQuery, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search post ids: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan search id row: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *postgresPostRepository) GetPostsByIDs(ctx context.Context, ids []int64) ([]*model.DiscussionPost, error) {
	if len(ids) == 0 {
		return []*model.DiscussionPost{}, nil
	}

	query, args, err := sqlx.In(`SELECT p.id, p.user_id, p.title, p.content, p.category, p.tags, p.like_count, p.comment_count, p.created_at, p.updated_at,
	                            u.name AS author_name, u.avatar_url AS author_avatar, u.faculty AS author_faculty
		FROM discussion_posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id IN (?)`, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to build IN query: %w", err)
	}
	query = r.db.Rebind(query)

	var posts []*model.DiscussionPost
	if err := r.db.SelectContext(ctx, &posts, query, args...); err != nil {
		return nil, fmt.Errorf("failed to fetch posts by ids: %w", err)
	}

	// Preserve input order.
	postMap := make(map[int64]*model.DiscussionPost, len(posts))
	for _, post := range posts {
		postMap[post.ID] = post
	}
	ordered := make([]*model.DiscussionPost, 0, len(ids))
	for _, id := range ids {
		if p, ok := postMap[id]; ok {
			ordered = append(ordered, p)
		}
	}
	return ordered, nil
}

func (r *postgresPostRepository) UpdateEngagementCounts(ctx context.Context, postID int64, likes, comments int64) error {
	query := `UPDATE discussion_posts 
			  SET like_count = $1, 
			      comment_count = $2, 
				  updated_at = NOW() 
			  WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, likes, comments, postID)
	return err
}

// CommentRepository implementation.
func (r *postgresCommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	query := `INSERT INTO post_comments (post_id, user_id, parent_id, content)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id, created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		comment.PostID,
		comment.UserID,
		comment.ParentID,
		comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
}

func (r *postgresCommentRepository) Update(ctx context.Context, comment *model.Comment, userID int64) error {
	query := `UPDATE post_comments
			  SET content = $1,
			      updated_at = NOW()
			  WHERE id = $2 AND user_id = $3
			  RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		comment.Content,
		comment.ID,
		userID,
	).Scan(&comment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrNotFound
		}
		return fmt.Errorf("failed to update comment: %w", err)
	}
	return nil
}

func (r *postgresCommentRepository) Delete(ctx context.Context, id int64, userID int64) error {
	query := `DELETE FROM post_comments WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch rows affected: %w", err)
	}
	if rows == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (r *postgresCommentRepository) ListByPostID(ctx context.Context, postID int64, limit, offset int) ([]*model.Comment, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at, c.updated_at,
	                 u.name AS author_name, u.avatar_url AS author_avatar
			  FROM post_comments c
			  JOIN users u ON c.user_id = u.id
			  WHERE c.post_id = $1
			  ORDER BY c.created_at ASC
			  LIMIT $2 OFFSET $3`

	var comments []*model.Comment
	if err := r.db.SelectContext(ctx, &comments, query, postID, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	return comments, nil
}

func (r *postgresCommentRepository) FindByID(ctx context.Context, id int64) (*model.Comment, error) {
	query := `SELECT c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at, c.updated_at,
	                 u.name AS author_name, u.avatar_url AS author_avatar
			  FROM post_comments c
			  JOIN users u ON c.user_id = u.id
			  WHERE c.id = $1`

	var comment model.Comment
	if err := r.db.GetContext(ctx, &comment, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find comment: %w", err)
	}
	return &comment, nil
}

// LikeRepository implementation.
func (r *postgresLikeRepository) Create(ctx context.Context, postID int64, userID int64) error {
	query := `INSERT INTO post_likes (post_id, user_id) VALUES ($1, $2)
			  ON CONFLICT (post_id, user_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, postID, userID)
	if err != nil {
		return fmt.Errorf("failed to create like: %w", err)
	}
	return nil
}

func (r *postgresLikeRepository) Delete(ctx context.Context, postID int64, userID int64) (bool, error) {
	query := `DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, postID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to delete like: %w", err)
	}
	rows, _ := res.RowsAffected()
	return rows > 0, nil
}

func (r *postgresLikeRepository) Exists(ctx context.Context, postID int64, userID int64) (bool, error) {
	query := `SELECT 1 FROM post_likes WHERE post_id = $1 AND user_id = $2`
	var exists int
	err := r.db.GetContext(ctx, &exists, query, postID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check like existence: %w", err)
	}
	return true, nil
}

func (r *postgresLikeRepository) CountByPostID(ctx context.Context, postID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM post_likes WHERE post_id = $1`
	var count int64
	if err := r.db.GetContext(ctx, &count, query, postID); err != nil {
		return 0, fmt.Errorf("failed to count likes: %w", err)
	}
	return count, nil
}
