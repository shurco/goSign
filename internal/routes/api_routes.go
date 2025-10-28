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
}

// ApiRoutes configures all API routes
func ApiRoutes(c *fiber.App, handlers *APIHandlers) {
	// Auth group (public routes)
	auth := c.Group("/auth")
	auth.Post("/signin", public.SignIn)
	// auth.Post("/refresh", public.RefreshKey) // TODO: implement
	auth.Post("/signout", middleware.Protected(), public.SignOut)

	// Public signing/verification (no authentication)
	verify := c.Group("/verify")
	verify.Post("/pdf", public.VerifyPDF)

	sign := c.Group("/sign")
	sign.Post("/", public.SignPDF)

	// Public upload (for testing)
	c.Post("/upload", public.Upload)

	// API v1 (protected routes with rate limiting)
	apiV1 := c.Group("/api/v1", middleware.Protected(), middleware.APIRateLimiter())

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

	// Legacy routes (backward compatibility)
	api := c.Group("/api")
	api.Get("/templates", private.Template) // TODO: migrate to v1
}
