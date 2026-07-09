package services

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/webhook"
)

const (
	// webhookDisableAfter is the consecutive failure count after which a webhook is disabled.
	webhookDisableAfter = 10
	// webhookDeliveryTimeout bounds a single delivery including dispatcher retries.
	webhookDeliveryTimeout = 2 * time.Minute
)

// WebhookNotifier finds account webhooks subscribed to an event and delivers
// the payload asynchronously via the dispatcher.
type WebhookNotifier struct {
	queries    *queries.WebhookQueries
	dispatcher *webhook.Dispatcher
}

// NewWebhookNotifier creates a webhook notifier.
func NewWebhookNotifier(q *queries.WebhookQueries, d *webhook.Dispatcher) *WebhookNotifier {
	return &WebhookNotifier{queries: q, dispatcher: d}
}

// SendEvent delivers the event to all enabled account webhooks subscribed to
// event.Type. Delivery runs in background goroutines and never blocks the caller.
func (n *WebhookNotifier) SendEvent(ctx context.Context, accountID string, event *models.WebhookEvent) {
	hooks, err := n.queries.EnabledWebhooksForEvent(ctx, accountID, event.Type)
	if err != nil {
		log.Error().Err(err).Str("account_id", accountID).Str("event", event.Type).Msg("Failed to load webhooks for event")
		return
	}

	for i := range hooks {
		hook := hooks[i]
		go n.deliver(&hook, event)
	}
}

// deliver sends the event to a single webhook and records the outcome.
// Uses a fresh context: the originating request context ends before retries do.
func (n *WebhookNotifier) deliver(hook *models.Webhook, event *models.WebhookEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), webhookDeliveryTimeout)
	defer cancel()

	if err := n.dispatcher.Send(ctx, hook, event); err != nil {
		log.Warn().Err(err).Str("webhook_id", hook.ID).Str("event", event.Type).Msg("Webhook delivery failed")
		if err := n.queries.MarkFailed(ctx, hook.ID, webhookDisableAfter); err != nil {
			log.Error().Err(err).Str("webhook_id", hook.ID).Msg("Failed to record webhook failure")
		}
		return
	}

	if err := n.queries.MarkTriggered(ctx, hook.ID); err != nil {
		log.Error().Err(err).Str("webhook_id", hook.ID).Msg("Failed to record webhook success")
	}
}
