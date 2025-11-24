package db

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Postgres wraps database connection
type Postgres struct {
	db *sqlx.DB
}

// NewPostgres creates a new PostgreSQL connection
func NewPostgres(dsn string) (*Postgres, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &Postgres{db: db}, nil
}

// GetDB returns the underlying *sqlx.DB
func (p *Postgres) GetDB() *sqlx.DB {
	return p.db
}

// GetSQLDB returns the underlying *sql.DB
func (p *Postgres) GetSQLDB() *sql.DB {
	return p.db.DB
}

// Close closes the database connection
func (p *Postgres) Close() error {
	return p.db.Close()
}

// HealthCheck checks if the database is healthy
func (p *Postgres) HealthCheck() error {
	return p.db.Ping()
}
