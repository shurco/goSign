package models

import "time"

// Webhook represents webhook endpoint for integrations
type Webhook struct {
	ID              string    `json:"id"`
	AccountID       string    `json:"account_id"`
	URL             string    `json:"url"`
	Events          []string  `json:"events"` // ["submission.created", "submission.completed", etc.]
	Secret          string    `json:"secret"`
	Enabled         bool      `json:"enabled"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	FailureCount    int       `json:"failure_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// WebhookEvent represents event for webhook
type WebhookEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]any `json:"data"`
}

