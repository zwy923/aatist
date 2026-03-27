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

const projectSelectColumns = `id, user_id, title, short_caption, service_category, client_name, description, year, tags,
	media_urls, related_services, co_creators, cover_image_url, project_link, is_published, is_public, created_at, updated_at`

type postgresProjectRepository struct {
	db *sqlx.DB
}

func NewPostgresProjectRepository(db *sqlx.DB) ProjectRepository {
	return &postgresProjectRepository{db: db}
}

func (r *postgresProjectRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Project, error) {
	var projects []*model.Project
	query := fmt.Sprintf(`SELECT %s FROM projects WHERE user_id = $1 ORDER BY created_at DESC`, projectSelectColumns)

	err := r.db.SelectContext(ctx, &projects, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by user id: %w", err)
	}

	return projects, nil
}

func (r *postgresProjectRepository) FindPublishedPublicByUserID(ctx context.Context, userID int64) ([]*model.Project, error) {
	var projects []*model.Project
	query := fmt.Sprintf(`SELECT %s FROM projects WHERE user_id = $1 AND is_published = TRUE AND is_public = TRUE ORDER BY created_at DESC`, projectSelectColumns)

	err := r.db.SelectContext(ctx, &projects, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find public projects by user id: %w", err)
	}

	return projects, nil
}

func (r *postgresProjectRepository) FindByID(ctx context.Context, id int64) (*model.Project, error) {
	var project model.Project
	query := fmt.Sprintf(`SELECT %s FROM projects WHERE id = $1`, projectSelectColumns)

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
	query := fmt.Sprintf(`SELECT %s FROM projects ORDER BY created_at DESC LIMIT $1 OFFSET $2`, projectSelectColumns)

	err := r.db.SelectContext(ctx, &projects, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find all projects: %w", err)
	}

	return projects, nil
}

func (r *postgresProjectRepository) FindAllPublic(ctx context.Context, limit, offset int) ([]*model.Project, error) {
	projects := make([]*model.Project, 0)
	query := fmt.Sprintf(`SELECT p.%s FROM projects p
		INNER JOIN users u ON p.user_id = u.id
		WHERE u.portfolio_visibility = 'public' AND p.is_published = TRUE AND p.is_public = TRUE
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2`, joinProjectColumns("p"))

	err := r.db.SelectContext(ctx, &projects, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find public projects: %w", err)
	}

	return projects, nil
}

// joinProjectColumns prefixes each column for SELECT p.col AS col style — use explicit list with alias
func joinProjectColumns(alias string) string {
	// Build "p.id, p.user_id, ..." for INNER JOIN queries
	cols := []string{
		"id", "user_id", "title", "short_caption", "service_category", "client_name", "description", "year", "tags",
		"media_urls", "related_services", "co_creators", "cover_image_url", "project_link", "is_published", "is_public", "created_at", "updated_at",
	}
	out := ""
	for i, c := range cols {
		if i > 0 {
			out += ", "
		}
		out += alias + "." + c
	}
	return out
}

func (r *postgresProjectRepository) Create(ctx context.Context, project *model.Project) error {
	query := `INSERT INTO projects (user_id, title, short_caption, service_category, client_name, description, year, tags,
		media_urls, related_services, co_creators, cover_image_url, project_link, is_published, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id`

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		project.UserID,
		project.Title,
		project.ShortCaption,
		project.ServiceCategory,
		project.ClientName,
		project.Description,
		project.Year,
		project.Tags,
		project.MediaURLs,
		project.RelatedServices,
		project.CoCreators,
		project.CoverImageURL,
		project.ProjectLink,
		project.IsPublished,
		project.IsPublic,
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
		SET title = $1, short_caption = $2, service_category = $3, client_name = $4, description = $5, year = $6, tags = $7,
		media_urls = $8, related_services = $9, co_creators = $10, cover_image_url = $11, project_link = $12,
		is_published = $13, is_public = $14, updated_at = $15
		WHERE id = $16 AND user_id = $17
		RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		project.Title,
		project.ShortCaption,
		project.ServiceCategory,
		project.ClientName,
		project.Description,
		project.Year,
		project.Tags,
		project.MediaURLs,
		project.RelatedServices,
		project.CoCreators,
		project.CoverImageURL,
		project.ProjectLink,
		project.IsPublished,
		project.IsPublic,
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
