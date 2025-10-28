package notification

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shurco/gosign/internal/models"
)

// NotificationRepository interface for database operations
type NotificationRepository interface {
	// GetPending returns notifications pending delivery
	GetPending(ctx context.Context, limit int) ([]*models.Notification, error)
	// UpdateStatus updates notification status
	UpdateStatus(ctx context.Context, id string, status models.NotificationStatus, errorMsg string, sentAt *time.Time) error
	// IncrementRetryCount increments retry counter
	IncrementRetryCount(ctx context.Context, id string) error
}

// Worker processes notification queue
type Worker struct {
	service    *Service
	repository NotificationRepository
	interval   time.Duration
	maxRetries int
}

// NewWorker creates new worker
func NewWorker(service *Service, repository NotificationRepository, interval time.Duration, maxRetries int) *Worker {
	return &Worker{
		service:    service,
		repository: repository,
		interval:   interval,
		maxRetries: maxRetries,
	}
}

// Start starts worker
func (w *Worker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Info().Msg("Notification worker started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Notification worker stopped")
			return nil
		case <-ticker.C:
			if err := w.processQueue(ctx); err != nil {
				log.Error().Err(err).Msg("Failed to process notification queue")
			}
		}
	}
}

// processQueue processes notification queue
func (w *Worker) processQueue(ctx context.Context) error {
	// Get pending notifications ready to send
	notifications, err := w.repository.GetPending(ctx, 100)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		// Check if retry limit exceeded
		if notification.RetryCount >= w.maxRetries {
			now := time.Now()
			if err := w.repository.UpdateStatus(
				ctx,
				notification.ID,
				models.NotificationStatusFailed,
				"max retries exceeded",
				&now,
			); err != nil {
				log.Error().Err(err).Str("notification_id", notification.ID).Msg("Failed to update notification status")
			}
			continue
		}

		// Check if it's time to send
		if notification.ScheduledAt.After(time.Now()) {
			continue
		}

		// Try to send
		if err := w.service.Send(notification); err != nil {
			log.Error().
				Err(err).
				Str("notification_id", notification.ID).
				Str("type", string(notification.Type)).
				Msg("Failed to send notification")

			// Increment retry counter
			if err := w.repository.IncrementRetryCount(ctx, notification.ID); err != nil {
				log.Error().Err(err).Str("notification_id", notification.ID).Msg("Failed to increment retry count")
			}

			// Update status with error
			if err := w.repository.UpdateStatus(
				ctx,
				notification.ID,
				models.NotificationStatusFailed,
				err.Error(),
				nil,
			); err != nil {
				log.Error().Err(err).Str("notification_id", notification.ID).Msg("Failed to update notification status")
			}
			continue
		}

		// Successful delivery
		now := time.Now()
		if err := w.repository.UpdateStatus(
			ctx,
			notification.ID,
			models.NotificationStatusSent,
			"",
			&now,
		); err != nil {
			log.Error().Err(err).Str("notification_id", notification.ID).Msg("Failed to update notification status")
		}

		log.Info().
			Str("notification_id", notification.ID).
			Str("type", string(notification.Type)).
			Str("recipient", notification.Recipient).
			Msg("Notification sent successfully")
	}

	return nil
}

