package queries

import (
	"database/sql"
	"time"

	"github.com/shurco/gosign/internal/models"
)

// APIKeyRepository implements API key storage operations
type APIKeyRepository struct {
	db *sql.DB
}

// NewAPIKeyRepository creates new API key repository
func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// GetByKeyHash retrieves API key by hash
func (r *APIKeyRepository) GetByKeyHash(keyHash string) (*models.APIKey, error) {
	var apiKey models.APIKey
	query := `
		SELECT id, name, key_hash, account_id, enabled, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE key_hash = $1
	`
	err := r.db.QueryRow(query, keyHash).Scan(
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
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// UpdateLastUsed updates last used timestamp
func (r *APIKeyRepository) UpdateLastUsed(keyID string, lastUsed time.Time) error {
	query := `UPDATE api_keys SET last_used_at = $1 WHERE id = $2`
	_, err := r.db.Exec(query, lastUsed, keyID)
	return err
}

// Create creates new API key
func (r *APIKeyRepository) Create(apiKey *models.APIKey) error {
	query := `
		INSERT INTO api_keys (name, key_hash, account_id, enabled, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	return r.db.QueryRow(
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

// Update updates API key
func (r *APIKeyRepository) Update(apiKey *models.APIKey) error {
	query := `
		UPDATE api_keys
		SET name = $1, enabled = $2, expires_at = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(
		query,
		apiKey.Name,
		apiKey.Enabled,
		apiKey.ExpiresAt,
		apiKey.UpdatedAt,
		apiKey.ID,
	)
	return err
}

// Delete deletes API key
func (r *APIKeyRepository) Delete(keyID string) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := r.db.Exec(query, keyID)
	return err
}

// ListByAccount lists all API keys for account
func (r *APIKeyRepository) ListByAccount(accountID string) ([]*models.APIKey, error) {
	query := `
		SELECT id, name, key_hash, account_id, enabled, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE account_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.APIKey
	for rows.Next() {
		var key models.APIKey
		err := rows.Scan(
			&key.ID,
			&key.Name,
			&key.KeyHash,
			&key.AccountID,
			&key.Enabled,
			&key.LastUsedAt,
			&key.ExpiresAt,
			&key.CreatedAt,
			&key.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}

	return keys, rows.Err()
}

