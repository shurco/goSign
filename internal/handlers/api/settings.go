package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// SettingsHandler handles requests to settings
type SettingsHandler struct {
	// TODO: add settingsRepository for saving to database
}

// NewSettingsHandler creates new handler
func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

// Get returns current settings
// @Summary Get settings
// @Description Returns current application settings
// @Tags settings
// @Produce json
// @Success 200 {object} config.Settings
// @Router /api/settings [get]
func (h *SettingsHandler) Get(c *fiber.Ctx) error {
	cfg := config.Data()
	
		// Hide sensitive data
	safSettings := map[string]any{
		"email": map[string]any{
			"provider":   cfg.Settings.Email["provider"],
			"smtp_host":  cfg.Settings.Email["smtp_host"],
			"smtp_port":  cfg.Settings.Email["smtp_port"],
			"from_email": cfg.Settings.Email["from_email"],
			"from_name":  cfg.Settings.Email["from_name"],
			// hide smtp_pass
		},
		"storage": map[string]any{
			"provider":  cfg.Settings.Storage["provider"],
			"bucket":    cfg.Settings.Storage["bucket"],
			"region":    cfg.Settings.Storage["region"],
			"base_path": cfg.Settings.Storage["base_path"],
			// hide access_key_id and secret_access_key
		},
		"webhook": cfg.Settings.Webhook,
		"features": cfg.Settings.Features,
	}

	return webutil.Response(c, fiber.StatusOK, "settings", safSettings)
}

// UpdateEmailRequest request body for updating email settings
type UpdateEmailRequest struct {
	Provider   string `json:"provider"`
	SMTPHost   string `json:"smtp_host"`
	SMTPPort   string `json:"smtp_port"`
	SMTPUser   string `json:"smtp_user"`
	SMTPPass   string `json:"smtp_pass,omitempty"` // optional - only if changing
	FromEmail  string `json:"from_email"`
	FromName   string `json:"from_name"`
}

// UpdateEmail updates email settings
// @Summary Update email settings
// @Description Updates email/SMTP configuration
// @Tags settings
// @Accept json
// @Produce json
// @Param body body UpdateEmailRequest true "Email settings"
// @Success 200 {object} map[string]any
// @Router /api/settings/email [put]
func (h *SettingsHandler) UpdateEmail(c *fiber.Ctx) error {
	var req UpdateEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, "Invalid request body")
	}

	// TODO: Validate settings
	// TODO: Save to DB (account.settings_jsonb)
	// TODO: Update in memory (config.Data())

	log.Info().Msg("Email settings updated")
	
	return webutil.Response(c, fiber.StatusOK, "email_settings", map[string]any{
		"status": "updated",
	})
}

// UpdateStorageRequest request body for updating storage settings
type UpdateStorageRequest struct {
	Provider        string `json:"provider" validate:"required,oneof=local s3 gcs azure"`
	Bucket          string `json:"bucket,omitempty"`
	Region          string `json:"region,omitempty"`
	BasePath        string `json:"base_path,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"`
	AccessKeyID     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
}

// UpdateStorage updates storage settings
// @Summary Update storage settings
// @Description Updates storage configuration (local, S3, GCS, Azure)
// @Tags settings
// @Accept json
// @Produce json
// @Param body body UpdateStorageRequest true "Storage settings"
// @Success 200 {object} map[string]any
// @Router /api/settings/storage [put]
func (h *SettingsHandler) UpdateStorage(c *fiber.Ctx) error {
	var req UpdateStorageRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, "Invalid request body")
	}

	// TODO: Validation based on provider
	// TODO: Save to DB
	// TODO: Update in memory

	log.Info().Str("provider", req.Provider).Msg("Storage settings updated")
	
	return webutil.Response(c, fiber.StatusOK, "storage_settings", map[string]any{
		"status": "updated",
	})
}

// UpdateBrandingRequest request body for updating branding settings
type UpdateBrandingRequest struct {
	CompanyName string `json:"company_name"`
	LogoURL     string `json:"logo_url"`
	PrimaryColor string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
}

// UpdateBranding updates branding settings
// @Summary Update branding settings
// @Description Updates company branding configuration
// @Tags settings
// @Accept json
// @Produce json
// @Param body body UpdateBrandingRequest true "Branding settings"
// @Success 200 {object} map[string]any
// @Router /api/settings/branding [put]
func (h *SettingsHandler) UpdateBranding(c *fiber.Ctx) error {
	var req UpdateBrandingRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, "Invalid request body")
	}

	// TODO: Save to DB (account.settings_jsonb)
	
	log.Info().Msg("Branding settings updated")
	
	return webutil.Response(c, fiber.StatusOK, "branding_settings", map[string]any{
		"status": "updated",
	})
}

// RegisterRoutes registers all routes for settings
func (h *SettingsHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.Get)
	router.Put("/email", h.UpdateEmail)
	router.Put("/storage", h.UpdateStorage)
	router.Put("/branding", h.UpdateBranding)
}

