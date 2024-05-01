package postgres

import (
	"context"
	"database/sql"
	"fmt"

	// Load jackc package

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var Client *sql.DB

// Config is ...
type Config struct {
	URL string `toml:"url"`
}

// New creates a new Connect object using the given PgSQLConfig.
func New(ctx context.Context, conf Config) (*pgxpool.Pool, error) {
	// ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()

	config, _ := pgxpool.ParseConfig(conf.URL)
	// config.MaxConnLifetime = time.Duration(conf.MaxLifetimeConn) * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	// defer pool.Close()

	// Ping the database to ensure connectivity.
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("could not ping database: %v", err)
	}

	return pool, nil
}
