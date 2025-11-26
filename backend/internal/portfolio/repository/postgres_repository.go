package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aalto-talent-network/backend/internal/portfolio/model"
	"github.com/aalto-talent-network/backend/pkg/errs"
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
	query := `SELECT id, user_id, title, description, year, tags, cover_image_url, project_link, created_at, updated_at
		FROM projects WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &projects, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by user id: %w", err)
	}

	return projects, nil
}

func (r *postgresProjectRepository) FindByID(ctx context.Context, id int64) (*model.Project, error) {
	var project model.Project
	query := `SELECT id, user_id, title, description, year, tags, cover_image_url, project_link, created_at, updated_at
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

func (r *postgresProjectRepository) Create(ctx context.Context, project *model.Project) error {
	query := `INSERT INTO projects (user_id, title, description, year, tags, cover_image_url, project_link, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		project.UserID,
		project.Title,
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
		SET title = $1, description = $2, year = $3, tags = $4, cover_image_url = $5, project_link = $6, updated_at = $7
		WHERE id = $8 AND user_id = $9
		RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		project.Title,
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



