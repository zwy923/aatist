package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aalto-talent-network/backend/internal/event/model"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"github.com/jmoiron/sqlx"
)

const eventSelectColumns = `id, title, organizer, type_tags, school_tags, is_external, is_free,
	location, languages, start_time, end_time, max_participants, description, created_by,
	published_at, status, cover_image_url, created_at, updated_at`

type (
	postgresEventRepository struct {
		db *sqlx.DB
	}
	postgresEventInterestRepository struct {
		db *sqlx.DB
	}
	postgresEventGoingRepository struct {
		db *sqlx.DB
	}
	postgresEventCommentRepository struct {
		db *sqlx.DB
	}
)

// Factory functions
func NewPostgresEventRepository(db *sqlx.DB) EventRepository {
	return &postgresEventRepository{db: db}
}

func NewPostgresEventInterestRepository(db *sqlx.DB) EventInterestRepository {
	return &postgresEventInterestRepository{db: db}
}

func NewPostgresEventGoingRepository(db *sqlx.DB) EventGoingRepository {
	return &postgresEventGoingRepository{db: db}
}

func NewPostgresEventCommentRepository(db *sqlx.DB) EventCommentRepository {
	return &postgresEventCommentRepository{db: db}
}

// EventRepository implementation
func (r *postgresEventRepository) Create(ctx context.Context, event *model.Event) error {
	query := `
		INSERT INTO events (title, organizer, type_tags, school_tags, is_external, is_free,
			location, languages, start_time, end_time, max_participants, description, created_by,
			published_at, status, cover_image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id
	`
	now := time.Now()
	event.CreatedAt = now
	event.UpdatedAt = now
	if event.PublishedAt.IsZero() {
		event.PublishedAt = now
	}

	err := r.db.GetContext(ctx, &event.ID, query,
		event.Title,
		event.Organizer,
		event.TypeTags,
		event.SchoolTags,
		event.IsExternal,
		event.IsFree,
		event.Location,
		event.Languages,
		event.StartTime,
		event.EndTime,
		event.MaxParticipants,
		event.Description,
		event.CreatedBy,
		event.PublishedAt,
		event.Status,
		event.CoverImageURL,
		event.CreatedAt,
		event.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}
	return nil
}

func (r *postgresEventRepository) Update(ctx context.Context, event *model.Event) error {
	query := `
		UPDATE events
		SET title = $1, organizer = $2, type_tags = $3, school_tags = $4, is_external = $5,
			is_free = $6, location = $7, languages = $8, start_time = $9, end_time = $10,
			max_participants = $11, description = $12, status = $13, cover_image_url = $14, updated_at = $15
		WHERE id = $16 AND created_by = $17
		RETURNING updated_at
	`
	event.UpdatedAt = time.Now()

	err := r.db.GetContext(ctx, &event.UpdatedAt, query,
		event.Title,
		event.Organizer,
		event.TypeTags,
		event.SchoolTags,
		event.IsExternal,
		event.IsFree,
		event.Location,
		event.Languages,
		event.StartTime,
		event.EndTime,
		event.MaxParticipants,
		event.Description,
		event.Status,
		event.CoverImageURL,
		event.UpdatedAt,
		event.ID,
		event.CreatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrNotFound
		}
		return fmt.Errorf("failed to update event: %w", err)
	}
	return nil
}

func (r *postgresEventRepository) FindByID(ctx context.Context, id int64) (*model.Event, error) {
	query := fmt.Sprintf("SELECT %s FROM events WHERE id = $1", eventSelectColumns)
	var event model.Event
	err := r.db.GetContext(ctx, &event, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find event: %w", err)
	}
	return &event, nil
}

func (r *postgresEventRepository) List(ctx context.Context, filter EventListFilter) ([]*model.Event, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (filter.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	var (
		args    []interface{}
		builder strings.Builder
		where   []string
	)

	// Build SELECT with computed fields
	builder.WriteString(fmt.Sprintf(`
		SELECT %s,
			(SELECT COUNT(*) FROM event_interests WHERE event_id = events.id) as interested_count,
			(SELECT COUNT(*) FROM event_going WHERE event_id = events.id) as going_count,
			(SELECT COUNT(*) FROM event_comments WHERE event_id = events.id) as comment_count
		FROM events`, eventSelectColumns))

	// Build WHERE clause
	if filter.Search != nil && *filter.Search != "" {
		args = append(args, *filter.Search)
		where = append(where, fmt.Sprintf("to_tsvector('english', title || ' ' || organizer) @@ plainto_tsquery('english', $%d)", len(args)))
	}

	if len(filter.Types) > 0 {
		args = append(args, filter.Types)
		where = append(where, fmt.Sprintf("type_tags && $%d", len(args)))
	}

	if len(filter.Schools) > 0 {
		args = append(args, filter.Schools)
		where = append(where, fmt.Sprintf("school_tags && $%d", len(args)))
	}

	if filter.IsFree != nil {
		args = append(args, *filter.IsFree)
		where = append(where, fmt.Sprintf("is_free = $%d", len(args)))
	}

	if len(filter.Languages) > 0 {
		args = append(args, filter.Languages)
		where = append(where, fmt.Sprintf("languages && $%d", len(args)))
	}

	if filter.Location != nil && *filter.Location != "" {
		args = append(args, *filter.Location)
		where = append(where, fmt.Sprintf("location = $%d", len(args)))
	}

	if filter.Status != nil && *filter.Status != "" {
		args = append(args, *filter.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	} else {
		where = append(where, "status = 'active'")
	}

	// Time filter
	if filter.TimeFilter != nil {
		switch *filter.TimeFilter {
		case TimeFilterToday:
			where = append(where, "DATE(start_time) = CURRENT_DATE")
		case TimeFilterThisWeek:
			// Monday to Sunday of current week
			where = append(where, "start_time >= date_trunc('week', CURRENT_DATE) AND start_time < date_trunc('week', CURRENT_DATE) + interval '7 days'")
		case TimeFilterThisMonth:
			where = append(where, "start_time >= date_trunc('month', CURRENT_DATE) AND start_time < date_trunc('month', CURRENT_DATE) + interval '1 month'")
		}
	}

	if len(where) > 0 {
		builder.WriteString(" WHERE " + strings.Join(where, " AND "))
	}

	// Build ORDER BY clause
	builder.WriteString(" ORDER BY ")
	switch filter.Sort {
	case SortStartingSoon:
		builder.WriteString("start_time ASC")
	case SortPopular:
		builder.WriteString("interested_count DESC")
	default: // SortNewEvents
		builder.WriteString("published_at DESC")
	}

	// Add LIMIT and OFFSET
	args = append(args, limit)
	builder.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))
	args = append(args, offset)
	builder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))

	var events []*model.Event
	rows, err := r.db.QueryxContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event model.Event
		var interestedCount, goingCount, commentCount sql.NullInt64

		err := rows.Scan(
			&event.ID, &event.Title, &event.Organizer, &event.TypeTags, &event.SchoolTags,
			&event.IsExternal, &event.IsFree, &event.Location, &event.Languages,
			&event.StartTime, &event.EndTime, &event.MaxParticipants, &event.Description,
			&event.CreatedBy, &event.PublishedAt, &event.Status, &event.CoverImageURL,
			&event.CreatedAt, &event.UpdatedAt,
			&interestedCount, &goingCount, &commentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if interestedCount.Valid {
			event.InterestedCount = &interestedCount.Int64
		}
		if goingCount.Valid {
			event.GoingCount = &goingCount.Int64
		}
		if commentCount.Valid {
			event.CommentCount = &commentCount.Int64
		}

		events = append(events, &event)
	}

	return events, nil
}

func (r *postgresEventRepository) Count(ctx context.Context, filter EventListFilter) (int64, error) {
	var (
		args    []interface{}
		builder strings.Builder
		where   []string
	)

	builder.WriteString("SELECT COUNT(*) FROM events")

	// Build WHERE clause (same as List)
	if filter.Search != nil && *filter.Search != "" {
		args = append(args, *filter.Search)
		where = append(where, fmt.Sprintf("to_tsvector('english', title || ' ' || organizer) @@ plainto_tsquery('english', $%d)", len(args)))
	}

	if len(filter.Types) > 0 {
		args = append(args, filter.Types)
		where = append(where, fmt.Sprintf("type_tags && $%d", len(args)))
	}

	if len(filter.Schools) > 0 {
		args = append(args, filter.Schools)
		where = append(where, fmt.Sprintf("school_tags && $%d", len(args)))
	}

	if filter.IsFree != nil {
		args = append(args, *filter.IsFree)
		where = append(where, fmt.Sprintf("is_free = $%d", len(args)))
	}

	if len(filter.Languages) > 0 {
		args = append(args, filter.Languages)
		where = append(where, fmt.Sprintf("languages && $%d", len(args)))
	}

	if filter.Location != nil && *filter.Location != "" {
		args = append(args, *filter.Location)
		where = append(where, fmt.Sprintf("location = $%d", len(args)))
	}

	if filter.Status != nil && *filter.Status != "" {
		args = append(args, *filter.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	} else {
		where = append(where, "status = 'active'")
	}

	// Time filter
	if filter.TimeFilter != nil {
		switch *filter.TimeFilter {
		case TimeFilterToday:
			where = append(where, "DATE(start_time) = CURRENT_DATE")
		case TimeFilterThisWeek:
			where = append(where, "start_time >= date_trunc('week', CURRENT_DATE) AND start_time < date_trunc('week', CURRENT_DATE) + interval '7 days'")
		case TimeFilterThisMonth:
			where = append(where, "start_time >= date_trunc('month', CURRENT_DATE) AND start_time < date_trunc('month', CURRENT_DATE) + interval '1 month'")
		}
	}

	if len(where) > 0 {
		builder.WriteString(" WHERE " + strings.Join(where, " AND "))
	}

	var count int64
	if err := r.db.GetContext(ctx, &count, builder.String(), args...); err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}
	return count, nil
}

func (r *postgresEventRepository) Delete(ctx context.Context, id int64, userID int64) error {
	query := `UPDATE events SET status = 'canceled' WHERE id = $1 AND created_by = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
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

func (r *postgresEventRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.Event, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`SELECT %s FROM events 
		WHERE created_by = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`, eventSelectColumns)

	var events []*model.Event
	if err := r.db.SelectContext(ctx, &events, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list user events: %w", err)
	}
	return events, nil
}

// EventInterestRepository implementation
func (r *postgresEventInterestRepository) Create(ctx context.Context, userID, eventID int64) error {
	query := `INSERT INTO event_interests (user_id, event_id, created_at) 
		VALUES ($1, $2, $3) ON CONFLICT (user_id, event_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, userID, eventID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create interest: %w", err)
	}
	return nil
}

func (r *postgresEventInterestRepository) Delete(ctx context.Context, userID, eventID int64) error {
	query := `DELETE FROM event_interests WHERE user_id = $1 AND event_id = $2`
	result, err := r.db.ExecContext(ctx, query, userID, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete interest: %w", err)
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

func (r *postgresEventInterestRepository) Exists(ctx context.Context, userID, eventID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM event_interests WHERE user_id = $1 AND event_id = $2)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, userID, eventID)
	if err != nil {
		return false, fmt.Errorf("failed to check interest existence: %w", err)
	}
	return exists, nil
}

func (r *postgresEventInterestRepository) CountByEventID(ctx context.Context, eventID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM event_interests WHERE event_id = $1`
	var count int64
	err := r.db.GetContext(ctx, &count, query, eventID)
	if err != nil {
		return 0, fmt.Errorf("failed to count interests: %w", err)
	}
	return count, nil
}

// EventGoingRepository implementation
func (r *postgresEventGoingRepository) Create(ctx context.Context, userID, eventID int64) error {
	query := `INSERT INTO event_going (user_id, event_id, created_at) 
		VALUES ($1, $2, $3) ON CONFLICT (user_id, event_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, userID, eventID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create going: %w", err)
	}
	return nil
}

func (r *postgresEventGoingRepository) Delete(ctx context.Context, userID, eventID int64) error {
	query := `DELETE FROM event_going WHERE user_id = $1 AND event_id = $2`
	result, err := r.db.ExecContext(ctx, query, userID, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete going: %w", err)
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

func (r *postgresEventGoingRepository) Exists(ctx context.Context, userID, eventID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM event_going WHERE user_id = $1 AND event_id = $2)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, userID, eventID)
	if err != nil {
		return false, fmt.Errorf("failed to check going existence: %w", err)
	}
	return exists, nil
}

func (r *postgresEventGoingRepository) CountByEventID(ctx context.Context, eventID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM event_going WHERE event_id = $1`
	var count int64
	err := r.db.GetContext(ctx, &count, query, eventID)
	if err != nil {
		return 0, fmt.Errorf("failed to count going: %w", err)
	}
	return count, nil
}

// EventCommentRepository implementation
func (r *postgresEventCommentRepository) Create(ctx context.Context, comment *model.EventComment) error {
	query := `
		INSERT INTO event_comments (event_id, user_id, content, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	comment.CreatedAt = time.Now()

	err := r.db.GetContext(ctx, &comment.ID, query,
		comment.EventID,
		comment.UserID,
		comment.Content,
		comment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	return nil
}

func (r *postgresEventCommentRepository) FindByID(ctx context.Context, id int64) (*model.EventComment, error) {
	query := `SELECT id, event_id, user_id, content, created_at
		FROM event_comments WHERE id = $1`
	var comment model.EventComment
	err := r.db.GetContext(ctx, &comment, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find comment: %w", err)
	}
	return &comment, nil
}

func (r *postgresEventCommentRepository) ListByEventID(ctx context.Context, eventID int64, limit, offset int) ([]*model.EventComment, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, event_id, user_id, content, created_at
		FROM event_comments 
		WHERE event_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	var comments []*model.EventComment
	if err := r.db.SelectContext(ctx, &comments, query, eventID, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list event comments: %w", err)
	}
	return comments, nil
}

func (r *postgresEventCommentRepository) Delete(ctx context.Context, id int64, userID int64) error {
	query := `DELETE FROM event_comments WHERE id = $1 AND user_id = $2`
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

func (r *postgresEventCommentRepository) CountByEventID(ctx context.Context, eventID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM event_comments WHERE event_id = $1`
	var count int64
	err := r.db.GetContext(ctx, &count, query, eventID)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}
	return count, nil
}
