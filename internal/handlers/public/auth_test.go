package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/storage/redis"
)

func setupTestApp() *fiber.App {
	app := fiber.New()
	return app
}

func TestRefreshTokenValidation(t *testing.T) {
	// Test JWT token creation and validation
	userID := "test-user-123"

	// Create refresh token
	refreshToken, err := middleware.CreateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	if refreshToken == "" {
		t.Error("CreateRefreshToken() returned empty token")
	}

	// Validate refresh token
	extractedUserID, err := middleware.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Errorf("ValidateRefreshToken() failed: %v", err)
	}

	if extractedUserID != userID {
		t.Errorf("ValidateRefreshToken() userID = %v, want %v", extractedUserID, userID)
	}
}

func TestRefreshTokenExpiration(t *testing.T) {
	// This test verifies that tokens are generated with proper expiration
	userID := "test-user-456"

	refreshToken, err := middleware.CreateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Token should be valid immediately
	_, err = middleware.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Errorf("ValidateRefreshToken() should be valid immediately, got error: %v", err)
	}

	// Note: Testing actual expiration would require time manipulation or waiting 7 days
	// In production, you might use a testing library that supports time mocking
}

func TestRefreshTokenInvalidToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "invalid format",
			token: "invalid-token-format",
		},
		{
			name:  "malformed JWT",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := middleware.ValidateRefreshToken(tt.token)
			if err == nil {
				t.Error("ValidateRefreshToken() should return error for invalid token")
			}
		})
	}
}

func TestAccessTokenCreation(t *testing.T) {
	// Mock user for testing
	user := &models.User{
		ID:    "user-123",
		Name:  "Test User",
		Email: "test@example.com",
	}

	accessToken, err := middleware.CreateToken(user)
	if err != nil {
		t.Fatalf("CreateToken() failed: %v", err)
	}

	if accessToken == "" {
		t.Error("CreateToken() returned empty token")
	}

	// Validate the access token
	extractedUser, err := middleware.ValidateToken(accessToken)
	if err != nil {
		t.Errorf("ValidateToken() failed: %v", err)
	}

	if extractedUser.ID != user.ID {
		t.Errorf("ValidateToken() user ID = %v, want %v", extractedUser.ID, user.ID)
	}
}

func TestRefreshTokenEndpoint(t *testing.T) {
	// Skip if Redis is not available
	if redis.Conn == nil {
		t.Skip("Redis not configured, skipping integration test")
	}

	app := setupTestApp()
	app.Post("/auth/refresh", RefreshToken)

	// Create a valid refresh token
	userID := "test-user-789"
	refreshToken, err := middleware.CreateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Store in Redis
	refreshKey := "refresh_token:" + refreshToken
	if err := redis.Conn.Set(refreshKey, userID, 7*24*time.Hour); err != nil {
		t.Fatalf("Failed to store refresh token in Redis: %v", err)
	}
	defer redis.Conn.Delete(refreshKey)

	// Test refresh endpoint
	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("RefreshToken endpoint status = %v, want %v", resp.StatusCode, http.StatusOK)
	}

	// Parse response
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that new tokens were returned
	data, ok := result["data"].(map[string]any)
	if !ok {
		t.Fatal("Response data is not a map")
	}

	if _, ok := data["access_token"]; !ok {
		t.Error("Response missing access_token")
	}

	if _, ok := data["refresh_token"]; !ok {
		t.Error("Response missing refresh_token")
	}
}

func TestRefreshTokenEndpointInvalidToken(t *testing.T) {
	app := setupTestApp()
	app.Post("/auth/refresh", RefreshToken)

	tests := []struct {
		name           string
		refreshToken   string
		wantStatusCode int
	}{
		{
			name:           "empty refresh token",
			refreshToken:   "",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid refresh token",
			refreshToken:   "invalid-token",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]string{
				"refresh_token": tt.refreshToken,
			}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() failed: %v", err)
			}

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("RefreshToken endpoint status = %v, want %v", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestTokenRotation(t *testing.T) {
	// Test that refresh token rotation works
	userID := "test-user-rotation"

	// Create first refresh token
	token1, err := middleware.CreateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to create first token: %v", err)
	}

	// Wait a moment to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Create second refresh token (simulating rotation)
	token2, err := middleware.CreateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to create second token: %v", err)
	}

	// Tokens should be different (different issued_at time)
	if token1 == token2 {
		t.Error("Token rotation should generate different tokens")
	}

	// Both should be valid and contain same user ID
	user1, err1 := middleware.ValidateRefreshToken(token1)
	user2, err2 := middleware.ValidateRefreshToken(token2)

	if err1 != nil || err2 != nil {
		t.Error("Both tokens should be valid")
	}

	if user1 != userID || user2 != userID {
		t.Error("Both tokens should contain the same user ID")
	}
}

