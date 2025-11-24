package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/aalto-talent-network/backend/internal/file/model"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"github.com/jmoiron/sqlx"
)

type postgresFileRepository struct {
	db *sqlx.DB
}

// NewPostgresFileRepository creates a new PostgreSQL file repository
func NewPostgresFileRepository(db *sqlx.DB) FileRepository {
	return &postgresFileRepository{db: db}
}

func (r *postgresFileRepository) Create(ctx context.Context, file *model.File) error {
	query := `
		INSERT INTO files (user_id, type, object_key, url, filename, content_type, size, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	now := time.Now()
	file.CreatedAt = now
	file.UpdatedAt = now

	err := r.db.GetContext(ctx, &file.ID, query,
		file.UserID,
		file.Type,
		file.ObjectKey,
		file.URL,
		file.Filename,
		file.ContentType,
		file.Size,
		file.Metadata,
		file.CreatedAt,
		file.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresFileRepository) FindByID(ctx context.Context, id int64) (*model.File, error) {
	var file model.File
	query := `SELECT id, user_id, type, object_key, url, filename, content_type, size, metadata, created_at, updated_at
	          FROM files WHERE id = $1`
	err := r.db.GetContext(ctx, &file, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.NewAppError(errs.ErrUserNotFound, 404, "file not found").WithCode(errs.CodeUserNotFound)
		}
		return nil, err
	}
	return &file, nil
}

func (r *postgresFileRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.File, error) {
	var files []*model.File
	query := `SELECT id, user_id, type, object_key, url, filename, content_type, size, metadata, created_at, updated_at
	          FROM files WHERE user_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &files, query, userID)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (r *postgresFileRepository) FindByUserIDAndType(ctx context.Context, userID int64, fileType model.FileType) ([]*model.File, error) {
	var files []*model.File
	query := `SELECT id, user_id, type, object_key, url, filename, content_type, size, metadata, created_at, updated_at
	          FROM files WHERE user_id = $1 AND type = $2 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &files, query, userID, fileType)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (r *postgresFileRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM files WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errs.NewAppError(errs.ErrUserNotFound, 404, "file not found").WithCode(errs.CodeUserNotFound)
	}
	return nil
}

func (r *postgresFileRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM files WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *postgresFileRepository) FindByObjectKey(ctx context.Context, objectKey string) (*model.File, error) {
	var file model.File
	query := `SELECT id, user_id, type, object_key, url, filename, content_type, size, metadata, created_at, updated_at
	          FROM files WHERE object_key = $1`
	err := r.db.GetContext(ctx, &file, query, objectKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.NewAppError(errs.ErrUserNotFound, 404, "file not found").WithCode(errs.CodeUserNotFound)
		}
		return nil, err
	}
	return &file, nil
}

func (r *postgresFileRepository) Update(ctx context.Context, file *model.File) error {
	query := `
		UPDATE files 
		SET url = $1, filename = $2, content_type = $3, size = $4, metadata = $5, updated_at = $6
		WHERE id = $7
	`
	file.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		file.URL,
		file.Filename,
		file.ContentType,
		file.Size,
		file.Metadata,
		file.UpdatedAt,
		file.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errs.NewAppError(errs.ErrUserNotFound, 404, "file not found").WithCode(errs.CodeUserNotFound)
	}
	return nil
}
