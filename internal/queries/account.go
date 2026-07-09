package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

// GetAccountSettings retrieves account settings from database
func (q *AccountQueries) GetAccountSettings(ctx context.Context, accountID string) (map[string]any, error) {
	var settingsJSON []byte
	err := q.pool.QueryRow(ctx, `
		SELECT COALESCE(settings, '{}'::jsonb)
		FROM account
		WHERE id = $1
	`, accountID).Scan(&settingsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get account settings: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(settingsJSON, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account settings: %w", err)
	}

	return settings, nil
}

// UpdateAccountSettings updates account settings in database
func (q *AccountQueries) UpdateAccountSettings(ctx context.Context, accountID string, settings map[string]any) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal account settings: %w", err)
	}

	_, err = q.pool.Exec(ctx, `
		UPDATE account
		SET settings = $1::jsonb, updated_at = NOW()
		WHERE id = $2
	`, settingsJSON, accountID)
	if err != nil {
		return fmt.Errorf("failed to update account settings: %w", err)
	}

	return nil
}

// DeleteAccountGeolocationMaxMindLicenseKey removes geolocation.maxmind_license_key from account settings.
func (q *AccountQueries) DeleteAccountGeolocationMaxMindLicenseKey(ctx context.Context, accountID string) error {
	query := `
		WITH base AS (
			SELECT
				CASE
					WHEN jsonb_typeof(settings) = 'object' THEN settings
					ELSE '{}'::jsonb
				END AS settings_obj
			FROM account
			WHERE id = $1
		)
		UPDATE account
		SET settings = base.settings_obj ||
			jsonb_build_object(
				'geolocation',
				(
					CASE
						WHEN jsonb_typeof(base.settings_obj->'geolocation') = 'object' THEN base.settings_obj->'geolocation'
						ELSE '{}'::jsonb
					END
				) - 'maxmind_license_key'
			),
			updated_at = NOW()
		FROM base
		WHERE account.id = $1
	`

	if _, err := q.pool.Exec(ctx, query, accountID); err != nil {
		return fmt.Errorf("failed to delete geolocation maxmind_license_key: %w", err)
	}
	return nil
}

// UpdateAccountGeolocationLastUpdate stores the last successful GeoLite2 database update info
// in account.settings->geolocation:
// - last_updated_at (RFC3339 UTC string)
// - last_updated_source ("maxmind" or "url")
func (q *AccountQueries) UpdateAccountGeolocationLastUpdate(ctx context.Context, accountID string, updatedAt time.Time, source string) error {
	ts := updatedAt.UTC().Format(time.RFC3339)
	source = strings.TrimSpace(source)
	if source == "" {
		source = "unknown"
	}

	query := `
		WITH base AS (
			SELECT
				CASE
					WHEN jsonb_typeof(settings) = 'object' THEN settings
					ELSE '{}'::jsonb
				END AS settings_obj
			FROM account
			WHERE id = $1
		)
		UPDATE account
		SET settings = base.settings_obj ||
			jsonb_build_object(
				'geolocation',
				(
					CASE
						WHEN jsonb_typeof(base.settings_obj->'geolocation') = 'object' THEN base.settings_obj->'geolocation'
						ELSE '{}'::jsonb
					END
				) || jsonb_build_object('last_updated_at', $2::text, 'last_updated_source', $3::text)
			),
			updated_at = NOW()
		FROM base
		WHERE account.id = $1
	`

	if _, err := q.pool.Exec(ctx, query, accountID, ts, source); err != nil {
		return fmt.Errorf("failed to update geolocation last update: %w", err)
	}
	return nil
}

// GetAccountGeolocationLicenseKey retrieves MaxMind license key from account settings
func (q *AccountQueries) GetAccountGeolocationLicenseKey(ctx context.Context, accountID string) (string, error) {
	var licenseKey string
	err := q.pool.QueryRow(ctx, `
		SELECT COALESCE(settings->'geolocation'->>'maxmind_license_key', '')
		FROM account
		WHERE id = $1
	`, accountID).Scan(&licenseKey)
	if err != nil {
		return "", fmt.Errorf("failed to get geolocation license key: %w", err)
	}
	return licenseKey, nil
}
