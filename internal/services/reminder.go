package services

import (
	"context"
	"fmt"
	"time"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/notification"
)

// ReminderRepository represents database operations for reminders
type ReminderRepository interface {
	GetSubmitters(ctx context.Context, submissionID string) ([]*models.Submitter, error)
}

// ReminderService manages reminders for submitters
type ReminderService struct {
	notificationSvc *notification.Service
	repo            ReminderRepository
}

// NewReminderService creates a new reminder service
func NewReminderService(notificationSvc *notification.Service, repo ReminderRepository) *ReminderService {
	return &ReminderService{
		notificationSvc: notificationSvc,
		repo:            repo,
	}
}

// ScheduleReminders schedules reminders for submission according to settings
func (s *ReminderService) ScheduleReminders(ctx context.Context, submission *models.Submission, template *models.Template) error {
	if template.Settings == nil || !template.Settings.ReminderEnabled {
		return nil // Reminders disabled
	}

	// Get submitters for this submission
	submitters, err := s.repo.GetSubmitters(ctx, submission.ID)
	if err != nil {
		return fmt.Errorf("failed to get submitters: %w", err)
	}

	// Get submission title from metadata or use template name
	submissionTitle := template.Name
	if title, ok := submission.Metadata["title"].(string); ok && title != "" {
		submissionTitle = title
	}

	// Create scheduled notification for each day from settings
	for _, days := range template.Settings.ReminderDays {
		scheduledAt := time.Now().Add(time.Duration(days) * 24 * time.Hour)

		// Only if submission hasn't expired yet
		if submission.ExpiredAt != nil && scheduledAt.After(*submission.ExpiredAt) {
			continue
		}

		// Create reminder for each incomplete submitter
		for _, submitter := range submitters {
			if submitter.Status == models.SubmitterStatusCompleted ||
				submitter.Status == models.SubmitterStatusDeclined {
				continue
			}

			notif := &models.Notification{
				Type:      models.NotificationTypeEmail,
				Recipient: submitter.Email,
				Template:  "reminder",
				Context: map[string]any{
					"submitter_name":   submitter.Name,
					"submission_title": submissionTitle,
					"signing_link":     fmt.Sprintf("/s/%s", submitter.Slug),
					"days_left":        days,
				},
				Status:      models.NotificationStatusPending,
				ScheduledAt: &scheduledAt,
				RelatedID:   &submitter.ID,
			}

			if err := s.notificationSvc.Schedule(notif); err != nil {
				return fmt.Errorf("failed to schedule reminder: %w", err)
			}
		}
	}

	return nil
}

// SendImmediate sends a reminder immediately (on user request)
func (s *ReminderService) SendImmediate(submitter *models.Submitter, submissionTitle string) error {
	notif := &models.Notification{
		Type:      models.NotificationTypeEmail,
		Recipient: submitter.Email,
		Template:  "reminder",
		Context: map[string]any{
			"submitter_name":   submitter.Name,
			"submission_title": submissionTitle,
			"signing_link":     fmt.Sprintf("/s/%s", submitter.Slug),
			"manual":           true, // Flag indicating this is a manual reminder
		},
		Status:    models.NotificationStatusPending,
		RelatedID: &submitter.ID,
	}

	return s.notificationSvc.Send(notif)
}

// CancelReminders cancels all scheduled reminders for submission
func (s *ReminderService) CancelReminders(submissionID string) error {
	return s.notificationSvc.CancelScheduled(submissionID)
}

// ProcessScheduled processes all scheduled notifications (called by worker)
func (s *ReminderService) ProcessScheduled() error {
	// Get all pending notifications with scheduledAt <= now
	notifications, err := s.notificationSvc.GetScheduledReady()
	if err != nil {
		return fmt.Errorf("failed to get scheduled notifications: %w", err)
	}

	for _, notif := range notifications {
		// Check that submission is still active
		// (could add submitter status check via repo)
		
		if err := s.notificationSvc.Send(notif); err != nil {
			// Log error but continue processing others
			fmt.Printf("failed to send scheduled notification %s: %v\n", notif.ID, err)
			continue
		}
	}

	return nil
}

