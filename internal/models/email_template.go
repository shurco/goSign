package models

import "time"

// EmailTemplate represents an email template stored in the database
type EmailTemplate struct {
	ID        string     `json:"id"`
	AccountID *string    `json:"account_id,omitempty"` // NULL for system templates
	Name      string     `json:"name"`                 // 'base', 'invitation', 'reminder', 'completed'
	Locale    string     `json:"locale"`               // Language code (en, ru, es, fr, de, it, pt)
	Subject   string     `json:"subject,omitempty"`     // Email subject (for content templates)
	Content   string     `json:"content"`              // HTML template content
	IsSystem  bool       `json:"is_system"`            // System templates cannot be deleted
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// EmailTemplateRequest represents a request to create/update an email template
type EmailTemplateRequest struct {
	Name    string `json:"name" validate:"required"`
	Locale  string `json:"locale" validate:"required"` // Language code
	Subject string `json:"subject,omitempty"`
	Content string `json:"content" validate:"required"`
}
