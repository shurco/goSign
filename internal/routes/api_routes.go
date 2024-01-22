package routes

import (
	"github.com/gofiber/fiber/v2"

	private "github.com/shurco/gosign/internal/handlers/private"
	public "github.com/shurco/gosign/internal/handlers/public"
	"github.com/shurco/gosign/internal/middleware"
)

// ApiRoutes is ...
func ApiRoutes(c *fiber.App) {
	auth := c.Group("/auth")
	auth.Post("/signin", public.SignIn)
	// sign.Post("/refresh", handlers.RefreshKey)
	auth.Post("/signout", middleware.Protected(), public.SignOut)

	// API group
	api := c.Group("/api")

	verify := api.Group("/verify")
	verify.Post("pdf", public.VerifyPDF)

	sign := api.Group("/sign")
	sign.Post("/", public.SignPDF)

	// test upload
	api.Post("/upload", public.Upload)

	// test template
	api.Get("/templates", private.Template)

	/*
		submissions := api.Group("/submissions", middleware.Protected())
		submissions.Get("/", handlers.Health)        // List all submissions
		submissions.Get("/{id}", handlers.Health)    // Get a submission
		submissions.Post("/", handlers.Health)       // Create a submission
		submissions.Post("/emails", handlers.Health) // Create submissions from emails
		submissions.Delete("/{id}", handlers.Health) // Archive a submission

		submitters := api.Group("/submitters", middleware.Protected())
		submitters.Get("/", handlers.Health)     // List all submitters
		submitters.Get("/{id}", handlers.Health) // Get a submitter
		submitters.Put("/{id}", handlers.Health) // Update a submitter

		templates := api.Group("/templates", middleware.Protected())
		templates.Get("/", handlers.Health)            // List all templates
		templates.Get("/{id}", handlers.Health)        // Get a template
		templates.Post("/docx", handlers.Health)       // Create a template from Word DOCX
		templates.Post("/html", handlers.Health)       // Create a template from HTML
		templates.Post("/pdf", handlers.Health)        // Create a template from existing PDF
		templates.Post("/{id}/clone", handlers.Health) // Clone a template
		templates.Put("/{id}", handlers.Health)        // Move a template to a different folder
		templates.Delete("/{id}", handlers.Health)     // Archive a template
	*/
}
