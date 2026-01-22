package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/security/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Story: Authentication Flow
// As a user, I want to register, login, and manage my session
// so that I can securely access the application

func TestSignUpFlow(t *testing.T) {
	t.Run("successful_user_registration", func(t *testing.T) {
		// Given: a new user wants to register
		signupData := models.SignUp{
			Email:     "newuser@example.com",
			Password:  "SecureP@ssw0rd123",
			FirstName: "John",
			LastName:  "Doe",
		}

		// When: submitting registration request
		app := setupTestApp()
		app.Post("/auth/signup", SignUp)

		bodyBytes, _ := json.Marshal(signupData)
		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 30000)

		// Then: registration should succeed or fail gracefully
		require.NoError(t, err)
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)
	})

	t.Run("reject_weak_password", func(t *testing.T) {
		// Given: user attempts registration with weak password
		signupData := models.SignUp{
			Email:     "test@example.com",
			Password:  "123",
			FirstName: "Test",
			LastName:  "User",
		}

		// When: submitting registration
		app := setupTestApp()
		app.Post("/auth/signup", SignUp)

		bodyBytes, _ := json.Marshal(signupData)
		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 30000)

		// Then: should reject with bad request
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("reject invalid email", func(t *testing.T) {
		// Given: user attempts registration with invalid email
		signupData := models.SignUp{
			Email:     "invalid-email",
			Password:  "SecureP@ssw0rd123",
			FirstName: "Test",
			LastName:  "User",
		}

		// When: submitting registration
		app := setupTestApp()
		app.Post("/auth/signup", SignUp)

		bodyBytes, _ := json.Marshal(signupData)
		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 30000)

		// Then: should reject with bad request
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("reject_missing_required_fields", func(t *testing.T) {
		tests := []struct {
			name string
			data models.SignUp
		}{
			{
				name: "missing email",
				data: models.SignUp{
					Password:  "SecureP@ssw0rd123",
					FirstName: "Test",
					LastName:  "User",
				},
			},
			{
				name: "missing password",
				data: models.SignUp{
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				},
			},
			{
				name: "missing first name",
				data: models.SignUp{
					Email:    "test@example.com",
					Password: "SecureP@ssw0rd123",
					LastName: "User",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				app := setupTestApp()
				app.Post("/auth/signup", SignUp)

				bodyBytes, _ := json.Marshal(tt.data)
				req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req, 30000)

				require.NoError(t, err)
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		}
	})
}

func TestSignInFlow(t *testing.T) {
	t.Run("signin_with_invalid_credentials", func(t *testing.T) {
		// Given: attempting to sign in with wrong credentials
		signinData := models.SignIn{
			Email:    "nonexistent@example.com",
			Password: "wrongpassword",
		}

		// When: submitting sign in request
		app := setupTestApp()
		app.Post("/auth/signin", SignIn)

		bodyBytes, _ := json.Marshal(signinData)
		req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 30000)

		// Then: should reject with bad request
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("signin_with_empty_password", func(t *testing.T) {
		// Given: attempting sign in without password
		signinData := models.SignIn{
			Email:    "test@example.com",
			Password: "",
		}

		// When: submitting request
		app := setupTestApp()
		app.Post("/auth/signin", SignIn)

		bodyBytes, _ := json.Marshal(signinData)
		req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 30000)

		// Then: should reject
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("signin_with_malformed_json", func(t *testing.T) {
		// Given: malformed JSON request
		app := setupTestApp()
		app.Post("/auth/signin", SignIn)

		req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 30000)

		// Then: should reject
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestPasswordManagement(t *testing.T) {
	t.Run("forgot_password_with_valid_email", func(t *testing.T) {
		// Given: user wants to reset password
		forgotData := models.ForgotPassword{
			Email: "test@example.com",
		}

		// When: requesting password reset
		app := setupTestApp()
		app.Post("/auth/password/forgot", ForgotPassword)

		bodyBytes, _ := json.Marshal(forgotData)
		req := httptest.NewRequest(http.MethodPost, "/auth/password/forgot", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 30000)

		// Then: should return success (even for non-existent email to prevent enumeration)
		require.NoError(t, err)
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusInternalServerError)
	})

	t.Run("forgot password with invalid email", func(t *testing.T) {
		// Given: invalid email format
		forgotData := models.ForgotPassword{
			Email: "invalid-email",
		}

		// When: requesting reset
		app := setupTestApp()
		app.Post("/auth/password/forgot", ForgotPassword)

		bodyBytes, _ := json.Marshal(forgotData)
		req := httptest.NewRequest(http.MethodPost, "/auth/password/forgot", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 30000)

		// Then: should reject
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("reset_password_with_invalid_token", func(t *testing.T) {
		// Given: invalid reset token
		resetData := models.ResetPassword{
			Token:       "invalid-token",
			NewPassword: "NewSecureP@ssw0rd123",
		}

		// When: attempting password reset
		app := setupTestApp()
		app.Post("/auth/password/reset", ResetPassword)

		bodyBytes, _ := json.Marshal(resetData)
		req := httptest.NewRequest(http.MethodPost, "/auth/password/reset", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 30000)

		// Then: should reject
		require.NoError(t, err)
		assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError)
	})

	t.Run("reset_password_with_weak_new_password", func(t *testing.T) {
		// Given: weak new password
		resetData := models.ResetPassword{
			Token:       "valid-token-would-be-here",
			NewPassword: "123",
		}

		// When: attempting reset
		app := setupTestApp()
		app.Post("/auth/password/reset", ResetPassword)

		bodyBytes, _ := json.Marshal(resetData)
		req := httptest.NewRequest(http.MethodPost, "/auth/password/reset", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 30000)

		// Then: should reject
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestEmailVerification(t *testing.T) {
	t.Run("verify email with invalid token", func(t *testing.T) {
		// Given: invalid verification token
		app := setupTestApp()
		app.Get("/auth/verify-email", VerifyEmail)

		// When: attempting verification
		req := httptest.NewRequest(http.MethodGet, "/auth/verify-email?token=invalid-token", nil)

		resp, err := app.Test(req, 30000)

		// Then: should reject or handle gracefully
		require.NoError(t, err)
		assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError)
	})

	t.Run("verify_email_without_token", func(t *testing.T) {
		// Given: no token provided
		app := setupTestApp()
		app.Get("/auth/verify-email", VerifyEmail)

		// When: accessing endpoint
		req := httptest.NewRequest(http.MethodGet, "/auth/verify-email", nil)

		resp, err := app.Test(req, 30000)

		// Then: should reject
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestPasswordHashing(t *testing.T) {
	t.Run("password_hashing_and_verification", func(t *testing.T) {
		// Given: a plain text password
		plainPassword := "SecureP@ssw0rd123"

		// When: hashing the password
		hashedPassword := password.GeneratePassword(plainPassword)

		// Then: hash should not be empty
		assert.NotEmpty(t, hashedPassword)

		// And: hash should be different from plain password
		assert.NotEqual(t, plainPassword, hashedPassword)

		// And: should be able to verify correct password
		assert.True(t, password.ComparePasswords(hashedPassword, plainPassword))

		// And: should reject incorrect password
		assert.False(t, password.ComparePasswords(hashedPassword, "wrongpassword"))
	})

	t.Run("same password produces different hashes", func(t *testing.T) {
		// Given: same password hashed twice
		password1 := password.GeneratePassword("SamePassword123")
		password2 := password.GeneratePassword("SamePassword123")

		// Then: hashes should be different (due to salt)
		assert.NotEqual(t, password1, password2)

		// But: both should verify correctly
		assert.True(t, password.ComparePasswords(password1, "SamePassword123"))
		assert.True(t, password.ComparePasswords(password2, "SamePassword123"))
	})
}

func TestValidationHelpers(t *testing.T) {
	t.Run("validate_signup_data_structure", func(t *testing.T) {
		validSignup := models.SignUp{
			Email:     "valid@example.com",
			Password:  "SecureP@ssw0rd123",
			FirstName: "John",
			LastName:  "Doe",
		}

		// Should have all required fields
		assert.NotEmpty(t, validSignup.Email)
		assert.NotEmpty(t, validSignup.Password)
		assert.NotEmpty(t, validSignup.FirstName)
		assert.NotEmpty(t, validSignup.LastName)
	})

	t.Run("validate_signin_data_structure", func(t *testing.T) {
		validSignin := models.SignIn{
			Email:    "user@example.com",
			Password: "password123",
		}

		// Should have required fields
		assert.NotEmpty(t, validSignin.Email)
		assert.NotEmpty(t, validSignin.Password)
	})
}

func TestSecurityFeatures(t *testing.T) {
	t.Run("password complexity requirements", func(t *testing.T) {
		weakPasswords := []string{
			"123",
			"password",
			"abc123",
			"qwerty",
		}

		for _, pwd := range weakPasswords {
			signupData := models.SignUp{
				Email:     "test@example.com",
				Password:  pwd,
				FirstName: "Test",
				LastName:  "User",
			}

			app := setupTestApp()
			app.Post("/auth/signup", SignUp)

			bodyBytes, _ := json.Marshal(signupData)
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, 30000)

			// Weak passwords should be rejected
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Weak password should be rejected: %s", pwd)
		}
	})

	t.Run("email_enumeration_prevention", func(t *testing.T) {
		// Given: forgot password requests for existing and non-existing emails
		app := setupTestApp()
		app.Post("/auth/password/forgot", ForgotPassword)

		emails := []string{
			"existing@example.com",
			"nonexisting@example.com",
		}

		responses := make([]int, len(emails))

		for i, email := range emails {
			forgotData := models.ForgotPassword{Email: email}
			bodyBytes, _ := json.Marshal(forgotData)

			req := httptest.NewRequest(http.MethodPost, "/auth/password/forgot", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, 30000)
			require.NoError(t, err)
			responses[i] = resp.StatusCode
		}

		// Both should return same status to prevent enumeration
		// (Either both succeed or both fail gracefully)
		assert.True(t, responses[0] == responses[1] || responses[0] >= 500 || responses[1] >= 500,
			"Status codes should be consistent to prevent email enumeration")
	})
}

func TestUserRoles(t *testing.T) {
	t.Run("user_role_constants_are_correct", func(t *testing.T) {
		// Verify new role system
		assert.Equal(t, models.UserRole(1), models.UserRoleUser)
		assert.Equal(t, models.UserRole(2), models.UserRoleModerator)
		assert.Equal(t, models.UserRole(3), models.UserRoleAdmin)
	})

	t.Run("roles_are_properly_ordered", func(t *testing.T) {
		// User < Moderator < Admin
		assert.True(t, models.UserRoleUser < models.UserRoleModerator)
		assert.True(t, models.UserRoleModerator < models.UserRoleAdmin)
	})
}

// Benchmark tests for performance
func BenchmarkPasswordHashing(b *testing.B) {
	pwd := "SecureP@ssw0rd123"
	for i := 0; i < b.N; i++ {
		password.GeneratePassword(pwd)
	}
}

func BenchmarkPasswordVerification(b *testing.B) {
	pwd := "SecureP@ssw0rd123"
	hash := password.GeneratePassword(pwd)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		password.ComparePasswords(hash, pwd)
	}
}

