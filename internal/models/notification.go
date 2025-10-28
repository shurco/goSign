package models

import "time"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail    NotificationType = "email"
	NotificationTypeSMS      NotificationType = "sms"
	NotificationTypeReminder NotificationType = "reminder"
)

// NotificationStatus represents notification status
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSending   NotificationStatus = "sending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// Notification represents a universal model for all notification types
type Notification struct {
	ID           string                 `json:"id" db:"id"`
	Type         NotificationType       `json:"type" db:"type"`
	Recipient    string                 `json:"recipient" db:"recipient"`
	Template     string                 `json:"template" db:"template"`
	Subject      string                 `json:"subject,omitempty" db:"subject"`
	Body         string                 `json:"body,omitempty" db:"body"`
	Context      map[string]any         `json:"context" db:"context"`
	Status       NotificationStatus     `json:"status" db:"status"`
	ScheduledAt  *time.Time             `json:"scheduled_at,omitempty" db:"scheduled_at"`
	SentAt       *time.Time             `json:"sent_at,omitempty" db:"sent_at"`
	RelatedType  string                 `json:"related_type,omitempty" db:"related_type"`
	RelatedID    *string                `json:"related_id,omitempty" db:"related_id"`
	ErrorMessage string                 `json:"error_message,omitempty" db:"error_message"`
	RetryCount   int                    `json:"retry_count" db:"retry_count"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

