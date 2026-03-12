package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aatist/backend/internal/user/model"
	"github.com/jmoiron/sqlx"
)

// UserServiceRepository defines operations for user service offerings
type UserServiceRepository interface {
	FindByUserID(ctx context.Context, userID int64) ([]*model.UserService, error)
	Create(ctx context.Context, s *model.UserService) error
	Update(ctx context.Context, s *model.UserService) error
	Delete(ctx context.Context, id, userID int64) error
	FindByID(ctx context.Context, id, userID int64) (*model.UserService, error)
}

type postgresUserServiceRepository struct {
	db *sqlx.DB
}

// NewPostgresUserServiceRepository creates a new PostgreSQL user service repository
func NewPostgresUserServiceRepository(db *sqlx.DB) UserServiceRepository {
	return &postgresUserServiceRepository{db: db}
}

func (r *postgresUserServiceRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.UserService, error) {
	var services []*model.UserService
	query := `SELECT id, user_id, category, experience_summary, title, description, short_description,
		price_type, price_min, price_max, media_urls, created_at, updated_at
		FROM user_services WHERE user_id = $1 ORDER BY created_at ASC`
	err := r.db.SelectContext(ctx, &services, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user services: %w", err)
	}
	return services, nil
}

func (r *postgresUserServiceRepository) Create(ctx context.Context, s *model.UserService) error {
	query := `INSERT INTO user_services (user_id, category, experience_summary, title, description, short_description,
		price_type, price_min, price_max, media_urls, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowxContext(ctx, query,
		s.UserID, s.Category, s.ExperienceSummary, s.Title, s.Description, s.ShortDescription,
		s.PriceType, s.PriceMin, s.PriceMax, s.MediaURLs,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *postgresUserServiceRepository) Update(ctx context.Context, s *model.UserService) error {
	query := `UPDATE user_services SET category = $1, experience_summary = $2, title = $3, description = $4,
		short_description = $5, price_type = $6, price_min = $7, price_max = $8, media_urls = $9, updated_at = NOW()
		WHERE id = $10 AND user_id = $11
		RETURNING updated_at`
	result := r.db.QueryRowxContext(ctx, query,
		s.Category, s.ExperienceSummary, s.Title, s.Description, s.ShortDescription,
		s.PriceType, s.PriceMin, s.PriceMax, s.MediaURLs, s.ID, s.UserID,
	)
	if err := result.Scan(&s.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user service not found")
		}
		return err
	}
	return nil
}

func (r *postgresUserServiceRepository) Delete(ctx context.Context, id, userID int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM user_services WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user service: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user service not found")
	}
	return nil
}

func (r *postgresUserServiceRepository) FindByID(ctx context.Context, id, userID int64) (*model.UserService, error) {
	var s model.UserService
	query := `SELECT id, user_id, category, experience_summary, title, description, short_description,
		price_type, price_min, price_max, media_urls, created_at, updated_at
		FROM user_services WHERE id = $1 AND user_id = $2`
	err := r.db.GetContext(ctx, &s, query, id, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user service not found")
		}
		return nil, err
	}
	return &s, nil
}
