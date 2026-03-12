package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aatist/backend/internal/portfolio/model"
	"github.com/aatist/backend/pkg/errs"
	"github.com/jmoiron/sqlx"
)

type postgresProjectRepository struct {
	db *sqlx.DB
}

func NewPostgresProjectRepository(db *sqlx.DB) ProjectRepository {
	return &postgresProjectRepository{db: db}
}

func (r *postgresProjectRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Project, error) {
	var projects []*model.Project
	query := `SELECT id, user_id, title, service_category, client_name, description, year, tags, cover_image_url, project_link, created_at, updated_at
		FROM projects WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &projects, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by user id: %w", err)
	}

	return projects, nil
}

func (r *postgresProjectRepository) FindByID(ctx context.Context, id int64) (*model.Project, error) {
	var project model.Project
	query := `SELECT id, user_id, title, service_category, client_name, description, year, tags, cover_image_url, project_link, created_at, updated_at
		FROM projects WHERE id = $1`

	err := r.db.GetContext(ctx, &project, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find project by id: %w", err)
	}

	return &project, nil
}

func (r *postgresProjectRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.Project, error) {
	projects := make([]*model.Project, 0)
	query := `SELECT id, user_id, title, service_category, client_name, description, year, tags, cover_image_url, project_link, created_at, updated_at
		FROM projects ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	err := r.db.SelectContext(ctx, &projects, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find all projects: %w", err)
	}

	return projects, nil
}

func (r *postgresProjectRepository) FindAllPublic(ctx context.Context, limit, offset int) ([]*model.Project, error) {
	projects := make([]*model.Project, 0)
	query := `SELECT p.id, p.user_id, p.title, p.service_category, p.client_name, p.description, p.year, p.tags, p.cover_image_url, p.project_link, p.created_at, p.updated_at
		FROM projects p
		INNER JOIN users u ON p.user_id = u.id
		WHERE u.portfolio_visibility = 'public'
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2`

	err := r.db.SelectContext(ctx, &projects, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find public projects: %w", err)
	}

	return projects, nil
}

func (r *postgresProjectRepository) Create(ctx context.Context, project *model.Project) error {
	query := `INSERT INTO projects (user_id, title, service_category, client_name, description, year, tags, cover_image_url, project_link, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		project.UserID,
		project.Title,
		project.ServiceCategory,
		project.ClientName,
		project.Description,
		project.Year,
		project.Tags,
		project.CoverImageURL,
		project.ProjectLink,
		project.CreatedAt,
		project.UpdatedAt,
	).Scan(&project.ID)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

func (r *postgresProjectRepository) Update(ctx context.Context, project *model.Project) error {
	query := `UPDATE projects 
		SET title = $1, service_category = $2, client_name = $3, description = $4, year = $5, tags = $6, cover_image_url = $7, project_link = $8, updated_at = $9
		WHERE id = $10 AND user_id = $11
		RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		project.Title,
		project.ServiceCategory,
		project.ClientName,
		project.Description,
		project.Year,
		project.Tags,
		project.CoverImageURL,
		project.ProjectLink,
		time.Now(),
		project.ID,
		project.UserID,
	).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrUserNotFound
		}
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

func (r *postgresProjectRepository) Delete(ctx context.Context, id int64, userID int64) error {
	query := `DELETE FROM projects WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errs.ErrUserNotFound
	}

	return nil
}
