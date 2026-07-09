package notification

import (
	"context"
	"testing"

	"github.com/shurco/gosign/internal/models"
)

// MockProvider is a mock for notification provider
type MockProvider struct {
	SendCalled       bool
	LastNotification *models.Notification
	ShouldFail       bool
}

func (m *MockProvider) Send(ctx context.Context, notification *models.Notification) error {
	m.SendCalled = true
	m.LastNotification = notification

	if m.ShouldFail {
		return &mockError{message: "send failed"}
	}

	return nil
}

func (m *MockProvider) Type() models.NotificationType {
	return models.NotificationTypeEmail
}

type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

func TestService_RegisterProvider(t *testing.T) {
	service := NewService()
	provider := &MockProvider{}

	service.RegisterProvider(provider)

	if !service.CanSend(models.NotificationTypeEmail) {
		t.Error("Expected service to support email notifications")
	}

	if service.CanSend(models.NotificationTypeSMS) {
		t.Error("Expected service to not support SMS notifications")
	}
}

func TestService_Send(t *testing.T) {
	tests := []struct {
		name          string
		notification  *models.Notification
		providerFails bool
		wantStatus    models.NotificationStatus
		wantErr       bool
	}{
		{
			name: "successful send",
			notification: &models.Notification{
				ID:        "notif-123",
				Type:      models.NotificationTypeEmail,
				Recipient: "test@example.com",
				Status:    models.NotificationStatusPending,
			},
			providerFails: false,
			wantStatus:    models.NotificationStatusSent,
			wantErr:       false,
		},
		{
			name: "send error",
			notification: &models.Notification{
				ID:        "notif-456",
				Type:      models.NotificationTypeEmail,
				Recipient: "test@example.com",
				Status:    models.NotificationStatusPending,
			},
			providerFails: true,
			wantStatus:    models.NotificationStatusFailed,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService()
			provider := &MockProvider{ShouldFail: tt.providerFails}
			service.RegisterProvider(provider)

			err := service.Send(tt.notification)

			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !provider.SendCalled {
				t.Error("Expected provider.Send to be called")
			}

			if provider.LastNotification.ID != tt.notification.ID {
				t.Error("Expected notification to be passed to provider")
			}

			if tt.notification.Status != tt.wantStatus {
				t.Errorf("Expected status %s, got %s", tt.wantStatus, tt.notification.Status)
			}
		})
	}
}

func TestService_SendNoProvider(t *testing.T) {
	service := NewService()

	err := service.Send(&models.Notification{Type: models.NotificationTypeEmail})
	if err == nil {
		t.Error("Expected error when no provider is registered")
	}
}

func TestService_CanSend(t *testing.T) {
	tests := []struct {
		name             string
		notificationType models.NotificationType
		registerProvider bool
		want             bool
	}{
		{
			name:             "registered provider",
			notificationType: models.NotificationTypeEmail,
			registerProvider: true,
			want:             true,
		},
		{
			name:             "unregistered provider",
			notificationType: models.NotificationTypeSMS,
			registerProvider: false,
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService()

			if tt.registerProvider {
				provider := &MockProvider{}
				service.RegisterProvider(provider)
			}

			got := service.CanSend(tt.notificationType)

			if got != tt.want {
				t.Errorf("CanSend() = %v, want %v", got, tt.want)
			}
		})
	}
}
