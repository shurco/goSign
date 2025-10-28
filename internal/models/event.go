package models

import "time"

// Event represents universal event logging
type Event struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"` // submission.created, submitter.completed, etc.
	ActorID      string                 `json:"actor_id,omitempty"`
	ResourceType string                 `json:"resource_type"` // submission, submitter, template, etc.
	ResourceID   string                 `json:"resource_id"`
	Metadata     map[string]any `json:"metadata"`
	IP           string                 `json:"ip,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// EventType constants for event types
const (
	EventSubmissionCreated   = "submission.created"
	EventSubmissionSent      = "submission.sent"
	EventSubmissionCompleted = "submission.completed"
	EventSubmissionExpired   = "submission.expired"
	EventSubmissionCancelled = "submission.cancelled"

	EventSubmitterSent      = "submitter.sent"
	EventSubmitterOpened    = "submitter.opened"
	EventSubmitterCompleted = "submitter.completed"
	EventSubmitterDeclined  = "submitter.declined"

	EventTemplateCreated = "template.created"
	EventTemplateUpdated = "template.updated"
	EventTemplateDeleted = "template.deleted"
)

