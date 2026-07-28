package queries

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *Base

// Base aggregates all query structs for convenient access.
type Base struct {
	SystemQueries
	TrustQueries
	TemplateQueries
	UserQueries
	EmailTemplateQueries
}

// New creates the global DB instance with all query structs.
func New(pool *pgxpool.Pool) {
	DB = &Base{
		SystemQueries:        SystemQueries{pool},
		TrustQueries:         TrustQueries{pool},
		TemplateQueries:      TemplateQueries{pool},
		UserQueries:          UserQueries{pool},
		EmailTemplateQueries: EmailTemplateQueries{pool},
	}
}

// CheckSchema verifies that database migrations have been applied.
// Migrations are managed externally: `./scripts/migration up` locally
// or the `migrate` compose service in Docker.
func CheckSchema(ctx context.Context, pool *pgxpool.Pool) error {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'migrate_db_version')`,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("queries: check schema: %w", err)
	}
	if !exists {
		return fmt.Errorf("queries: database schema is not initialized, apply migrations first")
	}
	return nil
}
