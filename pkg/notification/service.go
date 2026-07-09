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

// enabledProvider is an optional interface providers can implement to signal
// if they're configured/enabled at runtime.
type enabledProvider interface {
	Enabled() bool
}

// Service manages notification sending
type Service struct {
	providers map[models.NotificationType]Provider
}

// NewService creates a new notification service
func NewService() *Service {
	return &Service{
		providers: make(map[models.NotificationType]Provider),
	}
}

// RegisterProvider registers a notification provider
func (s *Service) RegisterProvider(provider Provider) {
	s.providers[provider.Type()] = provider
}

// Send sends a notification immediately
func (s *Service) Send(notification *models.Notification) error {
	provider, ok := s.providers[notification.Type]
	if !ok {
		return fmt.Errorf("provider for type %s not registered", notification.Type)
	}

	notification.Status = models.NotificationStatusSending
	now := time.Now()
	notification.SentAt = &now

	if err := provider.Send(context.Background(), notification); err != nil {
		notification.Status = models.NotificationStatusFailed
		return fmt.Errorf("failed to send notification: %w", err)
	}

	notification.Status = models.NotificationStatusSent
	return nil
}

// CanSend checks if the service can send notifications of the given type
func (s *Service) CanSend(notificationType models.NotificationType) bool {
	p, ok := s.providers[notificationType]
	if !ok {
		return false
	}
	if ep, ok := p.(enabledProvider); ok {
		return ep.Enabled()
	}
	return true
}
