package models

import "time"

type UserRole int

const (
	UserRoleViewer UserRole = iota
	UserRoleSigner
	UserRoleAdmin
)

type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt time.Time
}
