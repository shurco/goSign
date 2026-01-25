package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/appdir"
	"github.com/shurco/gosign/pkg/geolocation"
	"github.com/shurco/gosign/pkg/notification"
	"github.com/shurco/gosign/pkg/storage"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// SettingsHandler handles requests to settings
type SettingsHandler struct {
	notificationSvc *notification.Service
	accountQueries  *queries.AccountQueries
	userQueries     *queries.UserQueries
	geolocationSvc  *geolocation.Service
	settingQueries  *queries.SettingQueries
}

// NewSettingsHandler creates new handler
func NewSettingsHandler(notificationSvc *notification.Service, accountQueries *queries.AccountQueries, userQueries *queries.UserQueries, geolocationSvc *geolocation.Service, settingQueries *queries.SettingQueries) *SettingsHandler {
	return &SettingsHandler{
		notificationSvc: notificationSvc,
		accountQueries:  accountQueries,
		userQueries:     userQueries,
		geolocationSvc:  geolocationSvc,
		settingQueries: settingQueries,
	}
}

// Get returns current settings
// @Summary Get settings
// @Description Returns current application settings (global settings from DB, organization settings from account.settings)
// @Tags settings
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/settings [get]
func (h *SettingsHandler) Get(c *fiber.Ctx) error {
	safSettings := make(map[string]any)

	// Get global settings from database (SMTP, SMS, Storage, Geolocation)
	if h.settingQueries != nil {
		globalSettings, err := h.settingQueries.GetAllGlobalSettings(c.Context())
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load global settings from database, using config fallback")
		}
		if err == nil && len(globalSettings) > 0 {
			// Email/SMTP settings
			if smtpMap, ok := globalSettings["smtp"]; ok {
				safSettings["email"] = map[string]any{
					"provider":   getString(smtpMap, "provider", ""),
					"smtp_host":  getString(smtpMap, "smtp_host", ""),
					"smtp_port":  getString(smtpMap, "smtp_port", ""),
					"smtp_user":  getString(smtpMap, "smtp_user", ""),
					"from_email": getString(smtpMap, "from_email", ""),
					"from_name":  getString(smtpMap, "from_name", ""),
					// hide smtp_pass
				}
			}

			// SMS settings
			if smsMap, ok := globalSettings["sms"]; ok {
				safSettings["sms"] = map[string]any{
					"twilio_enabled":        getBool(smsMap, "twilio_enabled", false),
					"twilio_account_sid":     getString(smsMap, "twilio_account_sid", ""),
					"twilio_from_number":     getString(smsMap, "twilio_from_number", ""),
					"twilio_auth_token_set":  getString(smsMap, "twilio_auth_token", "") != "",
				}
			}

			// Storage settings (local path is fixed ./lc_uploads, not exposed)
			if storageMap, ok := globalSettings["storage"]; ok {
				safSettings["storage"] = map[string]any{
					"provider": getString(storageMap, "provider", ""),
					"bucket":   getString(storageMap, "bucket", ""),
					"region":   getString(storageMap, "region", ""),
					// hide access_key_id and secret_access_key
				}
			}

			// Geolocation settings (global; paths next to executable)
			if geolocMap, ok := globalSettings["geolocation"]; ok {
				safSettings["geolocation"] = map[string]any{
					"base_dir": appdir.Base(),
					"db_path":  filepath.Join(appdir.Base(), "GeoLite2-City.mmdb"),
					"maxmind_license_key_set": getString(geolocMap, "maxmind_license_key", "") != "",
					"download_url": getString(geolocMap, "download_url", ""),
					"download_method": getString(geolocMap, "download_method", ""),
					"last_updated_at":   getString(geolocMap, "last_updated_at", ""),
					"last_updated_source": getString(geolocMap, "last_updated_source", ""),
				}
				if licenseKey := getString(geolocMap, "maxmind_license_key", ""); licenseKey != "" {
					if geolocSettings, ok := safSettings["geolocation"].(map[string]any); ok {
						geolocSettings["maxmind_license_key_masked"] = maskSecretFirstLast4(licenseKey)
					}
				}
			}

		}
		// Defaults when not in DB
		if _, ok := safSettings["email"]; !ok {
			safSettings["email"] = map[string]any{
				"provider": "", "smtp_host": "", "smtp_port": "", "smtp_user": "",
				"from_email": "", "from_name": "",
			}
		}
		if _, ok := safSettings["sms"]; !ok {
			safSettings["sms"] = map[string]any{
				"twilio_enabled": false, "twilio_account_sid": "", "twilio_from_number": "",
				"twilio_auth_token_set": false,
			}
		}
		if _, ok := safSettings["storage"]; !ok {
			safSettings["storage"] = map[string]any{
				"provider": "", "bucket": "", "region": "",
			}
		}
		if _, ok := safSettings["geolocation"]; !ok {
			safSettings["geolocation"] = map[string]any{
				"base_dir": appdir.Base(),
				"db_path":  filepath.Join(appdir.Base(), "GeoLite2-City.mmdb"),
				"download_url": "", "download_method": "",
			}
		}
	}

	// Get organization-specific settings from account.settings (webhooks, branding, etc.)
	accountID, err := h.getAccountIDFromUser(c)
	if err == nil && h.accountQueries != nil && accountID != "" {
		accountSettings, err := h.accountQueries.GetAccountSettings(c.Context(), accountID)
		if err == nil {
			// Webhook settings (organization-level)
			if webhook, ok := accountSettings["webhook"].(map[string]any); ok {
				safSettings["webhook"] = webhook
			}

			// Branding settings (organization-level)
			if branding, ok := accountSettings["branding"].(map[string]any); ok {
				safSettings["branding"] = branding
			}

			// Merge geolocation last update info from account (set after download from URL/MaxMind)
			if geolocAccount, ok := accountSettings["geolocation"].(map[string]any); ok {
				if g, ok := safSettings["geolocation"].(map[string]any); ok {
					if v := getString(geolocAccount, "last_updated_at", ""); v != "" {
						g["last_updated_at"] = v
					}
					if v := getString(geolocAccount, "last_updated_source", ""); v != "" {
						g["last_updated_source"] = v
					}
				}
			}
		}
	}

	return webutil.Response(c, fiber.StatusOK, "settings", safSettings)
}

// Helper functions
func getString(m map[string]any, key string, defaultValue string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getBool(m map[string]any, key string, defaultValue bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
		if s, ok := v.(string); ok {
			return s == "true"
		}
	}
	return defaultValue
}

// maskSecretFirstLast4 returns the first 4 and last 4 characters of a secret.
// Example: "abcd1234WXYZ" -> "abcd…WXYZ"
func maskSecretFirstLast4(secret string) string {
	secret = strings.TrimSpace(secret)
	if len(secret) <= 8 {
		return secret
	}
	return secret[:4] + "…" + secret[len(secret)-4:]
}

// UpdateEmailRequest request body for updating email settings
type UpdateEmailRequest struct {
	// Accept both old and UI-friendly shapes.
	Provider string `json:"provider"`

	SMTPHost string `json:"smtp_host"`
	SMTPPort string `json:"smtp_port"`
	SMTPUser string `json:"smtp_user"`
	SMTPPass string `json:"smtp_pass,omitempty"` // optional - only if changing

	Host     string `json:"host"`
	Port     any    `json:"port"` // number or string from UI
	Username string `json:"username"`
	Password string `json:"password,omitempty"` // optional - only if changing

	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
}

// UpdateEmail updates email settings (global settings in DB)
// @Summary Update email settings
// @Description Updates email/SMTP configuration (global settings)
// @Tags settings
// @Accept json
// @Produce json
// @Param body body UpdateEmailRequest true "Email settings"
// @Success 200 {object} map[string]any
// @Router /api/settings/email [put]
func (h *SettingsHandler) UpdateEmail(c *fiber.Ctx) error {
	var req UpdateEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if h.settingQueries == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Settings queries not initialized", nil)
	}

	// Get current SMTP settings
	currentSettings, err := h.settingQueries.GetGlobalSetting(c.Context(), "smtp")
	if err != nil {
		currentSettings = make(map[string]any)
	}

	// Normalize payload from UI
	if req.SMTPHost == "" && req.Host != "" {
		req.SMTPHost = req.Host
	}
	if req.SMTPUser == "" && req.Username != "" {
		req.SMTPUser = req.Username
	}
	if req.SMTPPass == "" && req.Password != "" {
		req.SMTPPass = req.Password
	}
	if req.SMTPPort == "" && req.Port != nil {
		req.SMTPPort = fmt.Sprint(req.Port)
	}

	// Update settings map
	if req.Provider != "" {
		currentSettings["provider"] = req.Provider
	} else if currentSettings["provider"] == nil {
		currentSettings["provider"] = "smtp"
	}

	if strings.TrimSpace(req.SMTPHost) != "" {
		currentSettings["smtp_host"] = strings.TrimSpace(req.SMTPHost)
	}
	if strings.TrimSpace(req.SMTPPort) != "" {
		currentSettings["smtp_port"] = strings.TrimSpace(req.SMTPPort)
	}
	if strings.TrimSpace(req.SMTPUser) != "" {
		currentSettings["smtp_user"] = strings.TrimSpace(req.SMTPUser)
	}
	// Only overwrite password if provided
	if strings.TrimSpace(req.SMTPPass) != "" {
		currentSettings["smtp_pass"] = req.SMTPPass
	}
	if strings.TrimSpace(req.FromEmail) != "" {
		currentSettings["from_email"] = strings.TrimSpace(req.FromEmail)
	}
	if strings.TrimSpace(req.FromName) != "" {
		currentSettings["from_name"] = strings.TrimSpace(req.FromName)
	}

	// Save to database
	if err := h.settingQueries.UpdateGlobalSetting(c.Context(), "smtp", currentSettings, "email"); err != nil {
		log.Error().Err(err).Msg("Failed to save email settings to database")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save settings", nil)
	}

	// Update provider instance
	if h.notificationSvc != nil {
		provider := getString(currentSettings, "provider", "smtp")
		if provider == "smtp" {
			var port int
			portStr := getString(currentSettings, "smtp_port", "1025")
			_, _ = fmt.Sscanf(portStr, "%d", &port)
			if port == 0 {
				port = 1025
			}
			h.notificationSvc.RegisterProvider(notification.NewEmailProvider(notification.SMTPConfig{
				Host:      getString(currentSettings, "smtp_host", ""),
				Port:      port,
				User:      getString(currentSettings, "smtp_user", ""),
				Password:  getString(currentSettings, "smtp_pass", ""),
				FromEmail: getString(currentSettings, "from_email", ""),
				FromName:  getString(currentSettings, "from_name", ""),
			}))
		}
	}

	log.Info().Msg("Email settings updated in database")
	
	return webutil.Response(c, fiber.StatusOK, "email_settings", map[string]any{
		"status": "updated",
	})
}

type UpdateSMSRequest struct {
	TwilioEnabled     bool   `json:"twilio_enabled"`
	TwilioAccountSID  string `json:"twilio_account_sid,omitempty"`
	TwilioAuthToken   string `json:"twilio_auth_token,omitempty"` // write-only
	TwilioFromNumber  string `json:"twilio_from_number,omitempty"`
}

func (h *SettingsHandler) UpdateSMS(c *fiber.Ctx) error {
	var req UpdateSMSRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if h.settingQueries == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Settings queries not initialized", nil)
	}

	// Get current SMS settings
	currentSettings, err := h.settingQueries.GetGlobalSetting(c.Context(), "sms")
	if err != nil {
		currentSettings = make(map[string]any)
	}

	// Update settings
	currentSettings["twilio_enabled"] = req.TwilioEnabled

	if strings.TrimSpace(req.TwilioAccountSID) != "" {
		currentSettings["twilio_account_sid"] = strings.TrimSpace(req.TwilioAccountSID)
	}
	if strings.TrimSpace(req.TwilioFromNumber) != "" {
		currentSettings["twilio_from_number"] = strings.TrimSpace(req.TwilioFromNumber)
	}
	// Only overwrite token if provided
	if strings.TrimSpace(req.TwilioAuthToken) != "" {
		currentSettings["twilio_auth_token"] = req.TwilioAuthToken
	}

	// Save to database
	if err := h.settingQueries.UpdateGlobalSetting(c.Context(), "sms", currentSettings, "sms"); err != nil {
		log.Error().Err(err).Msg("Failed to save SMS settings to database")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save settings", nil)
	}

	// Update provider instance
	if h.notificationSvc != nil {
		h.notificationSvc.RegisterProvider(notification.NewSMSProvider(notification.TwilioConfig{
			AccountSID: getString(currentSettings, "twilio_account_sid", ""),
			AuthToken:  getString(currentSettings, "twilio_auth_token", ""),
			FromNumber: getString(currentSettings, "twilio_from_number", ""),
			Enabled:    getBool(currentSettings, "twilio_enabled", false),
		}))
	}

	return webutil.Response(c, fiber.StatusOK, "sms_settings", map[string]any{
		"status": "updated",
	})
}

type TestSMSRequest struct {
	ToPhone string `json:"to_phone" validate:"required"`
	Message string `json:"message,omitempty"`
}

func (h *SettingsHandler) TestSMS(c *fiber.Ctx) error {
	var req TestSMSRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if err := webutil.ValidateStruct(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if h.notificationSvc == nil || !h.notificationSvc.CanSend(models.NotificationTypeSMS) {
		return webutil.Response(c, fiber.StatusBadRequest, "SMS provider is not configured", nil)
	}
	body := strings.TrimSpace(req.Message)
	if body == "" {
		body = "goSign test SMS"
	}
	n := &models.Notification{
		Type:      models.NotificationTypeSMS,
		Recipient: strings.TrimSpace(req.ToPhone),
		Body:      body,
		Context:   map[string]any{},
	}
	if err := h.notificationSvc.Send(n); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to send test SMS", map[string]any{"error": err.Error()})
	}
	return webutil.Response(c, fiber.StatusOK, "test_sms", map[string]any{"status": "success"})
}

// (maskSecret removed: SID isn't secret; token is write-only)

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

// UpdateStorage updates storage settings (global settings in DB)
// @Summary Update storage settings
// @Description Updates storage configuration (local, S3, GCS, Azure) - global settings
// @Tags settings
// @Accept json
// @Produce json
// @Param body body UpdateStorageRequest true "Storage settings"
// @Success 200 {object} map[string]any
// @Router /api/settings/storage [put]
func (h *SettingsHandler) UpdateStorage(c *fiber.Ctx) error {
	var req UpdateStorageRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if h.settingQueries == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Settings queries not initialized", nil)
	}

	// Get current storage settings
	currentSettings, err := h.settingQueries.GetGlobalSetting(c.Context(), "storage")
	if err != nil {
		currentSettings = make(map[string]any)
	}

	// Update settings (local storage uses fixed path ./lc_uploads, base_path not stored)
	currentSettings["provider"] = req.Provider
	if req.Bucket != "" {
		currentSettings["bucket"] = req.Bucket
	}
	if req.Region != "" {
		currentSettings["region"] = req.Region
	}
	if req.Endpoint != "" {
		currentSettings["endpoint"] = req.Endpoint
	}
	if req.AccessKeyID != "" {
		currentSettings["access_key_id"] = req.AccessKeyID
	}
	if req.SecretAccessKey != "" {
		currentSettings["secret_access_key"] = req.SecretAccessKey
	}

	// Save to database
	if err := h.settingQueries.UpdateGlobalSetting(c.Context(), "storage", currentSettings, "storage"); err != nil {
		log.Error().Err(err).Msg("Failed to save storage settings to database")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save settings", nil)
	}

	log.Info().Str("provider", req.Provider).Msg("Storage settings updated in database")
	
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

// UpdateBranding updates branding settings (organization settings in account.settings)
// @Summary Update branding settings
// @Description Updates company branding configuration (organization-level settings)
// @Tags settings
// @Accept json
// @Produce json
// @Param body body UpdateBrandingRequest true "Branding settings"
// @Success 200 {object} map[string]any
// @Router /api/settings/branding [put]
func (h *SettingsHandler) UpdateBranding(c *fiber.Ctx) error {
	accountID, err := h.getAccountIDFromUser(c)
	if err != nil {
		return err
	}
	if accountID == "" {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req UpdateBrandingRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if h.accountQueries == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Account queries not initialized", nil)
	}

	// Get current account settings
	currentSettings, err := h.accountQueries.GetAccountSettings(c.Context(), accountID)
	if err != nil {
		currentSettings = make(map[string]any)
	}

	// Update branding settings
	branding, ok := currentSettings["branding"].(map[string]any)
	if !ok {
		branding = make(map[string]any)
	}

	if req.CompanyName != "" {
		branding["company_name"] = req.CompanyName
	}
	if req.LogoURL != "" {
		branding["logo_url"] = req.LogoURL
	}
	if req.PrimaryColor != "" {
		branding["primary_color"] = req.PrimaryColor
	}
	if req.SecondaryColor != "" {
		branding["secondary_color"] = req.SecondaryColor
	}

	currentSettings["branding"] = branding

	// Save to account.settings
	if err := h.accountQueries.UpdateAccountSettings(c.Context(), accountID, currentSettings); err != nil {
		log.Error().Err(err).Msg("Failed to update branding settings in database")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save settings", nil)
	}

	log.Info().Str("account_id", accountID).Msg("Branding settings updated in database")
	
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
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
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
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid SMTP port", nil)
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
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	// Create storage configuration (local uses fixed ./lc_uploads)
	var storageConfig storage.Config
	switch req.Provider {
	case "local":
		storageConfig = storage.Config{
			Provider: "local",
			BasePath: appdir.LcUploads(),
		}
	case "s3":
		if req.Bucket == "" || req.Region == "" {
			return webutil.Response(c, fiber.StatusBadRequest, "bucket and region are required for S3", nil)
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
		return webutil.Response(c, fiber.StatusBadRequest, "unsupported storage provider", nil)
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
// UpdateGeolocationRequest request body for updating geolocation settings
type UpdateGeolocationRequest struct {
	MaxMindLicenseKey string `json:"maxmind_license_key,omitempty"` // Optional: MaxMind license key
	DownloadURL       string `json:"download_url,omitempty"`         // Optional: Download URL
	DownloadMethod    string `json:"download_method,omitempty"`      // Optional: "maxmind" or "url"
}

// UpdateGeolocation updates geolocation settings (global settings in DB)
// @Summary Update geolocation settings
// @Description Updates geolocation download method and credentials - global settings
// @Tags settings
// @Accept json
// @Produce json
// @Param body body UpdateGeolocationRequest true "Geolocation settings"
// @Success 200 {object} map[string]any
// @Router /api/settings/geolocation [put]
func (h *SettingsHandler) UpdateGeolocation(c *fiber.Ctx) error {
	var req UpdateGeolocationRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if h.settingQueries == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Settings queries not initialized", nil)
	}

	// Get current geolocation settings
	currentSettings, err := h.settingQueries.GetGlobalSetting(c.Context(), "geolocation")
	if err != nil {
		currentSettings = make(map[string]any)
	}

	// Update only provided fields
	licenseKey := strings.TrimSpace(req.MaxMindLicenseKey)
	downloadURL := strings.TrimSpace(req.DownloadURL)
	downloadMethod := strings.TrimSpace(req.DownloadMethod)

	if licenseKey != "" {
		currentSettings["maxmind_license_key"] = licenseKey
	}
	if downloadURL != "" {
		currentSettings["download_url"] = downloadURL
	}
	if downloadMethod != "" {
		currentSettings["download_method"] = downloadMethod
	}

	// Check if at least one field is provided
	if len(currentSettings) == 0 {
		return webutil.Response(c, fiber.StatusBadRequest, "At least one field must be provided", nil)
	}

	// Save to database
	if err := h.settingQueries.UpdateGlobalSetting(c.Context(), "geolocation", currentSettings, "geolocation"); err != nil {
		log.Error().Err(err).Msg("Failed to update geolocation settings in database")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save settings", nil)
	}

	log.Info().
		Str("method", downloadMethod).
		Bool("has_license_key", licenseKey != "").
		Bool("has_download_url", downloadURL != "").
		Msg("Geolocation settings updated in database")

	return webutil.Response(c, fiber.StatusOK, "geolocation_settings", map[string]any{
		"status": "updated",
	})
}

// DownloadGeoLite2FromURLRequest request body for downloading GeoLite2 from URL
type DownloadGeoLite2FromURLRequest struct {
	URL   string `json:"url" validate:"required,url"`
	Force bool   `json:"force,omitempty"`
}

// DownloadGeoLite2FromURL downloads GeoLite2 database from URL
// @Summary Download GeoLite2 database from URL
// @Description Downloads GeoLite2-City.mmdb from provided URL
// @Tags settings
// @Accept json
// @Produce json
// @Param body body DownloadGeoLite2FromURLRequest true "Download request"
// @Success 200 {object} map[string]any
// @Router /api/settings/geolocation/download [post]
func (h *SettingsHandler) DownloadGeoLite2FromURL(c *fiber.Ctx) error {
	var req DownloadGeoLite2FromURLRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if err := webutil.ValidateStruct(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	baseDir := appdir.Base()
	dbPath := filepath.Join(baseDir, "GeoLite2-City.mmdb")

	// Check if database already exists
	if _, err := os.Stat(dbPath); err == nil {
		if !req.Force {
			return webutil.Response(c, fiber.StatusOK, "database_already_exists", map[string]any{
				"status":  "skipped",
				"message": "GeoLite2 database already exists",
				"path":    dbPath,
			})
		}
	}

	// Create base directory if it doesn't exist (next to executable)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Error().Err(err).Str("base_dir", baseDir).Msg("Failed to create base directory")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create base directory", nil)
	}

	// Download from URL (follow redirects automatically)
	client := &http.Client{
		Timeout: 5 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}
	resp, err := client.Get(req.URL)
	if err != nil {
		log.Error().Err(err).Str("url", req.URL).Msg("Failed to download file")
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to download file", map[string]any{
			"error": err.Error(),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("status", resp.StatusCode).Str("url", req.URL).Msg("Failed to download file")
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to download file", map[string]any{
			"error": fmt.Sprintf("HTTP status: %d", resp.StatusCode),
		})
	}

	// Determine file type from URL or Content-Type
	urlLower := strings.ToLower(req.URL)
	contentType := resp.Header.Get("Content-Type")
	isDirectMMDB := strings.HasSuffix(urlLower, ".mmdb") && !strings.HasSuffix(urlLower, ".mmdb.gz")
	isGzipMMDB := strings.HasSuffix(urlLower, ".mmdb.gz") || strings.HasSuffix(urlLower, ".gz")
	isTarGz := strings.HasSuffix(urlLower, ".tar.gz") || strings.Contains(contentType, "application/x-gzip") || strings.Contains(contentType, "application/gzip")

	// Extract into temp output, then atomically replace dbPath
	tmpDBPath := dbPath + ".tmp"
	_ = os.Remove(tmpDBPath)

	// Handle direct .mmdb file (no archive)
	if isDirectMMDB {
		// Create output directory if needed
		if err := os.MkdirAll(filepath.Dir(tmpDBPath), 0755); err != nil {
			log.Error().Err(err).Msg("Failed to create output directory")
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create output directory", nil)
		}

		// Create output file
		outFile, err := os.Create(tmpDBPath)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create output file")
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create output file", nil)
		}
		defer outFile.Close()

		// Copy file content directly
		if _, err := io.Copy(outFile, resp.Body); err != nil {
			log.Error().Err(err).Msg("Failed to save database file")
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save database file", nil)
		}
	} else {
		// Handle archive files (tar.gz, .gz, etc.)
		// Create temporary file for archive
		tmpFile, err := os.CreateTemp("", "geolite2-*")
		if err != nil {
			log.Error().Err(err).Msg("Failed to create temporary file")
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create temporary file", nil)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		// Save archive to temporary file
		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			log.Error().Err(err).Msg("Failed to save archive file")
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save archive file", nil)
		}

		// Extract GeoLite2-City.mmdb from archive
		if isTarGz {
			if err := extractGeoLite2FromTarGz(tmpFile.Name(), tmpDBPath); err != nil {
				log.Error().Err(err).Msg("Failed to extract database from tar.gz")
				return webutil.Response(c, fiber.StatusInternalServerError, "Failed to extract database from tar.gz", map[string]any{
					"error": err.Error(),
				})
			}
		} else if isGzipMMDB {
			// Handle gzip-compressed mmdb file
			if err := extractGeoLite2FromGzipMMDB(tmpFile.Name(), tmpDBPath); err != nil {
				log.Error().Err(err).Msg("Failed to extract database from mmdb.gz")
				return webutil.Response(c, fiber.StatusInternalServerError, "Failed to extract database from mmdb.gz", map[string]any{
					"error": err.Error(),
				})
			}
		} else {
			// Try tar.gz first, then fallback to gzip
			if err := extractGeoLite2FromTarGz(tmpFile.Name(), tmpDBPath); err != nil {
				if gzErr := extractGeoLite2FromGzipMMDB(tmpFile.Name(), tmpDBPath); gzErr != nil {
					log.Error().Err(err).Msg("Failed to extract database from tar.gz")
					log.Error().Err(gzErr).Msg("Failed to extract database from mmdb.gz")
					return webutil.Response(c, fiber.StatusInternalServerError, "Failed to extract database", map[string]any{
						"error": fmt.Sprintf("tar.gz error: %s; gzip error: %s", err.Error(), gzErr.Error()),
					})
				}
			}
		}
	}

	if err := os.Rename(tmpDBPath, dbPath); err != nil {
		_ = os.Remove(tmpDBPath)
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to replace database file", map[string]any{
			"error": err.Error(),
		})
	}

	// Reload in-memory geolocation DB so changes apply immediately
	if h.geolocationSvc != nil {
		if err := h.geolocationSvc.Reload(); err != nil {
			log.Error().Err(err).Msg("Failed to reload GeoLite2 database after update")
			return webutil.Response(c, fiber.StatusInternalServerError, "Database updated but failed to reload", map[string]any{
				"error": err.Error(),
			})
		}
	}

	// Persist last update timestamp (per-account)
	if h.accountQueries != nil {
		if accountID, err := h.getAccountIDFromUser(c); err == nil && accountID != "" {
			_ = h.accountQueries.UpdateAccountGeolocationLastUpdate(c.Context(), accountID, time.Now(), "url")
		}
	}

	log.Info().Str("path", dbPath).Str("url", req.URL).Msg("GeoLite2 database downloaded and extracted successfully")

	return webutil.Response(c, fiber.StatusOK, "database_downloaded", map[string]any{
		"status": "success",
		"path":   dbPath,
	})
}

// DownloadGeoLite2FromMaxMindRequest request body for downloading GeoLite2 from MaxMind
type DownloadGeoLite2FromMaxMindRequest struct {
	LicenseKey string `json:"license_key,omitempty"` // Optional: use from database if not provided
	Force      bool   `json:"force,omitempty"`
}

// DownloadGeoLite2FromMaxMind downloads GeoLite2 database from MaxMind API
// @Summary Download GeoLite2 database from MaxMind
// @Description Downloads GeoLite2-City.mmdb from MaxMind using license key from database or request
// @Tags settings
// @Accept json
// @Produce json
// @Param body body DownloadGeoLite2FromMaxMindRequest false "Download request (license_key optional, uses saved key if not provided)"
// @Success 200 {object} map[string]any
// @Router /api/settings/geolocation/download-maxmind [post]
func (h *SettingsHandler) DownloadGeoLite2FromMaxMind(c *fiber.Ctx) error {
	var req DownloadGeoLite2FromMaxMindRequest
	_ = c.BodyParser(&req) // Optional body

	// Get license key from request or from database
	licenseKey := strings.TrimSpace(req.LicenseKey)
	if licenseKey == "" {
		// Try to get from account settings
		accountID, err := h.getAccountIDFromUser(c)
		if err == nil && h.accountQueries != nil && accountID != "" {
			licenseKey, _ = h.accountQueries.GetAccountGeolocationLicenseKey(c.Context(), accountID)
		}
	}

	if licenseKey == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "MaxMind license key is required. Please configure it in settings first.", nil)
	}

	baseDir := appdir.Base()
	dbPath := filepath.Join(baseDir, "GeoLite2-City.mmdb")

	// Check if database already exists
	if _, err := os.Stat(dbPath); err == nil {
		if !req.Force {
			return webutil.Response(c, fiber.StatusOK, "database_already_exists", map[string]any{
				"status":  "skipped",
				"message": "GeoLite2 database already exists",
				"path":    dbPath,
			})
		}
	}

	// Create base directory if it doesn't exist (next to executable)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Error().Err(err).Str("base_dir", baseDir).Msg("Failed to create base directory")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create base directory", nil)
	}

	// Download from MaxMind API
	downloadURL := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz", licenseKey)
	
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download from MaxMind")
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to download from MaxMind", map[string]any{
			"error": err.Error(),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error().Int("status", resp.StatusCode).Bytes("body", body).Msg("MaxMind API error")
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to download from MaxMind", map[string]any{
			"error": fmt.Sprintf("HTTP status: %d", resp.StatusCode),
		})
	}

	// Create temporary file for tar.gz
	tmpFile, err := os.CreateTemp("", "geolite2-*.tar.gz")
	if err != nil {
		log.Error().Err(err).Msg("Failed to create temporary file")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create temporary file", nil)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Save tar.gz to temporary file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		log.Error().Err(err).Msg("Failed to save tar.gz file")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save tar.gz file", nil)
	}

	// Extract into temp output, then atomically replace dbPath
	tmpDBPath := dbPath + ".tmp"
	_ = os.Remove(tmpDBPath)

	// Extract GeoLite2-City.mmdb from tar.gz
	if err := extractGeoLite2FromTarGz(tmpFile.Name(), tmpDBPath); err != nil {
		log.Error().Err(err).Msg("Failed to extract database from tar.gz")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to extract database", map[string]any{
			"error": err.Error(),
		})
	}

	if err := os.Rename(tmpDBPath, dbPath); err != nil {
		_ = os.Remove(tmpDBPath)
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to replace database file", map[string]any{
			"error": err.Error(),
		})
	}

	// Reload in-memory geolocation DB so changes apply immediately
	if h.geolocationSvc != nil {
		if err := h.geolocationSvc.Reload(); err != nil {
			log.Error().Err(err).Msg("Failed to reload GeoLite2 database after update")
			return webutil.Response(c, fiber.StatusInternalServerError, "Database updated but failed to reload", map[string]any{
				"error": err.Error(),
			})
		}
	}

	// Persist last update timestamp (per-account)
	if h.accountQueries != nil {
		if accountID, err := h.getAccountIDFromUser(c); err == nil && accountID != "" {
			_ = h.accountQueries.UpdateAccountGeolocationLastUpdate(c.Context(), accountID, time.Now(), "maxmind")
		}
	}

	log.Info().Str("path", dbPath).Msg("GeoLite2 database downloaded from MaxMind successfully")

	return webutil.Response(c, fiber.StatusOK, "database_downloaded", map[string]any{
		"status": "success",
		"path":   dbPath,
	})
}

// DeleteGeolocationMaxMindKey removes the saved MaxMind license key from account settings.
func (h *SettingsHandler) DeleteGeolocationMaxMindKey(c *fiber.Ctx) error {
	accountID, err := h.getAccountIDFromUser(c)
	if err != nil {
		return err
	}
	if accountID == "" || h.accountQueries == nil {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	if err := h.accountQueries.DeleteAccountGeolocationMaxMindLicenseKey(c.Context(), accountID); err != nil {
		log.Error().Err(err).Msg("Failed to delete MaxMind license key")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to delete key", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "deleted", map[string]any{
		"status": "deleted",
	})
}

// extractGeoLite2FromTarGz extracts GeoLite2-City.mmdb from tar.gz archive
func extractGeoLite2FromTarGz(tarGzPath, outputPath string) error {
	// Open tar.gz file
	file, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Find and extract GeoLite2-City.mmdb
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Look for GeoLite2-City.mmdb file
		if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, "GeoLite2-City.mmdb") {
			// Create output directory if needed
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Create output file
			outFile, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer outFile.Close()

			// Copy file content
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("failed to extract file: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("GeoLite2-City.mmdb not found in archive")
}

// extractGeoLite2FromGzipMMDB extracts GeoLite2-City.mmdb from a gzip-compressed mmdb file (GeoLite2-City.mmdb.gz).
func extractGeoLite2FromGzipMMDB(gzPath, outputPath string) error {
	file, err := os.Open(gzPath)
	if err != nil {
		return fmt.Errorf("failed to open gzip file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, gzReader); err != nil {
		return fmt.Errorf("failed to extract gzip file: %w", err)
	}

	return nil
}

// getAccountIDFromUser resolves the current user's account_id via database.
// We don't rely on middleware setting account_id (JWT middleware only sets user_id).
func (h *SettingsHandler) getAccountIDFromUser(c *fiber.Ctx) (string, error) {
	if h.userQueries == nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, "User queries not initialized")
	}

	userID, err := GetUserID(c)
	if err != nil {
		return "", err
	}

	accountID, err := h.userQueries.GetUserAccountID(c.Context(), userID)
	if err != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, "Failed to resolve account")
	}
	if strings.TrimSpace(accountID) == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Account not found")
	}

	return accountID, nil
}

func (h *SettingsHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.Get)
	router.Put("/email", h.UpdateEmail)
	router.Put("/sms", h.UpdateSMS)
	router.Put("/storage", h.UpdateStorage)
	router.Put("/branding", h.UpdateBranding)
	router.Put("/geolocation", h.UpdateGeolocation)
	router.Delete("/geolocation/maxmind-key", h.DeleteGeolocationMaxMindKey)
	router.Post("/email/test", h.TestEmail)
	router.Post("/sms/test", h.TestSMS)
	router.Post("/storage/test", h.TestStorage)
	router.Post("/geolocation/download", h.DownloadGeoLite2FromURL)
	router.Post("/geolocation/download-maxmind", h.DownloadGeoLite2FromMaxMind)
}

