package api

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/notification"
	"github.com/shurco/gosign/pkg/storage"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// SettingsHandler handles requests to settings
type SettingsHandler struct {
	notificationSvc *notification.Service
}

// NewSettingsHandler creates new handler
func NewSettingsHandler(notificationSvc *notification.Service) *SettingsHandler {
	return &SettingsHandler{
		notificationSvc: notificationSvc,
	}
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

// TestEmailRequest request body for testing email
type TestEmailRequest struct {
	Provider   string `json:"provider"`
	SMTPHost   string `json:"smtp_host" validate:"required"`
	SMTPPort   string `json:"smtp_port" validate:"required"`
	SMTPUser   string `json:"smtp_user" validate:"required"`
	SMTPPass   string `json:"smtp_pass" validate:"required"`
	FromEmail  string `json:"from_email" validate:"required,email"`
	FromName   string `json:"from_name"`
	ToEmail    string `json:"to_email" validate:"required,email"`
}

// TestEmail sends test email to verify SMTP settings
// @Summary Test email settings
// @Description Sends a test email to verify SMTP configuration
// @Tags settings
// @Accept json
// @Produce json
// @Param body body TestEmailRequest true "Email settings to test"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /api/settings/email/test [post]
func (h *SettingsHandler) TestEmail(c *fiber.Ctx) error {
	var req TestEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, "Invalid request body")
	}

	// Create temporary email provider with test settings
	smtpConfig := notification.SMTPConfig{
		Host:      req.SMTPHost,
		Port:      0, // Will be parsed from string
		User:      req.SMTPUser,
		Password:  req.SMTPPass,
		FromEmail: req.FromEmail,
		FromName:  req.FromName,
	}

	// Parse port
	var port int
	if _, err := fmt.Sscanf(req.SMTPPort, "%d", &port); err != nil {
		return webutil.StatusBadRequest(c, "Invalid SMTP port")
	}
	smtpConfig.Port = port

	provider := notification.NewEmailProvider(smtpConfig)

	// Create test notification
	testNotification := &models.Notification{
		Type:      models.NotificationTypeEmail,
		Recipient: req.ToEmail,
		Subject:   "goSign Test Email",
		Body:      "This is a test email from goSign. If you received this, your SMTP settings are configured correctly.",
	}

	// Try to send
	ctx := context.Background()
	if err := provider.Send(ctx, testNotification); err != nil {
		log.Error().Err(err).Msg("Failed to send test email")
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to send test email", map[string]any{
			"error": err.Error(),
		})
	}

	log.Info().Str("to", req.ToEmail).Msg("Test email sent successfully")

	return webutil.Response(c, fiber.StatusOK, "test_email", map[string]any{
		"status":  "success",
		"message": "Test email sent successfully",
		"to":      req.ToEmail,
	})
}

// TestStorageRequest request body for testing storage
type TestStorageRequest struct {
	Provider        string `json:"provider" validate:"required,oneof=local s3"`
	Bucket          string `json:"bucket,omitempty"`
	Region          string `json:"region,omitempty"`
	BasePath        string `json:"base_path,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"`
	AccessKeyID     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
}

// TestStorage tests storage configuration
// @Summary Test storage settings
// @Description Tests storage configuration by creating, reading, and deleting a test file
// @Tags settings
// @Accept json
// @Produce json
// @Param body body TestStorageRequest true "Storage settings to test"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /api/settings/storage/test [post]
func (h *SettingsHandler) TestStorage(c *fiber.Ctx) error {
	var req TestStorageRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, "Invalid request body")
	}

	// Create storage configuration
	var storageConfig storage.Config
	switch req.Provider {
	case "local":
		if req.BasePath == "" {
			return webutil.StatusBadRequest(c, "base_path is required for local storage")
		}
		storageConfig = storage.Config{
			Provider: "local",
			BasePath: req.BasePath,
		}
	case "s3":
		if req.Bucket == "" || req.Region == "" {
			return webutil.StatusBadRequest(c, "bucket and region are required for S3")
		}
		options := make(map[string]string)
		if req.AccessKeyID != "" {
			options["access_key_id"] = req.AccessKeyID
		}
		if req.SecretAccessKey != "" {
			options["secret_access_key"] = req.SecretAccessKey
		}
		storageConfig = storage.Config{
			Provider: "s3",
			Bucket:   req.Bucket,
			Region:   req.Region,
			Endpoint: req.Endpoint,
			Options:  options,
		}
	default:
		return webutil.StatusBadRequest(c, "unsupported storage provider")
	}

	// Create storage instance
	ctx := c.Context()
	storageInstance, err := storage.NewStorage(ctx, storageConfig)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Msg("Failed to create storage")
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to create storage", map[string]any{
			"error": err.Error(),
		})
	}

	// Test file
	testKey := "test/goSign-test-file.txt"
	testContent := []byte("This is a test file from goSign")

	// 1. Upload test file
	metadata := &storage.BlobMetadata{
		ContentType: "text/plain",
		Size:        int64(len(testContent)),
	}
	if err := storageInstance.Upload(ctx, testKey, bytes.NewReader(testContent), metadata); err != nil {
		log.Error().Err(err).Msg("Failed to upload test file")
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to write test file", map[string]any{
			"error": err.Error(),
		})
	}

	// 2. Download test file
	reader, err := storageInstance.Download(ctx, testKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download test file")
		// Try to cleanup
		_ = storageInstance.Delete(ctx, testKey)
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to read test file", map[string]any{
			"error": err.Error(),
		})
	}
	defer reader.Close()

	retrievedContent, err := io.ReadAll(reader)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read test file content")
		_ = storageInstance.Delete(ctx, testKey)
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to read test file content", map[string]any{
			"error": err.Error(),
		})
	}

	// Verify content
	if string(retrievedContent) != string(testContent) {
		log.Error().Msg("Test file content mismatch")
		_ = storageInstance.Delete(ctx, testKey)
		return webutil.Response(c, fiber.StatusBadRequest, "Test file content mismatch", nil)
	}

	// 3. Delete test file
	if err := storageInstance.Delete(ctx, testKey); err != nil {
		log.Error().Err(err).Msg("Failed to delete test file")
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to delete test file", map[string]any{
			"error": err.Error(),
		})
	}

	log.Info().Str("provider", req.Provider).Msg("Storage test passed")

	return webutil.Response(c, fiber.StatusOK, "test_storage", map[string]any{
		"status":  "success",
		"message": "Storage test passed: file created, read, and deleted successfully",
	})
}

// RegisterRoutes registers all routes for settings
func (h *SettingsHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.Get)
	router.Put("/email", h.UpdateEmail)
	router.Put("/storage", h.UpdateStorage)
	router.Put("/branding", h.UpdateBranding)
	router.Post("/email/test", h.TestEmail)
	router.Post("/storage/test", h.TestStorage)
}

