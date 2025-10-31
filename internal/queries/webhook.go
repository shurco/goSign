package queries

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shurco/gosign/internal/models"
)

// GetPendingWebhooks retrieves webhooks that need to be triggered for given event
func GetPendingWebhooks(ctx context.Context, db *pgxpool.Pool, accountID string, eventType string) ([]*models.Webhook, error) {
	query := `
		SELECT id, account_id, name, url, events, enabled, secret, 
		       last_triggered_at, failure_count, created_at, updated_at
		FROM webhook
		WHERE account_id = $1 
		  AND enabled = true
		  AND events @> $2::jsonb
		ORDER BY last_triggered_at ASC NULLS FIRST
	`

	// Convert eventType to JSONB array format
	eventsJSON, err := json.Marshal([]string{eventType})
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, query, accountID, string(eventsJSON))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*models.Webhook
	for rows.Next() {
		webhook := &models.Webhook{}
		var eventsData []byte

		var name string
		err := rows.Scan(
			&webhook.ID,
			&webhook.AccountID,
			&name,
			&webhook.URL,
			&eventsData,
			&webhook.Enabled,
			&webhook.Secret,
			&webhook.LastTriggeredAt,
			&webhook.FailureCount,
			&webhook.CreatedAt,
			&webhook.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse events JSON
		if err := json.Unmarshal(eventsData, &webhook.Events); err != nil {
			return nil, err
		}

		webhooks = append(webhooks, webhook)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return webhooks, nil
}

// UpdateWebhookLastTriggered updates last_triggered_at timestamp
func UpdateWebhookLastTriggered(ctx context.Context, db *pgxpool.Pool, webhookID string) error {
	query := `
		UPDATE webhook
		SET last_triggered_at = NOW()
		WHERE id = $1
	`
	_, err := db.Exec(ctx, query, webhookID)
	return err
}

// IncrementWebhookFailure increments failure_count
func IncrementWebhookFailure(ctx context.Context, db *pgxpool.Pool, webhookID string) error {
	query := `
		UPDATE webhook
		SET failure_count = failure_count + 1
		WHERE id = $1
	`
	_, err := db.Exec(ctx, query, webhookID)
	return err
}

// DisableWebhook disables webhook after too many failures
func DisableWebhook(ctx context.Context, db *pgxpool.Pool, webhookID string) error {
	query := `
		UPDATE webhook
		SET enabled = false
		WHERE id = $1
	`
	_, err := db.Exec(ctx, query, webhookID)
	return err
}

