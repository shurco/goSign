package queries

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AccountQueries handles account-related database operations
type AccountQueries struct {
	pool *pgxpool.Pool
}

// NewAccountQueries creates a new AccountQueries instance
func NewAccountQueries(pool *pgxpool.Pool) *AccountQueries {
	return &AccountQueries{pool: pool}
}

// UpdateAccountLocale updates account locale
func (q *AccountQueries) UpdateAccountLocale(ctx context.Context, accountID, locale string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE account
		SET locale = $1, updated_at = NOW()
		WHERE id = $2
	`, locale, accountID)
	if err != nil {
		return fmt.Errorf("failed to update account locale: %w", err)
	}
	return nil
}

// GetAccountByID retrieves account by ID
func (q *AccountQueries) GetAccountByID(ctx context.Context, accountID string) (*AccountRecord, error) {
	var account AccountRecord
	err := q.pool.QueryRow(ctx, `
		SELECT id, name, timezone, locale
		FROM account
		WHERE id = $1
	`, accountID).Scan(&account.ID, &account.Name, &account.Timezone, &account.Locale)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return &account, nil
}
