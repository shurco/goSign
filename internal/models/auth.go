package models

// SignIn represents user login request
type SignIn struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	Code     string `json:"code,omitempty"` // 2FA code if enabled
}

// SignUp represents user registration request
type SignUp struct {
	Email     string `json:"email" validate:"required,email,min=5,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}

// ForgotPassword represents password reset request
type ForgotPassword struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPassword represents password reset confirmation
type ResetPassword struct {
	Token       string `json:"token" validate:"required,min=32,max=255"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// VerifyEmail represents email verification request
type VerifyEmail struct {
	Token string `json:"token" validate:"required,min=32,max=255"`
}

// Enable2FA represents 2FA enable request
type Enable2FA struct {
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// Verify2FA represents 2FA verification request
type Verify2FA struct {
	Code string `json:"code" validate:"required,len=6"`
}

// Disable2FA represents 2FA disable request
type Disable2FA struct {
	Password string `json:"password" validate:"required,min=8,max=128"`
	Code     string `json:"code" validate:"required,len=6"`
}

// TwoFactorSetupResponse represents 2FA setup response
type TwoFactorSetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"` // Base64 encoded QR code image
}

// OAuthCallback represents OAuth callback data
type OAuthCallback struct {
	Provider string `json:"provider" validate:"required,oneof=google github"`
	Code     string `json:"code" validate:"required"`
	State    string `json:"state" validate:"required"`
}

