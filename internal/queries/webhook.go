package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/models"
)

const webhookColumns = `id, account_id, url, events, secret, enabled, last_triggered_at, failure_count, created_at, updated_at`

// WebhookQueries provides account-scoped webhook persistence.
type WebhookQueries struct {
	pool *pgxpool.Pool
}

// NewWebhookQueries creates webhook queries.
func NewWebhookQueries(pool *pgxpool.Pool) *WebhookQueries {
	return &WebhookQueries{pool: pool}
}

func scanWebhook(row pgx.Row) (*models.Webhook, error) {
	var (
		webhook   models.Webhook
		eventsRaw []byte
	)
	if err := row.Scan(
		&webhook.ID,
		&webhook.AccountID,
		&webhook.URL,
		&eventsRaw,
		&webhook.Secret,
		&webhook.Enabled,
		&webhook.LastTriggeredAt,
		&webhook.FailureCount,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if len(eventsRaw) > 0 {
		if err := json.Unmarshal(eventsRaw, &webhook.Events); err != nil {
			return nil, fmt.Errorf("unmarshal webhook events: %w", err)
		}
	}
	return &webhook, nil
}

func collectWebhooks(rows pgx.Rows) ([]models.Webhook, error) {
	defer rows.Close()

	var webhooks []models.Webhook
	for rows.Next() {
		webhook, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, *webhook)
	}
	return webhooks, rows.Err()
}

// ListWebhooks returns all webhooks of the account.
func (q *WebhookQueries) ListWebhooks(ctx context.Context, accountID string) ([]models.Webhook, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT `+webhookColumns+`
		FROM webhook
		WHERE account_id = $1
		ORDER BY created_at DESC
	`, accountID)
	if err != nil {
		return nil, fmt.Errorf("queries: list webhooks: %w", err)
	}
	return collectWebhooks(rows)
}

// GetWebhook returns a webhook by id within the account. Returns pgx.ErrNoRows when missing.
func (q *WebhookQueries) GetWebhook(ctx context.Context, accountID, id string) (*models.Webhook, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT `+webhookColumns+`
		FROM webhook
		WHERE id = $1 AND account_id = $2
	`, id, accountID)
	return scanWebhook(row)
}

// CreateWebhook inserts a webhook and fills generated fields on the passed model.
func (q *WebhookQueries) CreateWebhook(ctx context.Context, webhook *models.Webhook) error {
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return fmt.Errorf("queries: create webhook: %w", err)
	}

	row := q.pool.QueryRow(ctx, `
		INSERT INTO webhook (account_id, url, events, secret, enabled)
		VALUES ($1, $2, $3::jsonb, $4, $5)
		RETURNING id, created_at, updated_at
	`, webhook.AccountID, webhook.URL, string(eventsJSON), webhook.Secret, webhook.Enabled)
	if err := row.Scan(&webhook.ID, &webhook.CreatedAt, &webhook.UpdatedAt); err != nil {
		return fmt.Errorf("queries: create webhook: %w", err)
	}
	return nil
}

// UpdateWebhook updates a webhook within the account. Returns pgx.ErrNoRows when missing.
func (q *WebhookQueries) UpdateWebhook(ctx context.Context, webhook *models.Webhook) error {
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return fmt.Errorf("queries: update webhook: %w", err)
	}

	tag, err := q.pool.Exec(ctx, `
		UPDATE webhook
		SET url = $3, events = $4::jsonb, secret = $5, enabled = $6, updated_at = NOW()
		WHERE id = $1 AND account_id = $2
	`, webhook.ID, webhook.AccountID, webhook.URL, string(eventsJSON), webhook.Secret, webhook.Enabled)
	if err != nil {
		return fmt.Errorf("queries: update webhook: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteWebhook removes a webhook within the account. Returns pgx.ErrNoRows when missing.
func (q *WebhookQueries) DeleteWebhook(ctx context.Context, accountID, id string) error {
	tag, err := q.pool.Exec(ctx, `DELETE FROM webhook WHERE id = $1 AND account_id = $2`, id, accountID)
	if err != nil {
		return fmt.Errorf("queries: delete webhook: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// EnabledWebhooksForEvent returns enabled account webhooks subscribed to the event type.
func (q *WebhookQueries) EnabledWebhooksForEvent(ctx context.Context, accountID, eventType string) ([]models.Webhook, error) {
	eventsJSON, err := json.Marshal([]string{eventType})
	if err != nil {
		return nil, fmt.Errorf("queries: webhooks for event: %w", err)
	}

	rows, err := q.pool.Query(ctx, `
		SELECT `+webhookColumns+`
		FROM webhook
		WHERE account_id = $1
		  AND enabled = true
		  AND (events @> $2::jsonb OR events @> '["*"]'::jsonb)
		ORDER BY last_triggered_at ASC NULLS FIRST
	`, accountID, string(eventsJSON))
	if err != nil {
		return nil, fmt.Errorf("queries: webhooks for event: %w", err)
	}
	return collectWebhooks(rows)
}

// MarkTriggered records a successful delivery and resets the failure counter.
func (q *WebhookQueries) MarkTriggered(ctx context.Context, webhookID string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE webhook
		SET last_triggered_at = NOW(), failure_count = 0
		WHERE id = $1
	`, webhookID)
	return err
}

// MarkFailed increments the failure counter and disables the webhook once
// disableAfter consecutive failures are reached.
func (q *WebhookQueries) MarkFailed(ctx context.Context, webhookID string, disableAfter int) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE webhook
		SET failure_count = failure_count + 1,
		    enabled = CASE WHEN failure_count + 1 >= $2 THEN false ELSE enabled END
		WHERE id = $1
	`, webhookID, disableAfter)
	return err
}
