package worker

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/trust"
	"github.com/shurco/gosign/pkg/notification"
	"github.com/shurco/gosign/pkg/webhook"
)

// NotificationTask task for processing notification queue
type NotificationTask struct {
	worker *notification.Worker
}

// NewNotificationTask creates new task for notifications
func NewNotificationTask(worker *notification.Worker) *NotificationTask {
	return &NotificationTask{
		worker: worker,
	}
}

// Execute performs the task
func (t *NotificationTask) Execute(ctx context.Context) error {
	// Start worker once
	return t.worker.Start(ctx)
}

// ShouldRetry determines if task should be retried on error
func (t *NotificationTask) ShouldRetry(err error) bool {
	// Always retry for notifications
	return true
}

// WebhookTask task for processing webhooks queue
type WebhookTask struct {
	dispatcher *webhook.Dispatcher
	db         *pgxpool.Pool
	accountID  string
}

// NewWebhookTask creates new task for webhooks
func NewWebhookTask(dispatcher *webhook.Dispatcher, db *pgxpool.Pool, accountID string) *WebhookTask {
	return &WebhookTask{
		dispatcher: dispatcher,
		db:         db,
		accountID:  accountID,
	}
}

// Execute performs the task - process pending webhooks
func (t *WebhookTask) Execute(ctx context.Context) error {
	// Get pending webhooks for each event type
	eventTypes := []string{
		"submission.created",
		"submission.completed",
		"submission.cancelled",
		"submission.expired",
		"submitter.completed",
		"submitter.declined",
	}

	processed := 0
	failed := 0

	for _, eventType := range eventTypes {
		webhooks, err := queries.GetPendingWebhooks(ctx, t.db, t.accountID, eventType)
		if err != nil {
			log.Error().Err(err).Str("event_type", eventType).Msg("Failed to get pending webhooks")
			continue
		}

		for _, webhook := range webhooks {
			// Check if webhook should be skipped (too many failures)
			if webhook.FailureCount >= 5 {
				if err := queries.DisableWebhook(ctx, t.db, webhook.ID); err != nil {
					log.Error().Err(err).Str("webhook_id", webhook.ID).Msg("Failed to disable webhook")
				}
				log.Warn().Str("webhook_id", webhook.ID).Msg("Webhook disabled due to too many failures")
				continue
			}

			// Try to send webhook
			// Note: Actual sending should be via dispatcher with proper event data
			// This is a placeholder for the queue processing mechanism
			
			// Update last triggered timestamp
			if err := queries.UpdateWebhookLastTriggered(ctx, t.db, webhook.ID); err != nil {
				log.Error().Err(err).Str("webhook_id", webhook.ID).Msg("Failed to update webhook timestamp")
				failed++
			} else {
				processed++
			}
		}
	}

	log.Info().Int("processed", processed).Int("failed", failed).Msg("Webhook task completed")
	return nil
}

// ShouldRetry determines if task should be retried on error
func (t *WebhookTask) ShouldRetry(err error) bool {
	return true
}

// CleanupTask task for cleaning expired data
type CleanupTask struct {
	db *pgxpool.Pool
}

// NewCleanupTask creates new task for cleanup
func NewCleanupTask(db *pgxpool.Pool) *CleanupTask {
	return &CleanupTask{
		db: db,
	}
}

// Execute performs the task - cleanup old data
func (t *CleanupTask) Execute(ctx context.Context) error {
	deletedCount := 0

	// 1. Delete old expired submissions
	result, err := t.db.Exec(ctx, `
		DELETE FROM submission 
		WHERE status = 'expired' 
		  AND expired_at < NOW() - INTERVAL '30 days'
	`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to cleanup expired submissions")
	} else {
		count := result.RowsAffected()
		deletedCount += int(count)
		log.Info().Int64("count", count).Msg("Cleaned up expired submissions")
	}

	// 2. Delete old events (keep last 90 days)
	result, err = t.db.Exec(ctx, `
		DELETE FROM event 
		WHERE created_at < NOW() - INTERVAL '90 days'
	`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to cleanup old events")
	} else {
		count := result.RowsAffected()
		deletedCount += int(count)
		log.Info().Int64("count", count).Msg("Cleaned up old events")
	}

	// 3. Delete old failed notifications (keep last 7 days)
	result, err = t.db.Exec(ctx, `
		DELETE FROM notification 
		WHERE status = 'failed' 
		  AND created_at < NOW() - INTERVAL '7 days'
	`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to cleanup old notifications")
	} else {
		count := result.RowsAffected()
		deletedCount += int(count)
		log.Info().Int64("count", count).Msg("Cleaned up old failed notifications")
	}

	// 4. Update expired submissions
	result, err = t.db.Exec(ctx, `
		UPDATE submission 
		SET status = 'expired' 
		WHERE expired_at < NOW() 
		  AND status IN ('pending', 'in_progress')
	`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update expired submissions")
	} else {
		count := result.RowsAffected()
		log.Info().Int64("count", count).Msg("Marked submissions as expired")
	}

	log.Info().Int("total_deleted", deletedCount).Msg("Cleanup task completed")
	return nil
}

// ShouldRetry determines if task should be retried on error
func (t *CleanupTask) ShouldRetry(err error) bool {
	return false // cleanup is not critical, no retry
}

// TrustUpdateTask task for updating trust certificates (eutl12, tl12; hardcoded).
type TrustUpdateTask struct{}

// NewTrustUpdateTask creates a new task for updating trust certificates
func NewTrustUpdateTask() *TrustUpdateTask {
	return &TrustUpdateTask{}
}

// Execute performs the task
func (t *TrustUpdateTask) Execute(ctx context.Context) error {
	return trust.Update()
}

// ShouldRetry determines if the task should be retried on error
func (t *TrustUpdateTask) ShouldRetry(err error) bool {
	return true // important task, retry on error
}

