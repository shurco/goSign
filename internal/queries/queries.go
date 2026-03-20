package queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/shurco/gosign/migrations"
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
func New(pool *pgxpool.Pool) error {
	DB = &Base{
		SystemQueries:        SystemQueries{pool},
		TrustQueries:         TrustQueries{pool},
		TemplateQueries:      TemplateQueries{pool},
		UserQueries:          UserQueries{pool},
		EmailTemplateQueries: EmailTemplateQueries{pool},
	}

	return nil
}

// Init initializes the database connection and runs pending migrations.
func Init(pool *pgxpool.Pool) error {
	New(pool)

	var exists bool
	err := pool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'migrate_db_version')`,
	).Scan(&exists)
	if err != nil || !exists {
		goose.SetBaseFS(migrations.Embed())
		goose.SetTableName("migrate_db_version")
		if err := goose.SetDialect("pgx"); err != nil {
			return err
		}

		db := stdlib.OpenDBFromPool(pool)
		if err := goose.Up(db, "."); err != nil {
			return err
		}
	}

	return nil
}
