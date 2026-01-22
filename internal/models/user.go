package models

import "time"

type UserRole int

const (
	UserRoleUser      UserRole = 1 // Regular user with basic permissions
	UserRoleModerator UserRole = 2 // Moderator with extended permissions
	UserRoleAdmin     UserRole = 3 // Administrator with full access
)

// User is ...
type User struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Password        string    `json:"password"`
	PreferredLocale string    `json:"preferred_locale,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
