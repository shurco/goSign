package models

import "time"

// APIKey represents an API key for authentication
type APIKey struct {
	ID         string     `json:"id" db:"id"`
	AccountID  string     `json:"account_id" db:"account_id"`
	Name       string     `json:"name" db:"name"`
	KeyHash    string     `json:"-" db:"key_hash"` // not exported to JSON
	Enabled    bool       `json:"enabled" db:"enabled"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// APIKeyWithPlainKey is used only when creating a new key
type APIKeyWithPlainKey struct {
	APIKey
	PlainKey string `json:"api_key"` // shown only once during creation
}

