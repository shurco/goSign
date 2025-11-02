package tasks

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/shurco/gosign/internal/services"
)

// RemindersTask processes scheduled reminders
type RemindersTask struct {
	reminderSvc *services.ReminderService
}

// NewRemindersTask creates a new task for processing reminders
func NewRemindersTask(reminderSvc *services.ReminderService) *RemindersTask {
	return &RemindersTask{
		reminderSvc: reminderSvc,
	}
}

// Execute processes scheduled notifications
func (t *RemindersTask) Execute(ctx context.Context) error {
	log.Info().Msg("Processing scheduled reminders")

	if err := t.reminderSvc.ProcessScheduled(); err != nil {
		return fmt.Errorf("failed to process reminders: %w", err)
	}

	return nil
}

// ShouldRetry determines if the task should be retried on error
func (t *RemindersTask) ShouldRetry(err error) bool {
	// Always retry on errors - scheduled notifications are important
	return true
}

