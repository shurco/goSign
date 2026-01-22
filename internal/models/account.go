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

// BrandingSettings extended configuration
type BrandingSettings struct {
	// Basic
	LogoURL         string `json:"logo_url,omitempty"`
	FaviconURL      string `json:"favicon_url,omitempty"`
	CompanyName     string `json:"company_name,omitempty"`
	
	// Colors
	PrimaryColor    string `json:"primary_color,omitempty"`    // #4F46E5
	SecondaryColor  string `json:"secondary_color,omitempty"`  // #6366F1
	AccentColor     string `json:"accent_color,omitempty"`     // #10B981
	BackgroundColor string `json:"background_color,omitempty"` // #FFFFFF
	TextColor       string `json:"text_color,omitempty"`       // #111827
	
	// Typography
	FontFamily      string `json:"font_family,omitempty"` // 'Inter', 'Roboto', etc.
	FontURL         string `json:"font_url,omitempty"`    // Google Fonts URL
	
	// Signing Page
	SigningPageTheme string `json:"signing_page_theme,omitempty"` // 'default', 'minimal', 'corporate'
	ShowPoweredBy    bool   `json:"show_powered_by"`              // default: true
	CustomCSS        string `json:"custom_css,omitempty"`         // Advanced customization
	
	// Email Templates
	EmailHeaderURL  string `json:"email_header_url,omitempty"`
	EmailFooterText string `json:"email_footer_text,omitempty"`
	EmailTheme      string `json:"email_theme,omitempty"` // 'default', 'minimal', 'corporate'
	
	// Custom Domain
	CustomDomain    string `json:"custom_domain,omitempty"`
	
	// Legal
	TermsURL        string `json:"terms_url,omitempty"`
	PrivacyURL      string `json:"privacy_url,omitempty"`
}

// BrandingAsset represents uploaded branding assets
type BrandingAsset struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	Type      string    `json:"type"` // logo, favicon, email_header, watermark
	FilePath  string    `json:"file_path"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CustomDomain represents custom domain configuration
type CustomDomain struct {
	ID                string     `json:"id"`
	AccountID         string     `json:"account_id"`
	Domain            string     `json:"domain"`
	Verified          bool       `json:"verified"`
	VerificationToken string     `json:"verification_token"`
	SSLEnabled        bool       `json:"ssl_enabled"`
	CreatedAt         time.Time  `json:"created_at"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
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
	// Extended fields (not in DB, populated from joins)
	Email      string `json:"email,omitempty"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	UserName   string `json:"user_name,omitempty"`
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

