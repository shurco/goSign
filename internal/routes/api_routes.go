package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/shurco/gosign/internal/handlers/api"
	public "github.com/shurco/gosign/internal/handlers/public"
	"github.com/shurco/gosign/internal/middleware"
)

// APIHandlers contains all API handlers
type APIHandlers struct {
	Submissions     *api.SubmissionHandler
	Submitters      *api.SubmitterHandler
	SigningLinks    *api.SigningLinkHandler
	Templates       *api.TemplateHandler
	Webhooks        *api.WebhookHandler
	Settings        *api.SettingsHandler
	APIKeys         *api.APIKeyHandler
	Stats           *api.StatsHandler
	Events          *api.EventHandler
	Organizations   *api.OrganizationHandler
	Members         *api.MemberHandler
	Invitations     *api.InvitationHandler
	Users           *api.UserHandler
	I18n            *api.I18nHandler
	Branding        *api.BrandingHandler
	EmailTemplates  *api.EmailTemplateHandler
	PublicSigning   *public.PublicSigningHandler
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

	// Public signer-facing API (no authentication)
	if handlers.PublicSigning != nil {
		publicAPI := c.Group("/public")
		handlers.PublicSigning.RegisterRoutes(publicAPI)
	}

	// API v1 (protected routes with rate limiting)
	apiV1 := c.Group("/api/v1", middleware.Protected(), middleware.APIRateLimiter())

	// Invitations (public routes for accepting invitations)
	if handlers.Invitations != nil {
		invitations := c.Group("/api/v1/invitations")
		handlers.Invitations.RegisterRoutes(invitations)
	}

	// Submissions API
	if handlers.Submissions != nil {
		submissions := apiV1.Group("/submissions")
		handlers.Submissions.RegisterRoutes(submissions)
	}

	// Direct signing links (protected; creates submission without email sending)
	if handlers.SigningLinks != nil {
		signingLinks := apiV1.Group("/signing-links")
		signingLinks.Get("/", handlers.SigningLinks.List)
		signingLinks.Get("/:submission_id/document", handlers.SigningLinks.DownloadCompletedDocument)
		signingLinks.Get("/:submission_id", handlers.SigningLinks.Get)
		signingLinks.Post("/", handlers.SigningLinks.Create)
		signingLinks.Post("/submitters/:id/reset", handlers.SigningLinks.ResetSubmitter)
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

	// Organizations API
	if handlers.Organizations != nil {
		organizations := apiV1.Group("/organizations")
		
		// Members API (organization members and invitations)
		// Register members routes FIRST to avoid route conflicts
		// More specific routes should be registered before less specific ones
		if handlers.Members != nil {
			handlers.Members.RegisterRoutes(organizations)
		}
		
		// Then register organization routes
		handlers.Organizations.RegisterRoutes(organizations)
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

	// Users API
	if handlers.Users != nil {
		users := apiV1.Group("/users")
		handlers.Users.RegisterRoutes(users)
	}

	// I18n API
	if handlers.I18n != nil {
		i18n := apiV1.Group("/i18n")
		handlers.I18n.RegisterRoutes(i18n)
	}

	// Branding API
	if handlers.Branding != nil {
		branding := apiV1.Group("/branding")
		handlers.Branding.RegisterRoutes(branding)
	}

	// Email Templates API
	if handlers.EmailTemplates != nil {
		emailTemplates := apiV1.Group("/email-templates", middleware.StrictRateLimiter())
		handlers.EmailTemplates.RegisterRoutes(emailTemplates)
	}
}
