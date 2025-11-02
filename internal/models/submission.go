package models

import "time"

// SigningMode represents signing mode
type SigningMode string

const (
	SigningModeSequential SigningMode = "sequential"
	SigningModeParallel   SigningMode = "parallel"
)

// SubmissionStatus represents submission status
type SubmissionStatus string

const (
	SubmissionStatusDraft      SubmissionStatus = "draft"
	SubmissionStatusPending    SubmissionStatus = "pending"
	SubmissionStatusInProgress SubmissionStatus = "in_progress"
	SubmissionStatusCompleted  SubmissionStatus = "completed"
	SubmissionStatusExpired    SubmissionStatus = "expired"
	SubmissionStatusCancelled  SubmissionStatus = "cancelled"
)

// Submission represents a document for signing
type Submission struct {
	ID          string           `json:"id"`
	TemplateID  string           `json:"template_id"`
	AccountID   string           `json:"account_id,omitempty"`
	CreatedByID string           `json:"created_by_id,omitempty"`
	Status      SubmissionStatus `json:"status"`
	SigningMode SigningMode      `json:"signing_mode"`
	ExpiredAt   *time.Time       `json:"expired_at,omitempty"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

