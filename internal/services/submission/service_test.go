package submission

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock notification provider for testing
type mockNotificationProvider struct{}

func (m *mockNotificationProvider) Send(ctx context.Context, notification *models.Notification) error {
	return nil // Always succeed
}

func (m *mockNotificationProvider) Type() models.NotificationType {
	return models.NotificationTypeEmail
}

// Mock notification repository for testing
type mockNotificationRepository struct{}

func (m *mockNotificationRepository) Create(notification *models.Notification) error {
	return nil
}

func (m *mockNotificationRepository) GetScheduledReady() ([]*models.Notification, error) {
	return nil, nil
}

func (m *mockNotificationRepository) UpdateStatus(id string, status models.NotificationStatus) error {
	return nil
}

func (m *mockNotificationRepository) CancelByRelatedID(relatedID string) error {
	return nil
}

// Create a mock notification service
func createMockNotificationService() *notification.Service {
	repo := &mockNotificationRepository{}
	service := notification.NewService(repo)
	provider := &mockNotificationProvider{}
	service.RegisterProvider(provider)
	return service
}

// Mock repository for testing
type mockRepository struct {
	submissions map[string]*models.Submission
	submitters  map[string]*models.Submitter
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		submissions: make(map[string]*models.Submission),
		submitters:  make(map[string]*models.Submitter),
	}
}

func (m *mockRepository) GetSubmission(ctx context.Context, id string) (*models.Submission, error) {
	if sub, ok := m.submissions[id]; ok {
		return sub, nil
	}
	return nil, errors.New("submission not found")
}

func (m *mockRepository) GetSubmitters(ctx context.Context, submissionID string) ([]*models.Submitter, error) {
	var result []*models.Submitter
	for _, sub := range m.submitters {
		if sub.SubmissionID == submissionID {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (m *mockRepository) GetSubmitter(ctx context.Context, id string) (*models.Submitter, error) {
	if sub, ok := m.submitters[id]; ok {
		return sub, nil
	}
	return nil, errors.New("submitter not found")
}

func (m *mockRepository) CreateSubmitter(ctx context.Context, submitter *models.Submitter) error {
	m.submitters[submitter.ID] = submitter
	return nil
}

func (m *mockRepository) GetSubmittersByOrder(ctx context.Context, submissionID string, order int) ([]*models.Submitter, error) {
	var result []*models.Submitter
	for _, sub := range m.submitters {
		if sub.SubmissionID == submissionID && sub.Order == order {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (m *mockRepository) UpdateSubmissionState(ctx context.Context, id string, state SubmissionState) error {
	if sub, ok := m.submissions[id]; ok {
		sub.Status = models.SubmissionStatus(state)
		return nil
	}
	return errors.New("submission not found")
}

func (m *mockRepository) UpdateSubmitterStatus(ctx context.Context, id string, status models.SubmitterStatus) error {
	if sub, ok := m.submitters[id]; ok {
		sub.Status = status
		// Simulate CompletedAt being set when status is completed
		if status == models.SubmitterStatusCompleted {
			now := time.Now()
			sub.CompletedAt = &now
		}
		return nil
	}
	return errors.New("submitter not found")
}

func (m *mockRepository) CreateEvent(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRepository) CreateSubmission(ctx context.Context, submission *models.Submission) error {
	m.submissions[submission.ID] = submission
	return nil
}


func TestCheckCompletion(t *testing.T) {
	tests := []struct {
		name         string
		submissionID string
		setupFunc    func(*mockRepository)
		wantStatus   models.SubmissionStatus
		wantErr      bool
	}{
		{
			name:         "all submitters completed - submission completes",
			submissionID: "sub1",
			setupFunc: func(repo *mockRepository) {
				repo.submissions["sub1"] = &models.Submission{
					ID:     "sub1",
					Status: models.SubmissionStatus(StateInProgress),
					// No CreatedByID to avoid notification sending
				}
				repo.submitters["submitter1"] = &models.Submitter{
					ID:           "submitter1",
					SubmissionID: "sub1",
					Status:       models.SubmitterStatusCompleted,
				}
				repo.submitters["submitter2"] = &models.Submitter{
					ID:           "submitter2",
					SubmissionID: "sub1",
					Status:       models.SubmitterStatusCompleted,
				}
			},
			wantStatus: models.SubmissionStatus(StateCompleted),
			wantErr:    false,
		},
		{
			name:         "one submitter pending - submission stays in progress",
			submissionID: "sub2",
			setupFunc: func(repo *mockRepository) {
				repo.submissions["sub2"] = &models.Submission{
					ID:          "sub2",
					Status:      models.SubmissionStatus(StateInProgress),
					CreatedByID: "user1",
				}
				repo.submitters["submitter3"] = &models.Submitter{
					ID:           "submitter3",
					SubmissionID: "sub2",
					Status:       models.SubmitterStatusCompleted,
				}
				repo.submitters["submitter4"] = &models.Submitter{
					ID:           "submitter4",
					SubmissionID: "sub2",
					Status:       models.SubmitterStatusPending,
				}
			},
			wantStatus: models.SubmissionStatus(StateInProgress),
			wantErr:    false,
		},
		{
			name:         "submission not found returns error",
			submissionID: "nonexistent",
			setupFunc:    func(repo *mockRepository) {},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newMockRepository()
			tt.setupFunc(repo)

			service := NewService(repo, nil, nil)
			err := service.CheckCompletion(context.Background(), tt.submissionID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			sub, err := repo.GetSubmission(context.Background(), tt.submissionID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, sub.Status)
		})
	}
}

func TestHandleDecline(t *testing.T) {
	tests := []struct {
		name         string
		submissionID string
		reason       string
		setupFunc    func(*mockRepository)
		wantStatus   models.SubmissionStatus
		wantErr      bool
	}{
		{
			name:         "decline submission successfully",
			submissionID: "sub1",
			reason:       "User declined",
			setupFunc: func(repo *mockRepository) {
				repo.submissions["sub1"] = &models.Submission{
					ID:     "sub1",
					Status: models.SubmissionStatus(StateInProgress),
					// No CreatedByID to avoid notification sending
				}
			},
			wantStatus: models.SubmissionStatus(StateCancelled),
			wantErr:    false,
		},
		{
			name:         "submission not found returns error",
			submissionID: "nonexistent",
			reason:       "Test reason",
			setupFunc:    func(repo *mockRepository) {},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newMockRepository()
			tt.setupFunc(repo)

			service := NewService(repo, nil, nil)
			err := service.HandleDecline(context.Background(), tt.submissionID, tt.reason)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			sub, err := repo.GetSubmission(context.Background(), tt.submissionID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, sub.Status)
		})
	}
}

func TestResendInvitation(t *testing.T) {
	tests := []struct {
		name        string
		submitterID string
		setupFunc   func(*mockRepository)
		wantErr     bool
	}{
		{
			name:        "resend invitation requires notification service",
			submitterID: "submitter1",
			setupFunc: func(repo *mockRepository) {
				repo.submissions["sub1"] = &models.Submission{
					ID:     "sub1",
					Status: models.SubmissionStatus(StateInProgress),
				}
				repo.submitters["submitter1"] = &models.Submitter{
					ID:           "submitter1",
					SubmissionID: "sub1",
					Status:       models.SubmitterStatusPending,
					Email:        "test@example.com",
				}
			},
			wantErr: true, // Fails without notification service (integration test needed)
		},
		{
			name:        "non-existent submitter returns error",
			submitterID: "nonexistent",
			setupFunc:   func(repo *mockRepository) {},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newMockRepository()
			tt.setupFunc(repo)

			service := NewService(repo, nil, nil)
			
			// Skip tests that require notification service for unit testing
			// These should be covered by integration tests
			if tt.name == "resend invitation requires notification service" {
				t.Skip("Requires notification service - integration test needed")
			}
			
			err := service.ResendInvitation(context.Background(), tt.submitterID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	tests := []struct {
		name        string
		submitterID string
		setupFunc   func(*mockRepository)
		wantStatus  models.SubmitterStatus
		wantErr     bool
	}{
		{
			name:        "complete submitter successfully",
			submitterID: "submitter1",
			setupFunc: func(repo *mockRepository) {
				repo.submitters["submitter1"] = &models.Submitter{
					ID:           "submitter1",
					SubmissionID: "sub1",
					Status:       models.SubmitterStatusPending,
				}
			},
			wantStatus: models.SubmitterStatusCompleted,
			wantErr:    false,
		},
		{
			name:        "non-existent submitter returns error",
			submitterID: "nonexistent",
			setupFunc:   func(repo *mockRepository) {},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newMockRepository()
			tt.setupFunc(repo)

			service := NewService(repo, nil, nil)
			err := service.Complete(context.Background(), tt.submitterID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			sub, err := repo.GetSubmitter(context.Background(), tt.submitterID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, sub.Status)
			assert.NotNil(t, sub.CompletedAt, "CompletedAt should be set")
		})
	}
}

func TestDecline(t *testing.T) {
	tests := []struct {
		name        string
		submitterID string
		reason      string
		setupFunc   func(*mockRepository)
		wantStatus  models.SubmitterStatus
		wantErr     bool
	}{
		{
			name:        "decline submitter successfully",
			submitterID: "submitter1",
			reason:      "Not interested",
			setupFunc: func(repo *mockRepository) {
				repo.submitters["submitter1"] = &models.Submitter{
					ID:           "submitter1",
					SubmissionID: "sub1",
					Status:       models.SubmitterStatusPending,
				}
			},
			wantStatus: models.SubmitterStatusDeclined,
			wantErr:    false,
		},
		{
			name:        "non-existent submitter returns error",
			submitterID: "nonexistent",
			reason:      "Test reason",
			setupFunc:   func(repo *mockRepository) {},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newMockRepository()
			tt.setupFunc(repo)

			service := NewService(repo, nil, nil)
			err := service.Decline(context.Background(), tt.submitterID, tt.reason)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			sub, err := repo.GetSubmitter(context.Background(), tt.submitterID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, sub.Status)
		})
	}
}

func TestService_Send_SequentialValidation(t *testing.T) {
	tests := []struct {
		name        string
		signingMode models.SigningMode
		numSubmitters int
		wantErr     bool
		errContains string
	}{
		{
			name:        "sequential mode with 1 submitter fails",
			signingMode: models.SigningModeSequential,
			numSubmitters: 1,
			wantErr:     true,
			errContains: "sequential signing mode requires at least 2 submitters",
		},
		{
			name:        "sequential mode with 2 submitters succeeds",
			signingMode: models.SigningModeSequential,
			numSubmitters: 2,
			wantErr:     false,
		},
		{
			name:        "parallel mode with 1 submitter succeeds",
			signingMode: models.SigningModeParallel,
			numSubmitters: 1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newMockRepository()

			// Create submission
			submission := &models.Submission{
				ID:          "test-submission",
				TemplateID:  "test-template",
				SigningMode: tt.signingMode,
				Status:      models.SubmissionStatusDraft,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			repo.submissions[submission.ID] = submission

			// Create submitters
			for i := 0; i < tt.numSubmitters; i++ {
				submitter := &models.Submitter{
					ID:           fmt.Sprintf("submitter-%d", i),
					Name:         fmt.Sprintf("Submitter %d", i),
					Email:        fmt.Sprintf("submitter%d@example.com", i),
					SubmissionID: submission.ID,
					Order:        i,
					Status:       models.SubmitterStatusPending,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				repo.submitters[submitter.ID] = submitter
			}

			service := NewService(repo, createMockNotificationService(), nil)
			err := service.Send(context.Background(), submission.ID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
