package notification

import (
	"context"
	"testing"

	"github.com/shurco/gosign/internal/models"
)

// MockProvider is a mock for notification provider
type MockProvider struct {
	SendCalled    bool
	LastNotification *models.Notification
	ShouldFail    bool
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

// MockRepository is a mock for notification repository
type MockRepository struct {
	CreateCalled    bool
	UpdateCalled    bool
	CancelCalled    bool
	Notifications   []*models.Notification
}

func (m *MockRepository) Create(notification *models.Notification) error {
	m.CreateCalled = true
	m.Notifications = append(m.Notifications, notification)
	return nil
}

func (m *MockRepository) GetScheduledReady() ([]*models.Notification, error) {
	return m.Notifications, nil
}

func (m *MockRepository) UpdateStatus(id string, status models.NotificationStatus) error {
	m.UpdateCalled = true
	for _, n := range m.Notifications {
		if n.ID == id {
			n.Status = status
			break
		}
	}
	return nil
}

func (m *MockRepository) CancelByRelatedID(relatedID string) error {
	m.CancelCalled = true
	return nil
}

func TestService_RegisterProvider(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
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
			repo := &MockRepository{}
			service := NewService(repo)
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
		})
	}
}

func TestService_Schedule(t *testing.T) {
	tests := []struct {
		name         string
		notification *models.Notification
		wantErr      bool
	}{
		{
			name: "schedule notification",
			notification: &models.Notification{
				ID:        "notif-123",
				Type:      models.NotificationTypeEmail,
				Recipient: "test@example.com",
				Status:    models.NotificationStatusPending,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{}
			service := NewService(repo)

			err := service.Schedule(tt.notification)

			if (err != nil) != tt.wantErr {
				t.Errorf("Schedule() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !repo.CreateCalled {
				t.Error("Expected repository.Create to be called")
			}

			if tt.notification.Status != models.NotificationStatusPending {
				t.Errorf("Expected status to be pending, got %s", tt.notification.Status)
			}
		})
	}
}

func TestService_GetScheduledReady(t *testing.T) {
	repo := &MockRepository{
		Notifications: []*models.Notification{
			{
				ID:     "notif-1",
				Type:   models.NotificationTypeEmail,
				Status: models.NotificationStatusPending,
			},
			{
				ID:     "notif-2",
				Type:   models.NotificationTypeEmail,
				Status: models.NotificationStatusPending,
			},
		},
	}
	service := NewService(repo)

	notifications, err := service.GetScheduledReady()

	if err != nil {
		t.Errorf("GetScheduledReady() error = %v", err)
	}

	if len(notifications) != 2 {
		t.Errorf("Expected 2 notifications, got %d", len(notifications))
	}
}

func TestService_CancelScheduled(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.CancelScheduled("sub-123")

	if err != nil {
		t.Errorf("CancelScheduled() error = %v", err)
	}

	if !repo.CancelCalled {
		t.Error("Expected repository.CancelByRelatedID to be called")
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
			repo := &MockRepository{}
			service := NewService(repo)

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

