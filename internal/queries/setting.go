package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SettingQueries handles global settings database operations
type SettingQueries struct {
	pool *pgxpool.Pool
}

// NewSettingQueries creates a new SettingQueries instance
func NewSettingQueries(pool *pgxpool.Pool) *SettingQueries {
	return &SettingQueries{pool: pool}
}

// GetGlobalSetting retrieves a global setting by key
func (q *SettingQueries) GetGlobalSetting(ctx context.Context, key string) (map[string]any, error) {
	var valueJSON []byte
	err := q.pool.QueryRow(ctx, `
		SELECT COALESCE(value, '{}'::jsonb)
		FROM setting
		WHERE key = $1
	`, key).Scan(&valueJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get global setting %s: %w", key, err)
	}

	var value map[string]any
	if err := json.Unmarshal(valueJSON, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal setting value: %w", err)
	}

	return value, nil
}

// GetAllGlobalSettings retrieves all global settings grouped by category
func (q *SettingQueries) GetAllGlobalSettings(ctx context.Context) (map[string]map[string]any, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT key, value, category
		FROM setting
		ORDER BY category, key
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get global settings: %w", err)
	}
	defer rows.Close()

	result := make(map[string]map[string]any)
	for rows.Next() {
		var key, category string
		var valueJSON []byte

		if err := rows.Scan(&key, &valueJSON, &category); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}

		var value map[string]any
		if err := json.Unmarshal(valueJSON, &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal setting value: %w", err)
		}

		result[key] = value
	}

	return result, nil
}

// GetGlobalSettingsByCategory retrieves all global settings for a specific category
func (q *SettingQueries) GetGlobalSettingsByCategory(ctx context.Context, category string) (map[string]map[string]any, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT key, value
		FROM setting
		WHERE category = $1
		ORDER BY key
	`, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get global settings by category: %w", err)
	}
	defer rows.Close()

	result := make(map[string]map[string]any)
	for rows.Next() {
		var key string
		var valueJSON []byte

		if err := rows.Scan(&key, &valueJSON); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}

		var value map[string]any
		if err := json.Unmarshal(valueJSON, &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal setting value: %w", err)
		}

		result[key] = value
	}

	return result, nil
}

// UpdateGlobalSetting updates or creates a global setting
func (q *SettingQueries) UpdateGlobalSetting(ctx context.Context, key string, value map[string]any, category string) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal setting value: %w", err)
	}

	if category == "" {
		category = "general"
	}

	_, err = q.pool.Exec(ctx, `
		INSERT INTO setting (key, value, category, updated_at)
		VALUES ($1, $2::jsonb, $3, NOW())
		ON CONFLICT (key) 
		DO UPDATE SET 
			value = EXCLUDED.value,
			category = EXCLUDED.category,
			updated_at = NOW()
	`, key, valueJSON, category)
	if err != nil {
		return fmt.Errorf("failed to update global setting: %w", err)
	}

	return nil
}

// UpdateGlobalSettingPartial updates only specific fields in a global setting
func (q *SettingQueries) UpdateGlobalSettingPartial(ctx context.Context, key string, updates map[string]any) error {
	// Get current value
	currentValue, err := q.GetGlobalSetting(ctx, key)
	if err != nil {
		// If setting doesn't exist, create it with the updates
		return q.UpdateGlobalSetting(ctx, key, updates, "general")
	}

	// Merge updates into current value
	for k, v := range updates {
		currentValue[k] = v
	}

	// Determine category from key if not set
	category := "general"
	switch key {
	case "smtp":
		category = "email"
	case "sms":
		category = "sms"
	case "storage":
		category = "storage"
	case "geolocation":
		category = "geolocation"
	}

	return q.UpdateGlobalSetting(ctx, key, currentValue, category)
}

// DeleteGlobalSetting deletes a global setting
func (q *SettingQueries) DeleteGlobalSetting(ctx context.Context, key string) error {
	_, err := q.pool.Exec(ctx, `
		DELETE FROM setting
		WHERE key = $1
	`, key)
	if err != nil {
		return fmt.Errorf("failed to delete global setting: %w", err)
	}

	return nil
}
