package models

import "time"

// AccountSettings contains account settings
type AccountSettings struct {
	Email    EmailSettings    `json:"email"`
	Storage  StorageSettings  `json:"storage"`
	Webhook  WebhookSettings  `json:"webhook"`
	Branding BrandingSettings `json:"branding"`
}

// EmailSettings email configuration
type EmailSettings struct {
	Provider  string `json:"provider"` // smtp, sendgrid, mailgun, ses
	SMTPHost  string `json:"smtp_host,omitempty"`
	SMTPPort  int    `json:"smtp_port,omitempty"`
	SMTPUser  string `json:"smtp_user,omitempty"`
	SMTPPass  string `json:"smtp_pass,omitempty"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
}

// StorageSettings storage configuration
type StorageSettings struct {
	Provider string            `json:"provider"` // local, s3, gcs, azure
	Bucket   string            `json:"bucket,omitempty"`
	Region   string            `json:"region,omitempty"`
	Config   map[string]string `json:"config,omitempty"` // additional settings
}

// WebhookSettings webhook configuration
type WebhookSettings struct {
	Enabled    bool `json:"enabled"`
	MaxRetries int  `json:"max_retries"`
	Timeout    int  `json:"timeout"` // in seconds
}

// BrandingSettings branding configuration
type BrandingSettings struct {
	LogoURL      string `json:"logo_url,omitempty"`
	PrimaryColor string `json:"primary_color,omitempty"`
	CompanyName  string `json:"company_name,omitempty"`
}

// AccountType represents the type of account
type AccountType string

const (
	AccountTypePersonal     AccountType = "personal"
	AccountTypeOrganization AccountType = "organization"
)

// OrganizationRole represents the role of a user in an organization
type OrganizationRole string

const (
	OrganizationRoleOwner  OrganizationRole = "owner"
	OrganizationRoleAdmin  OrganizationRole = "admin"
	OrganizationRoleMember OrganizationRole = "member"
	OrganizationRoleViewer OrganizationRole = "viewer"
)

// Account represents an account
type Account struct {
	ID        string           `json:"id"`
	Type      AccountType      `json:"type"`
	Name      string           `json:"name"`
	Timezone  string           `json:"timezone"`
	Locale    string           `json:"locale"`
	Settings  *AccountSettings `json:"settings,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// Organization represents an organization
type Organization struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organization_id"`
	UserID         string          `json:"user_id"`
	Role           OrganizationRole `json:"role"`
	JoinedAt       time.Time       `json:"joined_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// OrganizationInvitation represents an invitation to join an organization
type OrganizationInvitation struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organization_id"`
	Email          string          `json:"email"`
	Role           OrganizationRole `json:"role"`
	Token          string          `json:"token"`
	ExpiresAt      time.Time       `json:"expires_at"`
	InvitedByID    string          `json:"invited_by_id"`
	CreatedAt      time.Time       `json:"created_at"`
	AcceptedAt     *time.Time      `json:"accepted_at,omitempty"`
}

