package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/shurco/gosign/internal/models"
)

// Provider represents a notification provider interface
type Provider interface {
	// Send sends a notification
	Send(ctx context.Context, notification *models.Notification) error
	// Type returns the provider type
	Type() models.NotificationType
}

// Repository represents an interface for working with notification database
type Repository interface {
	Create(notification *models.Notification) error
	GetScheduledReady() ([]*models.Notification, error)
	UpdateStatus(id string, status models.NotificationStatus) error
	CancelByRelatedID(relatedID string) error
}

// Service manages notification sending
type Service struct {
	providers  map[models.NotificationType]Provider
	repository Repository
}

// NewService creates a new notification service
func NewService(repo Repository) *Service {
	return &Service{
		providers:  make(map[models.NotificationType]Provider),
		repository: repo,
	}
}

// RegisterProvider registers a notification provider
func (s *Service) RegisterProvider(provider Provider) {
	s.providers[provider.Type()] = provider
}

// Send sends a notification immediately
func (s *Service) Send(notification *models.Notification) error {
	ctx := context.Background()
	
	provider, ok := s.providers[notification.Type]
	if !ok {
		return fmt.Errorf("provider for type %s not registered", notification.Type)
	}

	// Update status to sending
	notification.Status = models.NotificationStatusSending
	now := time.Now()
	notification.SentAt = &now

	// Send notification
	if err := provider.Send(ctx, notification); err != nil {
		notification.Status = models.NotificationStatusFailed
		if s.repository != nil {
			_ = s.repository.UpdateStatus(notification.ID, models.NotificationStatusFailed)
		}
		return fmt.Errorf("failed to send notification: %w", err)
	}

	// Update status to sent
	notification.Status = models.NotificationStatusSent
	if s.repository != nil {
		_ = s.repository.UpdateStatus(notification.ID, models.NotificationStatusSent)
	}

	return nil
}

// Schedule schedules a notification for future sending
func (s *Service) Schedule(notification *models.Notification) error {
	if s.repository == nil {
		return fmt.Errorf("repository not configured")
	}

	notification.Status = models.NotificationStatusPending
	return s.repository.Create(notification)
}

// GetScheduledReady retrieves all scheduled notifications ready to be sent
func (s *Service) GetScheduledReady() ([]*models.Notification, error) {
	if s.repository == nil {
		return nil, fmt.Errorf("repository not configured")
	}

	return s.repository.GetScheduledReady()
}

// CancelScheduled cancels all scheduled notifications for submission/submitter
func (s *Service) CancelScheduled(relatedID string) error {
	if s.repository == nil {
		return nil // No repository - nothing to cancel
	}

	return s.repository.CancelByRelatedID(relatedID)
}

// CanSend checks if the service can send notifications of the given type
func (s *Service) CanSend(notificationType models.NotificationType) bool {
	_, ok := s.providers[notificationType]
	return ok
}

