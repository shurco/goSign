package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/security/password"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// SignUp handles user registration
func SignUp(c *fiber.Ctx) error {
	request := &models.SignUp{}
	if err := parseAndValidate(c, request); err != nil {
		return err
	}

	// Check if user already exists
	existingUser, _ := queries.DB.GetUserByEmail(context.Background(), request.Email)
	if existingUser != nil {
		return webutil.StatusBadRequest(c, "User with this email already exists")
	}

	// Hash password
	hashedPassword := password.GeneratePassword(request.Password)

	// Create user
	user, err := queries.DB.CreateUser(context.Background(), request.Email, hashedPassword, request.FirstName, request.LastName)
	if err != nil {
		logging.Log.Err(err).Send()
		return webutil.StatusInternalServerError(c)
	}

	// Create email verification token
	token, err := queries.DB.CreateEmailVerificationToken(context.Background(), user.ID, 24*time.Hour)
	if err != nil {
		logging.Log.Err(err).Msg("Failed to create email verification token")
		return webutil.StatusInternalServerError(c)
	}

	// TODO: Send verification email
	// verificationURL := fmt.Sprintf("http://localhost:8088/verify-email?token=%s", token)
	_ = token

	logging.Log.Info().Str("user_id", user.ID).Str("email", user.Email).Msg("User registered successfully")

	return webutil.StatusOK(c, "Registration successful. Please check your email to verify your account.", map[string]string{
		"user_id": user.ID,
	})
}

// VerifyEmail handles email verification
func VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return webutil.StatusBadRequest(c, "Token is required")
	}

	ctx := context.Background()

	// Validate token
	userID, err := queries.DB.ValidateEmailVerificationToken(ctx, token)
	if err != nil {
		return webutil.StatusBadRequest(c, "Invalid or expired token")
	}

	// Mark email as verified
	if err := queries.DB.MarkEmailAsVerified(ctx, userID); err != nil {
		logging.Log.Err(err).Send()
		return webutil.StatusInternalServerError(c)
	}

	// Mark token as used
	if err := queries.DB.MarkEmailVerificationTokenAsUsed(ctx, token); err != nil {
		logging.Log.Err(err).Msg("Failed to mark token as used")
	}

	logging.Log.Info().Str("user_id", userID).Msg("Email verified successfully")

	return webutil.StatusOK(c, "Email verified successfully", nil)
}

// SignIn handles user login with optional 2FA
func SignIn(c *fiber.Ctx) error {
	request := &models.SignIn{}
	if err := parseAndValidate(c, request); err != nil {
		return err
	}

	ctx := context.Background()

	// Get user from database
	user, err := queries.DB.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return webutil.StatusBadRequest(c, "Invalid email or password")
	}

	// Verify password
	if !password.ComparePasswords(user.Password, request.Password) {
		return webutil.StatusBadRequest(c, "Invalid email or password")
	}

	// Check if email is verified
	if !user.EmailVerified {
		return webutil.StatusBadRequest(c, "Please verify your email before logging in")
	}

	// Check 2FA if enabled
	if user.OTPEnabled {
		if request.Code == "" {
			return webutil.StatusOK(c, "2FA code required", map[string]any{
				"requires_2fa": true,
			})
		}

		// Validate TOTP code
		valid := totp.Validate(request.Code, user.OTPSecret)
		if !valid {
			return webutil.StatusBadRequest(c, "Invalid 2FA code")
		}
	}

	// Update login info
	if err := queries.DB.UpdateLoginInfo(ctx, user.ID, c.IP()); err != nil {
		logging.Log.Err(err).Msg("Failed to update login info")
	}

	// Create tokens
	accessToken, refreshToken, err := createAuthTokens(ctx, user)
	if err != nil {
		return webutil.StatusInternalServerError(c)
	}

	return webutil.StatusOK(c, "Login Successfully", map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

// SignOut logs out user and invalidates refresh token
func SignOut(c *fiber.Ctx) error {
	type SignOutRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req SignOutRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}

	invalidateRefreshToken(req.RefreshToken)
	return webutil.StatusOK(c, "Logout Successfully", nil)
}

// RefreshToken refreshes access token using refresh token
func RefreshToken(c *fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}

	if req.RefreshToken == "" {
		return webutil.StatusBadRequest(c, "refresh_token is required")
	}

	// Validate refresh token
	userID, err := middleware.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return webutil.StatusUnauthorized(c, "Invalid refresh token")
	}

	// Check if token exists in Redis
	refreshKey := fmt.Sprintf("refresh_token:%s", req.RefreshToken)
	storedUserID, err := redis.Conn.Get(refreshKey).Result()
	if err != nil || storedUserID != userID {
		return webutil.StatusUnauthorized(c, "Refresh token not found or revoked")
	}

	ctx := context.Background()

	// Get user from database
	user, err := queries.DB.GetUserByID(ctx, userID)
	if err != nil {
		logging.Log.Err(err).Send()
		return webutil.StatusInternalServerError(c)
	}

	// Delete old refresh token and create new ones
	invalidateRefreshToken(req.RefreshToken)
	
	newAccessToken, newRefreshToken, err := createAuthTokens(ctx, user)
	if err != nil {
		return webutil.StatusInternalServerError(c)
	}

	return webutil.StatusOK(c, "Token refreshed", map[string]string{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
	})
}

// ForgotPassword handles password reset request
func ForgotPassword(c *fiber.Ctx) error {
	request := &models.ForgotPassword{}
	if err := parseAndValidate(c, request); err != nil {
		return err
	}

	ctx := context.Background()

	// Get user
	user, err := queries.DB.GetUserByEmail(ctx, request.Email)
	if err != nil {
		// Don't reveal if user exists
		return webutil.StatusOK(c, "If the email exists, a password reset link has been sent", nil)
	}

	// Create password reset token
	token, err := queries.DB.CreatePasswordResetToken(ctx, user.ID, 1*time.Hour)
	if err != nil {
		logging.Log.Err(err).Msg("Failed to create password reset token")
		return webutil.StatusInternalServerError(c)
	}

	// TODO: Send reset email
	// resetURL := fmt.Sprintf("http://localhost:8088/reset-password?token=%s", token)
	_ = token

	logging.Log.Info().Str("user_id", user.ID).Msg("Password reset email sent")

	return webutil.StatusOK(c, "If the email exists, a password reset link has been sent", nil)
}

// ResetPassword handles password reset confirmation
func ResetPassword(c *fiber.Ctx) error {
	request := &models.ResetPassword{}
	if err := parseAndValidate(c, request); err != nil {
		return err
	}

	ctx := context.Background()

	// Validate token
	userID, err := queries.DB.ValidatePasswordResetToken(ctx, request.Token)
	if err != nil {
		return webutil.StatusBadRequest(c, "Invalid or expired token")
	}

	// Hash new password
	hashedPassword := password.GeneratePassword(request.NewPassword)

	// Update password
	if err := queries.DB.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		logging.Log.Err(err).Send()
		return webutil.StatusInternalServerError(c)
	}

	// Mark token as used
	if err := queries.DB.MarkPasswordResetTokenAsUsed(ctx, request.Token); err != nil {
		logging.Log.Err(err).Msg("Failed to mark token as used")
	}

	logging.Log.Info().Str("user_id", userID).Msg("Password reset successfully")

	return webutil.StatusOK(c, "Password reset successfully", nil)
}

// Enable2FA enables 2FA for authenticated user
func Enable2FA(c *fiber.Ctx) error {
	request := &models.Enable2FA{}
	if err := parseAndValidate(c, request); err != nil {
		return err
	}

	ctx := context.Background()
	userID := c.Locals("user_id").(string)

	// Get user
	user, err := queries.DB.GetUserByID(ctx, userID)
	if err != nil {
		logging.Log.Err(err).Send()
		return webutil.StatusInternalServerError(c)
	}

	// Verify password
	if !password.ComparePasswords(user.Password, request.Password) {
		return webutil.StatusBadRequest(c, "Invalid password")
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "goSign",
		AccountName: user.Email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		logging.Log.Err(err).Send()
		return webutil.StatusInternalServerError(c)
	}

	// Generate QR code
	qrImage, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		logging.Log.Err(err).Send()
		return webutil.StatusInternalServerError(c)
	}

	qrBase64 := base64.StdEncoding.EncodeToString(qrImage)

	logging.Log.Info().Str("user_id", userID).Msg("2FA setup initiated")

	return webutil.StatusOK(c, "Scan the QR code with your authenticator app", &models.TwoFactorSetupResponse{
		Secret: key.Secret(),
		QRCode: fmt.Sprintf("data:image/png;base64,%s", qrBase64),
	})
}

// Verify2FA verifies and activates 2FA
func Verify2FA(c *fiber.Ctx) error {
	request := &models.Verify2FA{}
	if err := parseAndValidate(c, request); err != nil {
		return err
	}

	ctx := context.Background()
	userID := c.Locals("user_id").(string)

	// Get secret from request body (sent from Enable2FA response)
	secret := c.FormValue("secret")
	if secret == "" {
		return webutil.StatusBadRequest(c, "Secret is required")
	}

	// Validate TOTP code
	valid := totp.Validate(request.Code, secret)
	if !valid {
		return webutil.StatusBadRequest(c, "Invalid 2FA code")
	}

	// Enable 2FA
	if err := queries.DB.Enable2FA(ctx, userID, secret); err != nil {
		logging.Log.Err(err).Send()
		return webutil.StatusInternalServerError(c)
	}

	logging.Log.Info().Str("user_id", userID).Msg("2FA enabled successfully")

	return webutil.StatusOK(c, "2FA enabled successfully", nil)
}

// Disable2FA disables 2FA for authenticated user
func Disable2FA(c *fiber.Ctx) error {
	request := &models.Disable2FA{}
	if err := parseAndValidate(c, request); err != nil {
		return err
	}

	ctx := context.Background()
	userID := c.Locals("user_id").(string)

	// Get user
	user, err := queries.DB.GetUserByID(ctx, userID)
	if err != nil {
		logging.Log.Err(err).Send()
		return webutil.StatusInternalServerError(c)
	}

	// Verify password
	if !password.ComparePasswords(user.Password, request.Password) {
		return webutil.StatusBadRequest(c, "Invalid password")
	}

	// Verify 2FA code
	valid := totp.Validate(request.Code, user.OTPSecret)
	if !valid {
		return webutil.StatusBadRequest(c, "Invalid 2FA code")
	}

	// Disable 2FA
	if err := queries.DB.Disable2FA(ctx, userID); err != nil {
		logging.Log.Err(err).Send()
		return webutil.StatusInternalServerError(c)
	}

	logging.Log.Info().Str("user_id", userID).Msg("2FA disabled successfully")

	return webutil.StatusOK(c, "2FA disabled successfully", nil)
}
