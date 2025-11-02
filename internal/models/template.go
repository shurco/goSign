package models

import "time"

// TemplateSettings contains template settings
type TemplateSettings struct {
	EmbeddingEnabled bool   `json:"embedding_enabled"`
	WebhookEnabled   bool   `json:"webhook_enabled"`
	ExpirationDays   int    `json:"expiration_days,omitempty"`
	CompanyLogoID    string `json:"company_logo_id,omitempty"`
	ReminderEnabled  bool   `json:"reminder_enabled"`
	ReminderDays     []int  `json:"reminder_days,omitempty"` // [1, 3, 7] - reminders after N days
}

// Template is ...
type Template struct {
	ID             string            `json:"id"`
	FolderID       string            `json:"folder_id"`
	OrganizationID string            `json:"organization_id,omitempty"`
	Slug           string            `json:"slug"`
	Name           string            `json:"name"`
	Description     string            `json:"description,omitempty"`
	Source          string            `json:"source,omitempty"`
	Author          *Author           `json:"author,omitempty"`
	Submitters      []Submitter       `json:"submitters"`
	Fields          []Field           `json:"fields"`
	Schema          []Schema          `json:"schema"`
	Documents       []Document        `json:"documents"`
	Settings        *TemplateSettings `json:"settings,omitempty"`
	Category        string            `json:"category,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	IsFavorite      bool              `json:"is_favorite"`
	PreviewImageID  string            `json:"preview_image_id,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	ArchivedAt      *time.Time        `json:"archived_at,omitempty"`
}

// Author is ...
type Author struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// SubmitterStatus represents signer status
type SubmitterStatus string

const (
	SubmitterStatusPending   SubmitterStatus = "pending"
	SubmitterStatusOpened    SubmitterStatus = "opened"
	SubmitterStatusCompleted SubmitterStatus = "completed"
	SubmitterStatusDeclined  SubmitterStatus = "declined"
)

// Submitter represents document signer
type Submitter struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Email         string           `json:"email"`
	Phone         string           `json:"phone,omitempty"`
	Slug          string           `json:"slug"` // unique signing link
	Status        SubmitterStatus  `json:"status"`
	SubmissionID  string           `json:"submission_id"`
	Order         int              `json:"order"` // signing order for sequential mode
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
	DeclinedAt    *time.Time       `json:"declined_at,omitempty"`
	SentAt        *time.Time       `json:"sent_at,omitempty"`
	OpenedAt      *time.Time       `json:"opened_at,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// FieldType represents field type in template
type FieldType string

// 14 field types for document templates
const (
	FieldTypeSignature   FieldType = "signature"
	FieldTypeInitials    FieldType = "initials"
	FieldTypeDate        FieldType = "date"
	FieldTypeText        FieldType = "text"
	FieldTypeNumber      FieldType = "number"
	FieldTypeCheckbox    FieldType = "checkbox"
	FieldTypeRadio       FieldType = "radio"
	FieldTypeSelect      FieldType = "select"
	FieldTypeMultiSelect FieldType = "multi_select"
	FieldTypeFile        FieldType = "file"
	FieldTypeImage       FieldType = "image"
	FieldTypeCells       FieldType = "cells"
	FieldTypeStamp       FieldType = "stamp"
	FieldTypePayment     FieldType = "payment"
)

// Field is ...
type Field struct {
	ID           string    `json:"id"`
	SubmitterID  string    `json:"submitter_id"`
	Name         string    `json:"name"`
	Type         FieldType `json:"type"`
	Required     bool      `json:"required"`
	DefaultValue string    `json:"default_value,omitempty"`
	Options      []string  `json:"options,omitempty"` // for select, radio
	Validation   string    `json:"validation,omitempty"`
	Areas        []*Areas  `json:"areas,omitempty"`
}

// Areas is ...
type Areas struct {
	AttachmentID string  `json:"attachment_id"`
	Page         int     `json:"page"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	W            float64 `json:"w"`
	H            float64 `json:"z"`
}

// Schema is ...
type Schema struct {
	AttachmentID string `json:"attachment_id"`
	Name         string `json:"name"`
}

// Document is ...
type Document struct {
	ID            string          `json:"id"`
	URL           string          `json:"url"`
	FileName      string          `json:"filename,omitempty"`
	Metadata      DocMetadata     `json:"metadata"`
	PreviewImages []PreviewImages `json:"preview_images"`
	CreatedAt     time.Time       `json:"created_at"`
}

// DocMetadata is ...
type DocMetadata struct {
	Analyzed bool   `json:"analyzed,omitempty"`
	Pdf      Pdf    `json:"pdf"`
	Sha256   string `json:"sha256,omitempty"`
}

// Pdf os ...
type Pdf struct {
	Annotations   []*Annotations `json:"annotations,omitempty"`
	NumberOfPages int            `json:"number_of_pages"`
}

// Annotations is ...
type Annotations struct {
	Type  string  `json:"type"`
	Value string  `json:"value"`
	Page  int     `json:"page"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w"`
	H     float64 `json:"z"`
}

// PreviewImages is ...
type PreviewImages struct {
	ID         string      `json:"id"`
	RecordType string      `json:"record_type,omitempty"`
	RecordID   string      `json:"record_id,omitempty"`
	BlobID     string      `json:"blob_id,omitempty"`
	Metadata   ImgMetadata `json:"metadata"`
	FileName   string      `json:"filename"`
}

// ImgMetadata is ...
type ImgMetadata struct {
	Analyzed   bool `json:"analyzed,omitempty"`
	Identified bool `json:"identified,omitempty"`
	Width      int  `json:"width"`
	Height     int  `json:"height"`
}

// TemplateFavorite represents a user's favorite template
type TemplateFavorite struct {
	ID         string    `json:"id"`
	TemplateID string    `json:"template_id"`
	UserID     string    `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// TemplateFolder represents a folder for organizing templates
type TemplateFolder struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parent_id,omitempty"` // For nested folders
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

