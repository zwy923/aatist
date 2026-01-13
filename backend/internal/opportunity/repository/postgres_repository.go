package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aatist/backend/internal/opportunity/model"
	"github.com/aatist/backend/pkg/errs"
	"github.com/jmoiron/sqlx"
)

const opportunitySelectColumns = `id, title, organization, category, budget_type, budget_value, 
	location, duration_months, languages, start_date, published_at, urgent, description, tags, 
	created_by, status, created_at, updated_at`

type (
	postgresOpportunityRepository struct {
		db *sqlx.DB
	}
	postgresOpportunityApplicationRepository struct {
		db *sqlx.DB
	}
)

// Factory functions
func NewPostgresOpportunityRepository(db *sqlx.DB) OpportunityRepository {
	return &postgresOpportunityRepository{db: db}
}

func NewPostgresOpportunityApplicationRepository(db *sqlx.DB) OpportunityApplicationRepository {
	return &postgresOpportunityApplicationRepository{db: db}
}

// OpportunityRepository implementation
func (r *postgresOpportunityRepository) Create(ctx context.Context, opp *model.Opportunity) error {
	query := `
		INSERT INTO opportunities (title, organization, category, budget_type, budget_value, 
			location, duration_months, languages, start_date, published_at, urgent, description, 
			tags, created_by, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id
	`
	now := time.Now()
	opp.CreatedAt = now
	opp.UpdatedAt = now
	if opp.PublishedAt.IsZero() {
		opp.PublishedAt = now
	}

	err := r.db.GetContext(ctx, &opp.ID, query,
		opp.Title,
		opp.Organization,
		opp.Category,
		opp.BudgetType,
		opp.BudgetValue,
		opp.Location,
		opp.DurationMonths,
		opp.Languages,
		opp.StartDate,
		opp.PublishedAt,
		opp.Urgent,
		opp.Description,
		opp.Tags,
		opp.CreatedBy,
		opp.Status,
		opp.CreatedAt,
		opp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create opportunity: %w", err)
	}
	return nil
}

func (r *postgresOpportunityRepository) Update(ctx context.Context, opp *model.Opportunity) error {
	query := `
		UPDATE opportunities
		SET title = $1, organization = $2, category = $3, budget_type = $4, budget_value = $5,
			location = $6, duration_months = $7, languages = $8, start_date = $9, urgent = $10,
			description = $11, tags = $12, status = $13, updated_at = $14
		WHERE id = $15 AND created_by = $16
		RETURNING updated_at
	`
	opp.UpdatedAt = time.Now()

	err := r.db.GetContext(ctx, &opp.UpdatedAt, query,
		opp.Title,
		opp.Organization,
		opp.Category,
		opp.BudgetType,
		opp.BudgetValue,
		opp.Location,
		opp.DurationMonths,
		opp.Languages,
		opp.StartDate,
		opp.Urgent,
		opp.Description,
		opp.Tags,
		opp.Status,
		opp.UpdatedAt,
		opp.ID,
		opp.CreatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrNotFound
		}
		return fmt.Errorf("failed to update opportunity: %w", err)
	}
	return nil
}

func (r *postgresOpportunityRepository) FindByID(ctx context.Context, id int64) (*model.Opportunity, error) {
	query := fmt.Sprintf("SELECT %s FROM opportunities WHERE id = $1", opportunitySelectColumns)
	var opp model.Opportunity
	err := r.db.GetContext(ctx, &opp, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find opportunity: %w", err)
	}
	return &opp, nil
}

func (r *postgresOpportunityRepository) List(ctx context.Context, filter OpportunityListFilter) ([]*model.Opportunity, error) {
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

	builder.WriteString(fmt.Sprintf("SELECT %s FROM opportunities", opportunitySelectColumns))

	// Build WHERE clause
	if filter.Category != nil && *filter.Category != "" {
		args = append(args, *filter.Category)
		where = append(where, fmt.Sprintf("category = $%d", len(args)))
	}

	if filter.Location != nil && *filter.Location != "" {
		args = append(args, *filter.Location)
		where = append(where, fmt.Sprintf("location = $%d", len(args)))
	}

	if filter.BudgetMin != nil {
		args = append(args, *filter.BudgetMin)
		where = append(where, fmt.Sprintf("budget_value >= $%d", len(args)))
	}

	if filter.BudgetMax != nil {
		args = append(args, *filter.BudgetMax)
		where = append(where, fmt.Sprintf("budget_value <= $%d", len(args)))
	}

	if filter.StartDateFrom != nil {
		args = append(args, *filter.StartDateFrom)
		where = append(where, fmt.Sprintf("start_date >= $%d", len(args)))
	}

	if filter.StartDateTo != nil {
		args = append(args, *filter.StartDateTo)
		where = append(where, fmt.Sprintf("start_date <= $%d", len(args)))
	}

	if len(filter.Languages) > 0 {
		args = append(args, filter.Languages)
		where = append(where, fmt.Sprintf("languages && $%d", len(args)))
	}

	if filter.Urgent != nil {
		args = append(args, *filter.Urgent)
		where = append(where, fmt.Sprintf("urgent = $%d", len(args)))
	}

	if filter.Status != nil && *filter.Status != "" {
		args = append(args, *filter.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	} else {
		// Default to active only
		where = append(where, "status = 'active'")
	}

	if len(where) > 0 {
		builder.WriteString(" WHERE " + strings.Join(where, " AND "))
	}

	// Build ORDER BY clause
	builder.WriteString(" ORDER BY ")
	switch filter.Sort {
	case SortStartDate:
		builder.WriteString("start_date")
	case SortBudget:
		builder.WriteString("budget_value")
	default:
		builder.WriteString("published_at")
	}

	if filter.Order == "asc" {
		builder.WriteString(" ASC")
	} else {
		builder.WriteString(" DESC")
	}

	// Add NULLS LAST for optional fields
	if filter.Sort == SortStartDate || filter.Sort == SortBudget {
		builder.WriteString(" NULLS LAST")
	}

	// Add LIMIT and OFFSET
	args = append(args, limit)
	builder.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))
	args = append(args, offset)
	builder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))

	var opportunities []*model.Opportunity
	if err := r.db.SelectContext(ctx, &opportunities, builder.String(), args...); err != nil {
		return nil, fmt.Errorf("failed to list opportunities: %w", err)
	}
	return opportunities, nil
}

func (r *postgresOpportunityRepository) Count(ctx context.Context, filter OpportunityListFilter) (int64, error) {
	var (
		args    []interface{}
		builder strings.Builder
		where   []string
	)

	builder.WriteString("SELECT COUNT(*) FROM opportunities")

	// Build WHERE clause (same as List)
	if filter.Category != nil && *filter.Category != "" {
		args = append(args, *filter.Category)
		where = append(where, fmt.Sprintf("category = $%d", len(args)))
	}

	if filter.Location != nil && *filter.Location != "" {
		args = append(args, *filter.Location)
		where = append(where, fmt.Sprintf("location = $%d", len(args)))
	}

	if filter.BudgetMin != nil {
		args = append(args, *filter.BudgetMin)
		where = append(where, fmt.Sprintf("budget_value >= $%d", len(args)))
	}

	if filter.BudgetMax != nil {
		args = append(args, *filter.BudgetMax)
		where = append(where, fmt.Sprintf("budget_value <= $%d", len(args)))
	}

	if filter.StartDateFrom != nil {
		args = append(args, *filter.StartDateFrom)
		where = append(where, fmt.Sprintf("start_date >= $%d", len(args)))
	}

	if filter.StartDateTo != nil {
		args = append(args, *filter.StartDateTo)
		where = append(where, fmt.Sprintf("start_date <= $%d", len(args)))
	}

	if len(filter.Languages) > 0 {
		args = append(args, filter.Languages)
		where = append(where, fmt.Sprintf("languages && $%d", len(args)))
	}

	if filter.Urgent != nil {
		args = append(args, *filter.Urgent)
		where = append(where, fmt.Sprintf("urgent = $%d", len(args)))
	}

	if filter.Status != nil && *filter.Status != "" {
		args = append(args, *filter.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	} else {
		where = append(where, "status = 'active'")
	}

	if len(where) > 0 {
		builder.WriteString(" WHERE " + strings.Join(where, " AND "))
	}

	var count int64
	if err := r.db.GetContext(ctx, &count, builder.String(), args...); err != nil {
		return 0, fmt.Errorf("failed to count opportunities: %w", err)
	}
	return count, nil
}

func (r *postgresOpportunityRepository) Delete(ctx context.Context, id int64, userID int64) error {
	query := `UPDATE opportunities SET status = 'closed' WHERE id = $1 AND created_by = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete opportunity: %w", err)
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

func (r *postgresOpportunityRepository) ListByUserID(ctx context.Context, userID int64, status *string, limit, offset int) ([]*model.Opportunity, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var (
		args    []interface{}
		builder strings.Builder
	)

	builder.WriteString(fmt.Sprintf("SELECT %s FROM opportunities WHERE created_by = $1", opportunitySelectColumns))
	args = append(args, userID)

	if status != nil && *status != "" {
		args = append(args, *status)
		builder.WriteString(fmt.Sprintf(" AND status = $%d", len(args)))
	}

	builder.WriteString(" ORDER BY created_at DESC")

	args = append(args, limit)
	builder.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))

	args = append(args, offset)
	builder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))

	var opportunities []*model.Opportunity
	if err := r.db.SelectContext(ctx, &opportunities, builder.String(), args...); err != nil {
		return nil, fmt.Errorf("failed to list user opportunities: %w", err)
	}
	return opportunities, nil
}

func (r *postgresOpportunityRepository) UpdateStatus(ctx context.Context, id int64, userID int64, status model.OpportunityStatus) error {
	query := `UPDATE opportunities SET status = $1, updated_at = $2 WHERE id = $3 AND created_by = $4`
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id, userID)
	if err != nil {
		return fmt.Errorf("failed to update opportunity status: %w", err)
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

func (r *postgresOpportunityRepository) GetStats(ctx context.Context, id int64, userID int64) (*OpportunityStats, error) {
	// Verify ownership first
	var exists bool
	err := r.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM opportunities WHERE id = $1 AND created_by = $2)", id, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrNotFound
	}

	var stats OpportunityStats
	query := `SELECT COUNT(*) FROM opportunity_applications WHERE opportunity_id = $1`
	err = r.db.GetContext(ctx, &stats.ApplicationCount, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get opportunity stats: %w", err)
	}

	return &stats, nil
}

// OpportunityApplicationRepository implementation
func (r *postgresOpportunityApplicationRepository) Create(ctx context.Context, app *model.OpportunityApplication) error {
	query := `
		INSERT INTO opportunity_applications (user_id, opportunity_id, message, cv_url, portfolio_url, 
			status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now

	err := r.db.GetContext(ctx, &app.ID, query,
		app.UserID,
		app.OpportunityID,
		app.Message,
		app.CVURL,
		app.PortfolioURL,
		app.Status,
		app.CreatedAt,
		app.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}
	return nil
}

func (r *postgresOpportunityApplicationRepository) Update(ctx context.Context, app *model.OpportunityApplication) error {
	query := `
		UPDATE opportunity_applications
		SET message = $1, cv_url = $2, portfolio_url = $3, status = $4, updated_at = $5
		WHERE id = $6
		RETURNING updated_at
	`
	app.UpdatedAt = time.Now()

	err := r.db.GetContext(ctx, &app.UpdatedAt, query,
		app.Message,
		app.CVURL,
		app.PortfolioURL,
		app.Status,
		app.UpdatedAt,
		app.ID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrNotFound
		}
		return fmt.Errorf("failed to update application: %w", err)
	}
	return nil
}

func (r *postgresOpportunityApplicationRepository) FindByID(ctx context.Context, id int64) (*model.OpportunityApplication, error) {
	query := `SELECT id, user_id, opportunity_id, message, cv_url, portfolio_url, status, created_at, updated_at
		FROM opportunity_applications WHERE id = $1`
	var app model.OpportunityApplication
	err := r.db.GetContext(ctx, &app, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find application: %w", err)
	}
	return &app, nil
}

func (r *postgresOpportunityApplicationRepository) FindByUserAndOpportunity(ctx context.Context, userID, opportunityID int64) (*model.OpportunityApplication, error) {
	query := `SELECT id, user_id, opportunity_id, message, cv_url, portfolio_url, status, created_at, updated_at
		FROM opportunity_applications WHERE user_id = $1 AND opportunity_id = $2`
	var app model.OpportunityApplication
	err := r.db.GetContext(ctx, &app, query, userID, opportunityID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find application: %w", err)
	}
	return &app, nil
}

func (r *postgresOpportunityApplicationRepository) ListByUserID(ctx context.Context, userID int64, status *string, limit, offset int) ([]*model.OpportunityApplication, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var (
		args    []interface{}
		builder strings.Builder
	)

	builder.WriteString("SELECT id, user_id, opportunity_id, message, cv_url, portfolio_url, status, created_at, updated_at FROM opportunity_applications WHERE user_id = $1")
	args = append(args, userID)

	if status != nil && *status != "" {
		args = append(args, *status)
		builder.WriteString(fmt.Sprintf(" AND status = $%d", len(args)))
	}

	builder.WriteString(" ORDER BY created_at DESC")

	args = append(args, limit)
	builder.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))

	args = append(args, offset)
	builder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))

	var applications []*model.OpportunityApplication
	if err := r.db.SelectContext(ctx, &applications, builder.String(), args...); err != nil {
		return nil, fmt.Errorf("failed to list user applications: %w", err)
	}
	return applications, nil
}

func (r *postgresOpportunityApplicationRepository) ListByOpportunityID(ctx context.Context, opportunityID int64, limit, offset int) ([]*model.OpportunityApplication, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, user_id, opportunity_id, message, cv_url, portfolio_url, status, created_at, updated_at
		FROM opportunity_applications 
		WHERE opportunity_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	var applications []*model.OpportunityApplication
	if err := r.db.SelectContext(ctx, &applications, query, opportunityID, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list opportunity applications: %w", err)
	}
	return applications, nil
}

func (r *postgresOpportunityApplicationRepository) Exists(ctx context.Context, userID, opportunityID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM opportunity_applications WHERE user_id = $1 AND opportunity_id = $2)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, userID, opportunityID)
	if err != nil {
		return false, fmt.Errorf("failed to check application existence: %w", err)
	}
	return exists, nil
}
