package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TemplateSettings contains template settings
type TemplateSettings struct {
	EmbeddingEnabled bool   `json:"embedding_enabled"`
	WebhookEnabled   bool   `json:"webhook_enabled"`
	ExpirationDays   int    `json:"expiration_days,omitempty"`
	CompanyLogoID    string `json:"company_logo_id,omitempty"`
	ReminderEnabled  bool   `json:"reminder_enabled"`
	ReminderDays     []int  `json:"reminder_days,omitempty"` // [1, 3, 7] - reminders after N days
}

// Translation represents template translations for different locales
type Translation struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Fields      map[string]string `json:"fields,omitempty"` // field_id -> translated label
}

// Template is ...
type Template struct {
	ID             string                 `json:"id"`
	FolderID       string                 `json:"folder_id"`
	OrganizationID string                 `json:"organization_id,omitempty"`
	Slug           string                 `json:"slug"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Source         string                 `json:"source,omitempty"`
	Author         *Author                `json:"author,omitempty"`
	Submitters     []Submitter            `json:"submitters"`
	// SubmitterCount is a computed field used by list/search endpoints to avoid
	// sending full submitters array when only the count is needed.
	SubmitterCount int                    `json:"submitter_count,omitempty"`
	Fields         []Field                `json:"fields"`
	Schema         []Schema               `json:"schema"`
	Documents      []Document             `json:"documents"`
	Settings       *TemplateSettings      `json:"settings,omitempty"`
	Category       string                 `json:"category,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	IsFavorite     bool                   `json:"is_favorite"`
	PreviewImageID string                 `json:"preview_image_id,omitempty"`
	DefaultLocale  string                 `json:"default_locale,omitempty"`
	Translations   map[string]Translation `json:"translations,omitempty"` // locale -> Translation
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	ArchivedAt     *time.Time             `json:"archived_at,omitempty"`
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

// ConditionOperator represents comparison operator
type ConditionOperator string

const (
	ConditionEquals       ConditionOperator = "equals"
	ConditionNotEquals    ConditionOperator = "not_equals"
	ConditionContains     ConditionOperator = "contains"
	ConditionNotContains  ConditionOperator = "not_contains"
	ConditionGreaterThan  ConditionOperator = "greater_than"
	ConditionLessThan     ConditionOperator = "less_than"
	ConditionIsEmpty      ConditionOperator = "is_empty"
	ConditionIsNotEmpty   ConditionOperator = "is_not_empty"
)

// ConditionAction represents action to take when condition is met
type ConditionAction string

const (
	ActionShow    ConditionAction = "show"
	ActionHide    ConditionAction = "hide"
	ActionRequire ConditionAction = "require"
	ActionDisable ConditionAction = "disable"
)

// LogicOperator for combining multiple conditions
type LogicOperator string

const (
	LogicAND LogicOperator = "AND"
	LogicOR  LogicOperator = "OR"
)

// FieldCondition single condition rule
type FieldCondition struct {
	FieldID  string            `json:"field_id"`  // target field to check
	Operator ConditionOperator `json:"operator"`
	Value    interface{}       `json:"value"` // value to compare against
}

// FieldConditionGroup allows AND/OR logic
type FieldConditionGroup struct {
	Logic      LogicOperator     `json:"logic"` // AND or OR
	Conditions []FieldCondition  `json:"conditions"`
	Action     ConditionAction   `json:"action"`
}

// Field is ...
type Field struct {
	ID              string               `json:"id"`
	SubmitterID     string               `json:"submitter_id"`
	Name            string               `json:"name"`
	Label           string               `json:"label,omitempty"`        // base label for display
	Type            FieldType            `json:"type"`
	Required        bool                 `json:"required"`
	DefaultValue    string               `json:"default_value,omitempty"`
	Options         FieldOptions         `json:"options,omitempty"` // for select, radio, multiple
	Validation      string               `json:"validation,omitempty"`
	Translations    map[string]string    `json:"translations,omitempty"` // locale -> translated label
	ConditionGroups []FieldConditionGroup `json:"condition_groups,omitempty"`
	Formula         string               `json:"formula,omitempty"`         // e.g., "field_1 + field_2 * 0.2"
	CalculationType string               `json:"calculation_type,omitempty"` // "number", "currency"
	Areas           []*Areas             `json:"areas,omitempty"`
}

// FieldOption represents a selectable option for select/radio/multiple fields.
type FieldOption struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// FieldOptions is a slice with backward-compatible JSON parsing.
// Historically, options were stored as []string. The editor uses []{id,value}.
type FieldOptions []FieldOption

func (o *FieldOptions) UnmarshalJSON(b []byte) error {
	// Accept null
	if string(b) == "null" {
		*o = nil
		return nil
	}

	// Preferred: array of objects
	var objs []FieldOption
	if err := json.Unmarshal(b, &objs); err == nil {
		*o = objs
		return nil
	}

	// Backward: array of strings
	var strs []string
	if err := json.Unmarshal(b, &strs); err == nil {
		out := make([]FieldOption, 0, len(strs))
		for _, s := range strs {
			// Stable ID derived from value to avoid changing IDs between loads.
			id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(s)).String()
			out = append(out, FieldOption{ID: id, Value: s})
		}
		*o = out
		return nil
	}

	return json.Unmarshal(b, &objs) // return original error context
}

// Areas is ...
type Areas struct {
	AttachmentID string  `json:"attachment_id"`
	Page         int     `json:"page"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	W            float64 `json:"w"`
	H            float64 `json:"h"`
}

// UnmarshalJSON accepts both `h` (current) and legacy `z` (height).
func (a *Areas) UnmarshalJSON(b []byte) error {
	type payload struct {
		AttachmentID string   `json:"attachment_id"`
		Page         int      `json:"page"`
		X            float64  `json:"x"`
		Y            float64  `json:"y"`
		W            float64  `json:"w"`
		H            *float64 `json:"h"`
		Z            *float64 `json:"z"`
		CellW        *float64 `json:"cell_w,omitempty"`
		OptionID     *string  `json:"option_id,omitempty"`
	}

	var p payload
	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	a.AttachmentID = p.AttachmentID
	a.Page = p.Page
	a.X = p.X
	a.Y = p.Y
	a.W = p.W
	if p.H != nil {
		a.H = *p.H
	} else if p.Z != nil {
		a.H = *p.Z
	}

	return nil
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
	H     float64 `json:"h"`
}

// UnmarshalJSON accepts both `h` (current) and legacy `z` (height).
func (a *Annotations) UnmarshalJSON(b []byte) error {
	type payload struct {
		Type  string   `json:"type"`
		Value string   `json:"value"`
		Page  int      `json:"page"`
		X     float64  `json:"x"`
		Y     float64  `json:"y"`
		W     float64  `json:"w"`
		H     *float64 `json:"h"`
		Z     *float64 `json:"z"`
	}
	var p payload
	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}
	a.Type = p.Type
	a.Value = p.Value
	a.Page = p.Page
	a.X = p.X
	a.Y = p.Y
	a.W = p.W
	if p.H != nil {
		a.H = *p.H
	} else if p.Z != nil {
		a.H = *p.Z
	}
	return nil
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

