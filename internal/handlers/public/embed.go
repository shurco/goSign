package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// EmbedHandler handles requests for embedded signing interface
type EmbedHandler struct{}

// NewEmbedHandler creates new embed handler
func NewEmbedHandler() *EmbedHandler {
	return &EmbedHandler{}
}

// GetEmbedPage returns HTML page for embedding in iframe
// @Summary Get embeddable signing page
// @Description Returns HTML page that can be embedded in iframe
// @Tags Embed
// @Produce html
// @Param slug path string true "Submitter slug"
// @Success 200 {string} string "HTML page"
// @Failure 404 {object} map[string]any "Submission not found"
// @Router /embed/{slug} [get]
func (h *EmbedHandler) GetEmbedPage(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return webutil.StatusNotFound(c)
	}

	// Render Vue app with required route
	// In production this will proxy to existing Sign UI
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sign Document</title>
    <style>
        body, html {
            margin: 0;
            padding: 0;
            height: 100%;
            overflow: hidden;
        }
        iframe {
            width: 100%;
            height: 100%;
            border: none;
        }
    </style>
</head>
<body>
    <iframe src="/s/` + slug + `" allow="camera;microphone"></iframe>
    <script>
        // Post message for communication with parent window
        window.addEventListener('message', function(event) {
            // Handle messages from parent window
            console.log('Received message:', event.data);
        });

        // Send events to parent window
        function notifyParent(event, data) {
            if (window.parent !== window) {
                window.parent.postMessage({
                    source: 'gosign-embed',
                    event: event,
                    data: data
                }, '*');
            }
        }

        // Listen for events from Sign UI
        window.addEventListener('message', function(event) {
            if (event.data && event.data.source === 'gosign-sign') {
                notifyParent(event.data.event, event.data.data);
            }
        });

        // Notify about readiness
        notifyParent('ready', { slug: '` + slug + `' });
    </script>
</body>
</html>`

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// GetEmbedConfig returns configuration for embedding
// @Summary Get embed configuration
// @Description Returns configuration and URLs for embedding
// @Tags Embed
// @Produce json
// @Param slug path string true "Submitter slug"
// @Success 200 {object} map[string]any "Embed configuration"
// @Failure 404 {object} map[string]any "Submission not found"
// @Router /embed/{slug}/config [get]
func (h *EmbedHandler) GetEmbedConfig(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return webutil.StatusNotFound(c)
	}

	// TODO: Validate slug in database

	config := map[string]any{
		"slug":      slug,
		"embed_url": "/embed/" + slug,
		"direct_url": "/s/" + slug,
		"events": []string{
			"ready",
			"opened",
			"field_filled",
			"completed",
			"declined",
			"error",
		},
	}

	return webutil.Response(c, fiber.StatusOK, "Embed configuration retrieved", config)
}

// RegisterRoutes registers routes for embed
func (h *EmbedHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/embed/:slug", h.GetEmbedPage)
	router.Get("/embed/:slug/config", h.GetEmbedConfig)
}

