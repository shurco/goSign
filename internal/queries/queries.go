package queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/shurco/gosign/migrations"
)

var DB *Base

// Base is ...
type Base struct {
	SystemQueries
	TrustQueries
	TemplateQueries
	UserQueries
	EmailTemplateQueries
}

// New is ...
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

// Init is ...
func Init(pool *pgxpool.Pool) error {
	New(pool)

	if _, err := pool.Query(context.Background(), `SELECT * FROM migrate_db_version`); err != nil {
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

/*
// SQLPagination is ...
// example query's for sortBy - id:DESC or id:ASC
func (db *Base) SQLPagination(params webutil.PaginationQuery) string {
	if params.Offset < 0 {
		params.Offset = 0
	}

	if params.Limit <= 0 {
		params.Limit = 30
	}

	var showSortBy string
	if len(params.SortBy) > 0 {
		showSortBy = "ORDER BY "

		var orderParts []string
		sorts := strings.Split(params.SortBy, ",")
		for _, sort := range sorts {
			parts := strings.SplitN(sort, ":", 2)
			if len(parts) == 1 {
				orderParts = append(orderParts, parts[0])
			}
			if len(parts) == 2 {
				orderParts = append(orderParts, fmt.Sprintf("%s %s", parts[0], parts[1]))
			}
		}
		showSortBy = showSortBy + strings.Join(orderParts, ", ")
	}

	return fmt.Sprintf(" %s LIMIT %d OFFSET %d", showSortBy, params.Limit, params.Offset)
}
*/
