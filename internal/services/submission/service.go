package submission

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/notification"
	"github.com/shurco/gosign/pkg/webhook"
)

// SubmissionState represents submission state
type SubmissionState string

const (
	StateDraft      SubmissionState = "draft"
	StatePending    SubmissionState = "pending"
	StateInProgress SubmissionState = "in_progress"
	StateCompleted  SubmissionState = "completed"
	StateExpired    SubmissionState = "expired"
	StateCancelled  SubmissionState = "cancelled"
)

// Repository is an interface for database operations
type Repository interface {
	CreateSubmission(ctx context.Context, submission *models.Submission) error
	GetSubmission(ctx context.Context, id string) (*models.Submission, error)
	UpdateSubmissionState(ctx context.Context, id string, state SubmissionState) error
	GetSubmitters(ctx context.Context, submissionID string) ([]*models.Submitter, error)
	GetSubmitter(ctx context.Context, id string) (*models.Submitter, error)
	UpdateSubmitterStatus(ctx context.Context, id string, status models.SubmitterStatus) error
	CreateEvent(ctx context.Context, event *models.Event) error
}

// Service manages submission workflow
type Service struct {
	repo              Repository
	notificationSvc   *notification.Service
	webhookDispatcher *webhook.Dispatcher
}

// NewService creates a new service
func NewService(repo Repository, notificationSvc *notification.Service, webhookDispatcher *webhook.Dispatcher) *Service {
	return &Service{
		repo:              repo,
		notificationSvc:   notificationSvc,
		webhookDispatcher: webhookDispatcher,
	}
}

// CreateSubmissionInput is input data for creating a submission
type CreateSubmissionInput struct {
	TemplateID  string
	CreatedByID string
	Submitters  []SubmitterInput
}

// SubmitterInput is submitter data
type SubmitterInput struct {
	Name  string
	Email string
	Phone string
}

// Create creates a new submission in draft status
func (s *Service) Create(ctx context.Context, input CreateSubmissionInput) (*models.Submission, error) {
	submission := &models.Submission{
		ID:         uuid.New().String(),
		TemplateID: input.TemplateID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repo.CreateSubmission(ctx, submission); err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	_ = s.logEvent(ctx, models.EventSubmissionCreated, input.CreatedByID, "submission", submission.ID, nil)

	log.Info().Str("submission_id", submission.ID).Msg("Submission created")
	return submission, nil
}

// Send sends invitations to submitters and changes status to pending
func (s *Service) Send(ctx context.Context, submissionID string) error {
	submission, err := s.repo.GetSubmission(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("failed to get submission: %w", err)
	}

	// Get submitters
	submitters, err := s.repo.GetSubmitters(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("failed to get submitters: %w", err)
	}

	if len(submitters) == 0 {
		return fmt.Errorf("no submitters found for submission")
	}

	// Send invitation to first submitter (or all, if parallel workflow)
	firstSubmitter := submitters[0]
	if err := s.sendInvitation(ctx, submission, firstSubmitter); err != nil {
		return fmt.Errorf("failed to send invitation: %w", err)
	}

	// Update submission status
	if err := s.repo.UpdateSubmissionState(ctx, submissionID, StateInProgress); err != nil {
		return fmt.Errorf("failed to update submission state: %w", err)
	}

	// Send webhook
	s.sendWebhook(ctx, models.EventSubmissionCreated, submission)

	log.Info().Str("submission_id", submissionID).Msg("Submission sent")
	return nil
}

// Complete finishes signing for a specific submitter
func (s *Service) Complete(ctx context.Context, submitterID string) error {
	if err := s.repo.UpdateSubmitterStatus(ctx, submitterID, models.SubmitterStatusCompleted); err != nil {
		return fmt.Errorf("failed to update submitter status: %w", err)
	}

	_ = s.logEvent(ctx, models.EventSubmitterCompleted, "", "submitter", submitterID, nil)

	log.Info().Str("submitter_id", submitterID).Msg("Submitter completed")
	return nil
}

// CheckCompletion checks if all submitters completed and finalizes submission
func (s *Service) CheckCompletion(ctx context.Context, submissionID string) error {
	// Get all submitters
	submitters, err := s.repo.GetSubmitters(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("failed to get submitters: %w", err)
	}

	// Check if all completed
	for _, submitter := range submitters {
		if submitter.Status != models.SubmitterStatusCompleted {
			return nil
		}
	}

	submission, err := s.repo.GetSubmission(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("failed to get submission: %w", err)
	}

	if err := s.repo.UpdateSubmissionState(ctx, submissionID, StateCompleted); err != nil {
		return fmt.Errorf("failed to update submission state: %w", err)
	}

	// Send notification to creator
	if submission.CreatedByID != "" {
		notification := s.createNotification("completion", "Document completed", map[string]any{
			"document_name": "Document",
			"submission_id": submissionID,
		}, "submission", submissionID)

		_ = s.notificationSvc.Send(notification)
	}

	s.sendWebhook(ctx, models.EventSubmissionCompleted, submission)

	log.Info().Str("submission_id", submissionID).Msg("Submission completed")
	return nil
}

// Decline rejects the signing
func (s *Service) Decline(ctx context.Context, submitterID, reason string) error {
	if err := s.repo.UpdateSubmitterStatus(ctx, submitterID, models.SubmitterStatusDeclined); err != nil {
		return fmt.Errorf("failed to update submitter status: %w", err)
	}

	_ = s.logEvent(ctx, models.EventSubmitterDeclined, "", "submitter", submitterID, map[string]any{"reason": reason})

	log.Info().Str("submitter_id", submitterID).Str("reason", reason).Msg("Submitter declined")
	return nil
}

// HandleDecline handles decline process and notifies creator
func (s *Service) HandleDecline(ctx context.Context, submissionID string, reason string) error {
	submission, err := s.repo.GetSubmission(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("failed to get submission: %w", err)
	}

	if err := s.repo.UpdateSubmissionState(ctx, submissionID, StateCancelled); err != nil {
		return fmt.Errorf("failed to update submission state: %w", err)
	}

	// Send notification to creator
	if submission.CreatedByID != "" {
		notification := s.createNotification("declined", "Document signing declined", map[string]any{
			"document_name": "Document",
			"submission_id": submissionID,
			"reason":        reason,
		}, "submission", submissionID)

		_ = s.notificationSvc.Send(notification)
	}

	s.sendWebhook(ctx, models.EventSubmissionCancelled, submission)

	log.Info().Str("submission_id", submissionID).Str("reason", reason).Msg("Submission declined and cancelled")
	return nil
}

// Expire marks submission as expired
func (s *Service) Expire(ctx context.Context, submissionID string) error {
	if err := s.repo.UpdateSubmissionState(ctx, submissionID, StateExpired); err != nil {
		return fmt.Errorf("failed to expire submission: %w", err)
	}

	_ = s.logEvent(ctx, models.EventSubmissionExpired, "", "submission", submissionID, nil)

	log.Info().Str("submission_id", submissionID).Msg("Submission expired")
	return nil
}

// ResendInvitation resends invitation to a submitter
func (s *Service) ResendInvitation(ctx context.Context, submitterID string) error {
	submitter, err := s.repo.GetSubmitter(ctx, submitterID)
	if err != nil {
		return fmt.Errorf("failed to get submitter: %w", err)
	}

	submission, err := s.repo.GetSubmission(ctx, submitter.SubmissionID)
	if err != nil {
		return fmt.Errorf("failed to get submission: %w", err)
	}

	if err := s.sendInvitation(ctx, submission, submitter); err != nil {
		return fmt.Errorf("failed to send invitation: %w", err)
	}

	now := time.Now()
	submitter.SentAt = &now

	_ = s.logEvent(ctx, models.EventSubmitterSent, "", "submitter", submitterID, nil)

	log.Info().Str("submitter_id", submitterID).Msg("Invitation resent")
	return nil
}

// sendInvitation sends an invitation to a submitter
func (s *Service) sendInvitation(ctx context.Context, submission *models.Submission, submitter *models.Submitter) error {
	now := time.Now()
	notification := &models.Notification{
		ID:          uuid.New().String(),
		Type:        models.NotificationTypeEmail,
		Recipient:   submitter.Email,
		Template:    "invitation",
		Subject:     "Document for signing",
		Context: map[string]any{
			"submitter_name": submitter.Name,
			"document_name":  "Document",
			"signing_url":    fmt.Sprintf("/s/%s", submitter.Slug),
			"company_name":   "goSign",
		},
		Status:      models.NotificationStatusPending,
		ScheduledAt: &now,
		RelatedType: "submitter",
		RelatedID:   &submitter.ID,
		CreatedAt:   now,
	}

	if err := s.notificationSvc.Send(notification); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	_ = s.repo.UpdateSubmitterStatus(ctx, submitter.ID, models.SubmitterStatusOpened)
	_ = s.logEvent(ctx, models.EventSubmitterSent, "", "submitter", submitter.ID, nil)

	return nil
}

// sendWebhook sends a webhook event
func (s *Service) sendWebhook(ctx context.Context, eventType string, submission *models.Submission) {
	// TODO: Get webhooks for account from database
	// TODO: Send via dispatcher
	
	webhookEvent := &models.WebhookEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data: map[string]any{
			"submission_id": submission.ID,
			"template_id":   submission.TemplateID,
		},
	}

	_ = webhookEvent // stub
}

// createNotification creates a notification with common fields
func (s *Service) createNotification(template, subject string, context map[string]any, relatedType, relatedID string) *models.Notification {
	return &models.Notification{
		ID:          uuid.New().String(),
		Type:        models.NotificationTypeEmail,
		Template:    template,
		Subject:     subject,
		Context:     context,
		Status:      models.NotificationStatusPending,
		RelatedType: relatedType,
		RelatedID:   &relatedID,
		CreatedAt:   time.Now(),
	}
}

// logEvent logs an event to the database
func (s *Service) logEvent(ctx context.Context, eventType, actorID, resourceType, resourceID string, metadata map[string]any) error {
	event := &models.Event{
		ID:           uuid.New().String(),
		Type:         eventType,
		ActorID:      actorID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
		CreatedAt:    time.Now(),
	}

	return s.repo.CreateEvent(ctx, event)
}

