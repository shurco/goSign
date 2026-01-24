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

// UpdateAccountGeolocationSettings updates geolocation settings in account.settings jsonb
func (q *AccountQueries) UpdateAccountGeolocationSettings(ctx context.Context, accountID string, maxmindLicenseKey, downloadURL, downloadMethod string) error {
	// Build jsonb object dynamically - only include non-empty fields
	updates := []string{}
	args := []any{accountID}
	argIndex := 2

	if maxmindLicenseKey != "" {
		// Cast to text so Postgres can infer parameter type inside jsonb_build_object
		updates = append(updates, fmt.Sprintf("jsonb_build_object('maxmind_license_key', $%d::text)", argIndex))
		args = append(args, maxmindLicenseKey)
		argIndex++
	}
	if downloadURL != "" {
		updates = append(updates, fmt.Sprintf("jsonb_build_object('download_url', $%d::text)", argIndex))
		args = append(args, downloadURL)
		argIndex++
	}
	if downloadMethod != "" {
		updates = append(updates, fmt.Sprintf("jsonb_build_object('download_method', $%d::text)", argIndex))
		args = append(args, downloadMethod)
		argIndex++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no settings to update")
	}

	// Combine all updates
	combinedUpdate := strings.Join(updates, " || ")

	query := fmt.Sprintf(`
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
				) || (%s)
			),
			updated_at = NOW()
		FROM base
		WHERE account.id = $1
	`, combinedUpdate)

	_, err := q.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update account geolocation settings: %w", err)
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

// UpdateAccountGeolocationLastUpdatedAt is kept for backward compatibility.
func (q *AccountQueries) UpdateAccountGeolocationLastUpdatedAt(ctx context.Context, accountID string, updatedAt time.Time) error {
	return q.UpdateAccountGeolocationLastUpdate(ctx, accountID, updatedAt, "unknown")
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

// GetAccountGeolocationSettings retrieves all geolocation settings from account
func (q *AccountQueries) GetAccountGeolocationSettings(ctx context.Context, accountID string) (map[string]any, error) {
	var settingsJSON []byte
	err := q.pool.QueryRow(ctx, `
		SELECT COALESCE(settings->'geolocation', '{}'::jsonb)
		FROM account
		WHERE id = $1
	`, accountID).Scan(&settingsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get geolocation settings: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(settingsJSON, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal geolocation settings: %w", err)
	}

	return settings, nil
}

// GetAllAccountsGeolocationLicenseKeys retrieves all MaxMind license keys from all accounts
// Returns the first non-empty license key found
func (q *AccountQueries) GetAllAccountsGeolocationLicenseKeys(ctx context.Context) (string, error) {
	var licenseKey string
	err := q.pool.QueryRow(ctx, `
		SELECT COALESCE(settings->'geolocation'->>'maxmind_license_key', '')
		FROM account
		WHERE settings->'geolocation'->>'maxmind_license_key' IS NOT NULL
		  AND settings->'geolocation'->>'maxmind_license_key' != ''
		LIMIT 1
	`).Scan(&licenseKey)
	if err != nil {
		// No accounts with license key configured
		return "", nil
	}
	return licenseKey, nil
}
