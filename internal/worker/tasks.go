package worker

import (
	"context"

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
	// TODO: add repository for getting pending webhooks
}

// NewWebhookTask creates new task for webhooks
func NewWebhookTask(dispatcher *webhook.Dispatcher) *WebhookTask {
	return &WebhookTask{
		dispatcher: dispatcher,
	}
}

// Execute performs the task
func (t *WebhookTask) Execute(ctx context.Context) error {
	// TODO: implement getting pending webhooks from DB and sending
	// For now placeholder
	return nil
}

// ShouldRetry determines if task should be retried on error
func (t *WebhookTask) ShouldRetry(err error) bool {
	return true
}

// CleanupTask task for cleaning expired data
type CleanupTask struct {
	// TODO: add repository for DB operations
}

// NewCleanupTask creates new task for cleanup
func NewCleanupTask() *CleanupTask {
	return &CleanupTask{}
}

// Execute performs the task
func (t *CleanupTask) Execute(ctx context.Context) error {
	// TODO: implement cleanup of expired submissions, old events, etc.
	return nil
}

// ShouldRetry determines if task should be retried on error
func (t *CleanupTask) ShouldRetry(err error) bool {
	return false // cleanup is not critical, no retry
}

// TrustUpdateTask task for updating trust certificates
type TrustUpdateTask struct {
	trustConfig trust.Config
}

// NewTrustUpdateTask creates a new task for updating trust certificates
func NewTrustUpdateTask(trustConfig trust.Config) *TrustUpdateTask {
	return &TrustUpdateTask{
		trustConfig: trustConfig,
	}
}

// Execute performs the task
func (t *TrustUpdateTask) Execute(ctx context.Context) error {
	return trust.Update(t.trustConfig)
}

// ShouldRetry determines if the task should be retried on error
func (t *TrustUpdateTask) ShouldRetry(err error) bool {
	return true // important task, retry on error
}

