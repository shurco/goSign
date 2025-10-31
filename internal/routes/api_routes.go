package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/shurco/gosign/internal/handlers/api"
	private "github.com/shurco/gosign/internal/handlers/private"
	public "github.com/shurco/gosign/internal/handlers/public"
	"github.com/shurco/gosign/internal/middleware"
)

// APIHandlers contains all API handlers
type APIHandlers struct {
	Submissions *api.SubmissionHandler
	Submitters  *api.SubmitterHandler
	Templates   *api.TemplateHandler
	Webhooks    *api.WebhookHandler
	Settings    *api.SettingsHandler
	APIKeys     *api.APIKeyHandler
	Stats       *api.StatsHandler
	Events      *api.EventHandler
}

// ApiRoutes configures all API routes
func ApiRoutes(c *fiber.App, handlers *APIHandlers) {
	// Auth group (public routes)
	auth := c.Group("/auth")
	
	// Basic authentication
	auth.Post("/signup", public.SignUp)
	auth.Post("/signin", public.SignIn)
	auth.Post("/refresh", public.RefreshToken)
	auth.Post("/signout", middleware.Protected(), public.SignOut)
	
	// Email verification
	auth.Get("/verify-email", public.VerifyEmail)
	
	// Password management
	password := auth.Group("/password")
	password.Post("/forgot", public.ForgotPassword)
	password.Post("/reset", public.ResetPassword)
	
	// Two-factor authentication (protected routes)
	twoFactor := auth.Group("/2fa", middleware.Protected())
	twoFactor.Post("/enable", public.Enable2FA)
	twoFactor.Post("/verify", public.Verify2FA)
	twoFactor.Post("/disable", public.Disable2FA)
	
	// OAuth routes
	oauth := auth.Group("/oauth")
	oauth.Get("/google", public.GoogleLogin)
	oauth.Get("/google/callback", public.GoogleCallback)
	oauth.Get("/github", public.GitHubLogin)
	oauth.Get("/github/callback", public.GitHubCallback)

	// Public signing/verification (no authentication)
	verify := c.Group("/verify")
	verify.Post("/pdf", public.VerifyPDF)

	sign := c.Group("/sign")
	sign.Post("/", public.SignPDF)

	// API v1 (protected routes with rate limiting)
	apiV1 := c.Group("/api/v1", middleware.Protected(), middleware.APIRateLimiter())
	
	// Upload endpoint (protected)
	apiV1.Post("/upload", public.Upload)

	// Submissions API
	if handlers.Submissions != nil {
		submissions := apiV1.Group("/submissions")
		handlers.Submissions.RegisterRoutes(submissions)
	}

	// Submitters API
	if handlers.Submitters != nil {
		submitters := apiV1.Group("/submitters")
		handlers.Submitters.RegisterRoutes(submitters)
	}

	// Templates API
	if handlers.Templates != nil {
		templates := apiV1.Group("/templates")
		handlers.Templates.RegisterRoutes(templates)
	}

	// Webhooks API
	if handlers.Webhooks != nil {
		webhooks := apiV1.Group("/webhooks")
		handlers.Webhooks.RegisterRoutes(webhooks)
	}

	// Settings API (with stricter rate limiting)
	if handlers.Settings != nil {
		settings := apiV1.Group("/settings", middleware.StrictRateLimiter())
		handlers.Settings.RegisterRoutes(settings)
	}

	// API Keys management
	if handlers.APIKeys != nil {
		apikeys := apiV1.Group("/apikeys", middleware.StrictRateLimiter())
		handlers.APIKeys.RegisterRoutes(apikeys)
	}

	// Stats API
	if handlers.Stats != nil {
		stats := apiV1.Group("/stats")
		handlers.Stats.RegisterRoutes(stats)
	}

	// Events API
	if handlers.Events != nil {
		events := apiV1.Group("/events")
		handlers.Events.RegisterRoutes(events)
	}

	// Legacy routes (backward compatibility)
	api := c.Group("/api")
	api.Get("/templates", private.Template) // TODO: migrate to v1
}
