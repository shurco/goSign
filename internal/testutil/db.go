package testutil

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/goosemigrator"

	"github.com/shurco/gosign/migrations"
)

const (
	testDBHost    = "localhost"
	testDBPort    = "5453"
	testDBUser    = "postgres"
	testDBPass    = "password"
	testDBOptions = "sslmode=disable"
)

type chainedMigrator struct {
	migrators []pgtestdb.Migrator
}

func (m chainedMigrator) Hash() (string, error) {
	sum := sha256.New()
	for _, migrator := range m.migrators {
		hash, err := migrator.Hash()
		if err != nil {
			return "", err
		}
		if _, err := io.WriteString(sum, hash); err != nil {
			return "", err
		}
		if _, err := io.WriteString(sum, "|"); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

func (m chainedMigrator) Migrate(ctx context.Context, db *sql.DB, conf pgtestdb.Config) error {
	for _, migrator := range m.migrators {
		if err := migrator.Migrate(ctx, db, conf); err != nil {
			return err
		}
	}
	return nil
}

// NewTestDB creates isolated pgxpool backed by pgtestdb clone DB.
// It skips tests gracefully when test postgres is unavailable.
func NewTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if !isTestPostgresAvailable() {
		t.Skipf("skip: pgtestdb unavailable at %s:%s", testDBHost, testDBPort)
	}

	repoRoot := mustRepoRoot(t)
	dbConf := pgtestdb.Config{
		DriverName: "pgx",
		Host:       testDBHost,
		Port:       testDBPort,
		User:       testDBUser,
		Password:   testDBPass,
		Database:   "postgres",
		Options:    testDBOptions,
	}

	migrator := chainedMigrator{
		migrators: []pgtestdb.Migrator{
			goosemigrator.New(".", goosemigrator.WithFS(migrations.Embed())),
			goosemigrator.New(
				"fixtures/migration",
				goosemigrator.WithFS(os.DirFS(repoRoot)),
				// Keep fixture migration history separate to allow older fixture versions
				// after applying the main schema migrations.
				goosemigrator.WithTableName("goose_db_version_fixtures"),
			),
		},
	}

	sqlDB := pgtestdb.New(t, dbConf, migrator)

	nameCtx, nameCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer nameCancel()

	var cloneDBName string
	if err := sqlDB.QueryRowContext(nameCtx, "SELECT current_database()").Scan(&cloneDBName); err != nil {
		t.Fatalf("get clone db name: %v", err)
	}

	poolDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?%s",
		pgtestdb.DefaultRoleUsername,
		pgtestdb.DefaultRolePassword,
		testDBHost,
		testDBPort,
		cloneDBName,
		testDBOptions,
	)

	poolCtx, poolCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer poolCancel()

	pool, err := pgxpool.New(poolCtx, poolDSN)
	if err != nil {
		t.Fatalf("create pgxpool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func isTestPostgresAvailable() bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(testDBHost, testDBPort), 700*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func mustRepoRoot(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve caller path")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	if _, err := fs.Stat(os.DirFS(root), "fixtures/migration"); err != nil {
		t.Fatalf("resolve repo root from %q: %v", root, err)
	}
	return root
}
