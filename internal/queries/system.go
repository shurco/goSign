package queries

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// SystemQueries is ...
type SystemQueries struct {
	*pgxpool.Pool
}
