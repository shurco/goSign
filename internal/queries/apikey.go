package queries

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/models"
)

// APIKeyRepository implements API key storage operations
type APIKeyRepository struct {
	pool *pgxpool.Pool
}

// NewAPIKeyRepository creates new API key repository
func NewAPIKeyRepository(pool *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{pool: pool}
}

// GetByKeyHash retrieves API key by hash; returns nil, nil when not found.
func (r *APIKeyRepository) GetByKeyHash(keyHash string) (*models.APIKey, error) {
	const query = `
		SELECT id, name, key_hash, account_id, enabled, last_used_at, expires_at, created_at, updated_at
		FROM api_key
		WHERE key_hash = $1
	`
	var apiKey models.APIKey
	err := r.pool.QueryRow(context.Background(), query, keyHash).Scan(
		&apiKey.ID,
		&apiKey.Name,
		&apiKey.KeyHash,
		&apiKey.AccountID,
		&apiKey.Enabled,
		&apiKey.LastUsedAt,
		&apiKey.ExpiresAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// UpdateLastUsed updates last used timestamp; returns pgx.ErrNoRows if key not found.
func (r *APIKeyRepository) UpdateLastUsed(keyID string, lastUsed time.Time) error {
	const query = `UPDATE api_key SET last_used_at = $1 WHERE id = $2`
	tag, err := r.pool.Exec(context.Background(), query, lastUsed, keyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// Create inserts a new API key and populates its ID from RETURNING.
func (r *APIKeyRepository) Create(apiKey *models.APIKey) error {
	const query = `
		INSERT INTO api_key (name, key_hash, account_id, enabled, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	return r.pool.QueryRow(
		context.Background(),
		query,
		apiKey.Name,
		apiKey.KeyHash,
		apiKey.AccountID,
		apiKey.Enabled,
		apiKey.ExpiresAt,
		apiKey.CreatedAt,
		apiKey.UpdatedAt,
	).Scan(&apiKey.ID)
}

// Update saves changes to an existing API key; returns pgx.ErrNoRows if key not found.
func (r *APIKeyRepository) Update(apiKey *models.APIKey) error {
	const query = `
		UPDATE api_key
		SET name = $1, enabled = $2, expires_at = $3, updated_at = $4
		WHERE id = $5
	`
	tag, err := r.pool.Exec(
		context.Background(),
		query,
		apiKey.Name,
		apiKey.Enabled,
		apiKey.ExpiresAt,
		apiKey.UpdatedAt,
		apiKey.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// Delete removes an API key by ID; returns pgx.ErrNoRows if key not found.
func (r *APIKeyRepository) Delete(keyID string) error {
	const query = `DELETE FROM api_key WHERE id = $1`
	tag, err := r.pool.Exec(context.Background(), query, keyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListByAccount returns all API keys for an account ordered by creation date.
func (r *APIKeyRepository) ListByAccount(accountID string) ([]*models.APIKey, error) {
	const query = `
		SELECT id, name, key_hash, account_id, enabled, last_used_at, expires_at, created_at, updated_at
		FROM api_key
		WHERE account_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(context.Background(), query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.APIKey
	for rows.Next() {
		var key models.APIKey
		if err := rows.Scan(
			&key.ID,
			&key.Name,
			&key.KeyHash,
			&key.AccountID,
			&key.Enabled,
			&key.LastUsedAt,
			&key.ExpiresAt,
			&key.CreatedAt,
			&key.UpdatedAt,
		); err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}
	return keys, rows.Err()
}
