package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// SignIn represents user login request
type SignIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code,omitempty"` // 2FA code if enabled
}

// Validate validates SignIn request
func (v SignIn) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Email, validation.Required, is.Email),
		validation.Field(&v.Password, validation.Required, validation.Length(8, 128)),
	)
}

// SignUp represents user registration request
type SignUp struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Validate validates SignUp request
func (v SignUp) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Email, validation.Required, is.Email, validation.Length(5, 255)),
		validation.Field(&v.Password, validation.Required, validation.Length(8, 128)),
		validation.Field(&v.FirstName, validation.Required, validation.Length(1, 100)),
		validation.Field(&v.LastName, validation.Required, validation.Length(1, 100)),
	)
}

// ForgotPassword represents password reset request
type ForgotPassword struct {
	Email string `json:"email"`
}

// Validate validates ForgotPassword request
func (v ForgotPassword) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Email, validation.Required, is.Email),
	)
}

// ResetPassword represents password reset confirmation
type ResetPassword struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// Validate validates ResetPassword request
func (v ResetPassword) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Token, validation.Required, validation.Length(32, 255)),
		validation.Field(&v.NewPassword, validation.Required, validation.Length(8, 128)),
	)
}

// VerifyEmail represents email verification request
type VerifyEmail struct {
	Token string `json:"token"`
}

// Validate validates VerifyEmail request
func (v VerifyEmail) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Token, validation.Required, validation.Length(32, 255)),
	)
}

// Enable2FA represents 2FA enable request
type Enable2FA struct {
	Password string `json:"password"`
}

// Validate validates Enable2FA request
func (v Enable2FA) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Password, validation.Required, validation.Length(8, 128)),
	)
}

// Verify2FA represents 2FA verification request
type Verify2FA struct {
	Code string `json:"code"`
}

// Validate validates Verify2FA request
func (v Verify2FA) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Code, validation.Required, validation.Length(6, 6)),
	)
}

// Disable2FA represents 2FA disable request
type Disable2FA struct {
	Password string `json:"password"`
	Code     string `json:"code"`
}

// Validate validates Disable2FA request
func (v Disable2FA) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Password, validation.Required, validation.Length(8, 128)),
		validation.Field(&v.Code, validation.Required, validation.Length(6, 6)),
	)
}

// TwoFactorSetupResponse represents 2FA setup response
type TwoFactorSetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"` // Base64 encoded QR code image
}

// OAuthCallback represents OAuth callback data
type OAuthCallback struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
	State    string `json:"state"`
}

// Validate validates OAuthCallback request
func (v OAuthCallback) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Provider, validation.Required, validation.In("google", "github")),
		validation.Field(&v.Code, validation.Required),
		validation.Field(&v.State, validation.Required),
	)
}

