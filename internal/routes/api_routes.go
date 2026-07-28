package routes

import (
	"github.com/gofiber/fiber/v3"

	"github.com/shurco/gosign/internal/handlers/api"
	public "github.com/shurco/gosign/internal/handlers/public"
	"github.com/shurco/gosign/internal/middleware"
)

// APIHandlers contains all API handlers
type APIHandlers struct {
	Submissions    *api.SubmissionHandler
	Submitters     *api.SubmitterHandler
	SigningLinks   *api.SigningLinkHandler
	Templates      *api.TemplateHandler
	Webhooks       *api.WebhookHandler
	Settings       *api.SettingsHandler
	APIKeys        *api.APIKeyHandler
	Stats          *api.StatsHandler
	Events         *api.EventHandler
	Organizations  *api.OrganizationHandler
	Members        *api.MemberHandler
	Invitations    *api.InvitationHandler
	Users          *api.UserHandler
	I18n           *api.I18nHandler
	Branding       *api.BrandingHandler
	EmailTemplates *api.EmailTemplateHandler
	PublicSigning  *public.PublicSigningHandler
	Embed          *public.EmbedHandler
}

// APIRoutes configures all API routes.
// The whole HTTP API lives under the /v1 prefix so it can be served
// on a dedicated domain (api.example.com/v1/...). Only infrastructure
// endpoints (/health) and the /embed iframe page stay at the root.
func APIRoutes(c *fiber.App, handlers *APIHandlers) {
	v1 := c.Group("/v1")

	// Auth group (public routes)
	auth := v1.Group("/auth")

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
	verify := v1.Group("/verify")
	verify.Post("/pdf", public.VerifyPDF)

	// Public signer-facing API (no authentication)
	if handlers.PublicSigning != nil {
		publicAPI := v1.Group("/public")
		handlers.PublicSigning.RegisterRoutes(publicAPI)
	}

	// Embedded signing iframe (no authentication, root path:
	// customer-facing page, not a versioned API endpoint)
	if handlers.Embed != nil {
		handlers.Embed.RegisterRoutes(c)
	}

	// Invitations (public routes for accepting invitations)
	if handlers.Invitations != nil {
		invitations := v1.Group("/invitations")
		handlers.Invitations.RegisterRoutes(invitations)
	}

	// Protected resource groups. Middleware is attached per group prefix
	// (Group middleware in Fiber is a prefix-scoped Use), so public /v1
	// routes above stay open.
	protected := func(prefix string) fiber.Router {
		return v1.Group(prefix, middleware.Protected(), middleware.APIRateLimiter())
	}
	strict := func(prefix string) fiber.Router {
		return v1.Group(prefix, middleware.Protected(), middleware.StrictRateLimiter())
	}

	// Submissions API
	if handlers.Submissions != nil {
		handlers.Submissions.RegisterRoutes(protected("/submissions"))
	}

	// Direct signing links (protected; creates submission without email sending)
	if handlers.SigningLinks != nil {
		signingLinks := protected("/signing-links")
		signingLinks.Get("/", handlers.SigningLinks.List)
		signingLinks.Get("/:submission_id/document", handlers.SigningLinks.DownloadCompletedDocument)
		signingLinks.Get("/:submission_id", handlers.SigningLinks.Get)
		signingLinks.Post("/", handlers.SigningLinks.Create)
	}

	// Submitters API
	if handlers.Submitters != nil {
		handlers.Submitters.RegisterRoutes(protected("/submitters"))
	}

	// Templates API
	if handlers.Templates != nil {
		handlers.Templates.RegisterRoutes(protected("/templates"))
	}

	// Company API (formerly organizations)
	if handlers.Organizations != nil {
		company := protected("/company")

		// Members API (company members and invitations)
		// Register members routes FIRST to avoid route conflicts
		if handlers.Members != nil {
			handlers.Members.RegisterRoutes(company)
		}

		handlers.Organizations.RegisterRoutes(company)
	}

	// Settings API (with stricter rate limiting)
	settings := strict("/settings")
	if handlers.Settings != nil {
		handlers.Settings.RegisterRoutes(settings)
	}
	if handlers.APIKeys != nil {
		handlers.APIKeys.RegisterRoutes(settings.Group("/api"))
	}
	if handlers.Webhooks != nil {
		handlers.Webhooks.RegisterRoutes(settings.Group("/webhooks"))
	}
	if handlers.I18n != nil {
		handlers.I18n.RegisterRoutes(settings.Group("/i18n"))
	}
	if handlers.Branding != nil {
		handlers.Branding.RegisterRoutes(settings.Group("/branding"))
	}
	if handlers.EmailTemplates != nil {
		handlers.EmailTemplates.RegisterRoutes(settings.Group("/email/templates"))
	}

	// Stats API
	if handlers.Stats != nil {
		handlers.Stats.RegisterRoutes(protected("/stats"))
	}

	// Events API
	if handlers.Events != nil {
		handlers.Events.RegisterRoutes(protected("/events"))
	}

	// Users API
	if handlers.Users != nil {
		handlers.Users.RegisterRoutes(protected("/users"))
	}
}
