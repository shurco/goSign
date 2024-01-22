package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Load jackc package
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var Client *sql.DB

// PgSQLConfig is ...
type PgSQLConfig struct {
	DSN             string
	MaxConn         int
	MaxIdleConn     int
	MaxLifetimeConn int
}

var Pool *pgxpool.Pool

// NewClient creates a new Connect object using the given PgSQLConfig.
func NewClient(ctx context.Context, conf *PgSQLConfig) error {
	p, err := pgxpool.New(ctx, conf.DSN)
	if err != nil {
		return err
	}

	// Ping the database to ensure connectivity.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("could not ping database: %v", err)
	}

	Pool = p

	// Return the new Connect object.
	return nil
}
