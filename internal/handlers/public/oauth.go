package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// OAuth configuration should be loaded from environment variables
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

// getGoogleOAuthConfig returns Google OAuth configuration
func getGoogleOAuthConfig() *OAuthConfig {
	// TODO: Load from environment variables or config
	return &OAuthConfig{
		ClientID:     "", // Set via env: GOOGLE_CLIENT_ID
		ClientSecret: "", // Set via env: GOOGLE_CLIENT_SECRET
		RedirectURL:  "http://localhost:8088/auth/oauth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
	}
}

// getGitHubOAuthConfig returns GitHub OAuth configuration
func getGitHubOAuthConfig() *OAuthConfig {
	// TODO: Load from environment variables or config
	return &OAuthConfig{
		ClientID:     "", // Set via env: GITHUB_CLIENT_ID
		ClientSecret: "", // Set via env: GITHUB_CLIENT_SECRET
		RedirectURL:  "http://localhost:8088/auth/oauth/github/callback",
		Scopes:       []string{"read:user", "user:email"},
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
	}
}

// GoogleLogin redirects to Google OAuth
func GoogleLogin(c *fiber.Ctx) error {
	config := getGoogleOAuthConfig()
	
	if config.ClientID == "" {
		return webutil.StatusInternalServerError(c)
	}

	// Generate state token for CSRF protection
	state, err := generateStateToken()
	if err != nil {
		return webutil.StatusInternalServerError(c)
	}

	// Store state in Redis with 10 minute expiration
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	if err := redis.Conn.Set(stateKey, "google", 10*time.Minute); err != nil {
		logging.Log.Err(err).Msg("Failed to store OAuth state")
	}

	// Build authorization URL
	params := url.Values{}
	params.Add("client_id", config.ClientID)
	params.Add("redirect_uri", config.RedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("state", state)
	
	authURL := fmt.Sprintf("%s?%s", config.AuthURL, params.Encode())
	
	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}

// GoogleCallback handles Google OAuth callback
func GoogleCallback(c *fiber.Ctx) error {
	log := logging.Log
	ctx := context.Background()

	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return webutil.StatusBadRequest(c, "Invalid OAuth callback")
	}

	// Verify state token
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	provider, err := redis.Conn.Get(stateKey).Result()
	if err != nil || provider != "google" {
		return webutil.StatusBadRequest(c, "Invalid state token")
	}

	// Delete used state token
	redis.Conn.Delete(stateKey)

	config := getGoogleOAuthConfig()

	// Exchange code for token
	tokenResp, err := exchangeCodeForToken(config, code)
	if err != nil {
		log.Err(err).Msg("Failed to exchange code for token")
		return webutil.StatusInternalServerError(c)
	}

	// Get user info from Google
	userInfo, err := getUserInfo(config, tokenResp.AccessToken)
	if err != nil {
		log.Err(err).Msg("Failed to get user info")
		return webutil.StatusInternalServerError(c)
	}

	// Extract user data
	email := userInfo["email"].(string)
	googleID := userInfo["id"].(string)
	firstName := ""
	lastName := ""
	
	if givenName, ok := userInfo["given_name"].(string); ok {
		firstName = givenName
	}
	if familyName, ok := userInfo["family_name"].(string); ok {
		lastName = familyName
	}

	// Check if user exists with this OAuth account
	user, err := queries.DB.GetUserByOAuthProvider(ctx, "google", googleID)
	if err != nil {
		// User doesn't exist, check if email exists
		existingUser, _ := queries.DB.GetUserByEmail(ctx, email)
		if existingUser != nil {
			// Link OAuth to existing account
			expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
			if err := queries.DB.CreateOrUpdateOAuthAccount(ctx, existingUser.ID, "google", googleID, tokenResp.AccessToken, tokenResp.RefreshToken, &expiresAt); err != nil {
				log.Err(err).Msg("Failed to link OAuth account")
				return webutil.StatusInternalServerError(c)
			}
			user = existingUser
		} else {
			// Create new user
			newUser, err := queries.DB.CreateUser(ctx, email, "", firstName, lastName)
			if err != nil {
				log.Err(err).Msg("Failed to create user")
				return webutil.StatusInternalServerError(c)
			}
			
			// Mark email as verified (trusted from OAuth provider)
			if err := queries.DB.MarkEmailAsVerified(ctx, newUser.ID); err != nil {
				log.Err(err).Msg("Failed to mark email as verified")
			}

			// Link OAuth account
			expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
			if err := queries.DB.CreateOrUpdateOAuthAccount(ctx, newUser.ID, "google", googleID, tokenResp.AccessToken, tokenResp.RefreshToken, &expiresAt); err != nil {
				log.Err(err).Msg("Failed to link OAuth account")
			}
			
			user = newUser
		}
	}

	// Create JWT tokens
	accessToken, refreshToken, err := createAuthTokens(ctx, user)
	if err != nil {
		return webutil.StatusInternalServerError(c)
	}

	// Update login info
	if err := queries.DB.UpdateLoginInfo(ctx, user.ID, c.IP()); err != nil {
		log.Err(err).Msg("Failed to update login info")
	}

	// Redirect to frontend with tokens
	// TODO: Configure frontend URL
	frontendURL := fmt.Sprintf("http://localhost:3000/auth/callback?access_token=%s&refresh_token=%s", accessToken, refreshToken)
	return c.Redirect(frontendURL, fiber.StatusTemporaryRedirect)
}

// GitHubLogin redirects to GitHub OAuth
func GitHubLogin(c *fiber.Ctx) error {
	config := getGitHubOAuthConfig()
	
	if config.ClientID == "" {
		return webutil.StatusInternalServerError(c)
	}

	// Generate state token for CSRF protection
	state, err := generateStateToken()
	if err != nil {
		return webutil.StatusInternalServerError(c)
	}

	// Store state in Redis with 10 minute expiration
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	if err := redis.Conn.Set(stateKey, "github", 10*time.Minute); err != nil {
		logging.Log.Err(err).Msg("Failed to store OAuth state")
	}

	// Build authorization URL
	params := url.Values{}
	params.Add("client_id", config.ClientID)
	params.Add("redirect_uri", config.RedirectURL)
	params.Add("scope", "read:user user:email")
	params.Add("state", state)
	
	authURL := fmt.Sprintf("%s?%s", config.AuthURL, params.Encode())
	
	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}

// GitHubCallback handles GitHub OAuth callback
func GitHubCallback(c *fiber.Ctx) error {
	log := logging.Log
	ctx := context.Background()

	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return webutil.StatusBadRequest(c, "Invalid OAuth callback")
	}

	// Verify state token
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	provider, err := redis.Conn.Get(stateKey).Result()
	if err != nil || provider != "github" {
		return webutil.StatusBadRequest(c, "Invalid state token")
	}

	// Delete used state token
	redis.Conn.Delete(stateKey)

	config := getGitHubOAuthConfig()

	// Exchange code for token
	tokenResp, err := exchangeCodeForToken(config, code)
	if err != nil {
		log.Err(err).Msg("Failed to exchange code for token")
		return webutil.StatusInternalServerError(c)
	}

	// Get user info from GitHub
	userInfo, err := getUserInfo(config, tokenResp.AccessToken)
	if err != nil {
		log.Err(err).Msg("Failed to get user info")
		return webutil.StatusInternalServerError(c)
	}

	// Extract user data
	githubID := fmt.Sprintf("%.0f", userInfo["id"].(float64))
	email := ""
	if userEmail, ok := userInfo["email"].(string); ok && userEmail != "" {
		email = userEmail
	} else {
		// Get primary email from GitHub API
		emails, err := getGitHubEmails(config, tokenResp.AccessToken)
		if err == nil && len(emails) > 0 {
			email = emails[0]
		}
	}

	if email == "" {
		return webutil.StatusBadRequest(c, "Email not provided by GitHub")
	}

	name := userInfo["name"].(string)
	firstName := name
	lastName := ""
	
	// Try to split name
	if name != "" {
		parts := splitName(name)
		if len(parts) > 0 {
			firstName = parts[0]
		}
		if len(parts) > 1 {
			lastName = parts[1]
		}
	}

	// Check if user exists with this OAuth account
	user, err := queries.DB.GetUserByOAuthProvider(ctx, "github", githubID)
	if err != nil {
		// User doesn't exist, check if email exists
		existingUser, _ := queries.DB.GetUserByEmail(ctx, email)
		if existingUser != nil {
			// Link OAuth to existing account
			if err := queries.DB.CreateOrUpdateOAuthAccount(ctx, existingUser.ID, "github", githubID, tokenResp.AccessToken, tokenResp.RefreshToken, nil); err != nil {
				log.Err(err).Msg("Failed to link OAuth account")
				return webutil.StatusInternalServerError(c)
			}
			user = existingUser
		} else {
			// Create new user
			newUser, err := queries.DB.CreateUser(ctx, email, "", firstName, lastName)
			if err != nil {
				log.Err(err).Msg("Failed to create user")
				return webutil.StatusInternalServerError(c)
			}
			
			// Mark email as verified (trusted from OAuth provider)
			if err := queries.DB.MarkEmailAsVerified(ctx, newUser.ID); err != nil {
				log.Err(err).Msg("Failed to mark email as verified")
			}

			// Link OAuth account
			if err := queries.DB.CreateOrUpdateOAuthAccount(ctx, newUser.ID, "github", githubID, tokenResp.AccessToken, tokenResp.RefreshToken, nil); err != nil {
				log.Err(err).Msg("Failed to link OAuth account")
			}
			
			user = newUser
		}
	}

	// Create JWT tokens
	accessToken, refreshToken, err := createAuthTokens(ctx, user)
	if err != nil {
		return webutil.StatusInternalServerError(c)
	}

	// Update login info
	if err := queries.DB.UpdateLoginInfo(ctx, user.ID, c.IP()); err != nil {
		log.Err(err).Msg("Failed to update login info")
	}

	// Redirect to frontend with tokens
	// TODO: Configure frontend URL
	frontendURL := fmt.Sprintf("http://localhost:3000/auth/callback?access_token=%s&refresh_token=%s", accessToken, refreshToken)
	return c.Redirect(frontendURL, fiber.StatusTemporaryRedirect)
}

// Helper functions

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func exchangeCodeForToken(config *OAuthConfig, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", config.RedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", config.TokenURL, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = data.Encode()
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func getUserInfo(config *OAuthConfig, accessToken string) (map[string]any, error) {
	req, err := http.NewRequest("GET", config.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get user info failed: %s", string(body))
	}

	var userInfo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func getGitHubEmails(config *OAuthConfig, accessToken string) ([]string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(emails))
	for _, e := range emails {
		if e.Primary {
			result = append([]string{e.Email}, result...)
		} else {
			result = append(result, e.Email)
		}
	}

	return result, nil
}

func generateStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func splitName(fullName string) []string {
	// Simple name splitting - can be improved
	parts := make([]string, 0, 2)
	for _, part := range []rune(fullName) {
		if part == ' ' {
			break
		}
	}
	
	// Basic split by first space
	for i, char := range fullName {
		if char == ' ' {
			parts = append(parts, fullName[:i])
			parts = append(parts, fullName[i+1:])
			break
		}
	}
	
	if len(parts) == 0 {
		parts = append(parts, fullName)
	}
	
	return parts
}

