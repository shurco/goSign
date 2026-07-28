package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	workerv1 "github.com/shurco/gosign/internal/rpc/workerv1"
	"github.com/shurco/gosign/pkg/appdir"
	"github.com/shurco/gosign/pkg/notification"
	"github.com/shurco/gosign/pkg/storage"
	"github.com/shurco/gosign/pkg/utils"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// SettingsHandler handles requests to settings.
// GeoLite2 operations (download, refresh) are delegated to the worker via
// gRPC; the server process never touches the database file itself.
type SettingsHandler struct {
	notificationSvc *notification.Service
	accountQueries  *queries.AccountQueries
	userQueries     *queries.UserQueries
	worker          workerv1.WorkerServiceClient
	settingQueries  *queries.SettingQueries
}

// NewSettingsHandler creates new handler; worker may be nil (GeoLite2
// download endpoints respond 503 until workers are reachable).
func NewSettingsHandler(notificationSvc *notification.Service, accountQueries *queries.AccountQueries, userQueries *queries.UserQueries, worker workerv1.WorkerServiceClient, settingQueries *queries.SettingQueries) *SettingsHandler {
	return &SettingsHandler{
		notificationSvc: notificationSvc,
		accountQueries:  accountQueries,
		userQueries:     userQueries,
		worker:          worker,
		settingQueries:  settingQueries,
	}
}

// Get returns current settings
// @Summary Get settings
// @Description Returns current application settings (global settings from DB, organization settings from account.settings)
// @Tags settings
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/settings [get]
func (h *SettingsHandler) Get(c fiber.Ctx) error {
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
					"provider":   utils.GetStringFromMap(smtpMap, "provider", ""),
					"smtp_host":  utils.GetStringFromMap(smtpMap, "smtp_host", ""),
					"smtp_port":  utils.GetStringFromMap(smtpMap, "smtp_port", ""),
					"smtp_user":  utils.GetStringFromMap(smtpMap, "smtp_user", ""),
					"from_email": utils.GetStringFromMap(smtpMap, "from_email", ""),
					"from_name":  utils.GetStringFromMap(smtpMap, "from_name", ""),
					// hide smtp_pass
				}
			}

			// SMS settings
			if smsMap, ok := globalSettings["sms"]; ok {
				safSettings["sms"] = map[string]any{
					"twilio_enabled":        utils.GetBoolFromMap(smsMap, "twilio_enabled", false),
					"twilio_account_sid":    utils.GetStringFromMap(smsMap, "twilio_account_sid", ""),
					"twilio_from_number":    utils.GetStringFromMap(smsMap, "twilio_from_number", ""),
					"twilio_auth_token_set": utils.GetStringFromMap(smsMap, "twilio_auth_token", "") != "",
				}
			}

			// Storage settings (local path is fixed ./lc_uploads, not exposed)
			if storageMap, ok := globalSettings["storage"]; ok {
				safSettings["storage"] = map[string]any{
					"provider": utils.GetStringFromMap(storageMap, "provider", ""),
					"bucket":   utils.GetStringFromMap(storageMap, "bucket", ""),
					"region":   utils.GetStringFromMap(storageMap, "region", ""),
					// hide access_key_id and secret_access_key
				}
			}

			// Geolocation settings (global; paths next to executable)
			if geolocMap, ok := globalSettings["geolocation"]; ok {
				safSettings["geolocation"] = map[string]any{
					"base_dir":                appdir.Base(),
					"db_path":                 appdir.GeoLite2(),
					"maxmind_license_key_set": utils.GetStringFromMap(geolocMap, "maxmind_license_key", "") != "",
					"download_url":            utils.GetStringFromMap(geolocMap, "download_url", ""),
					"download_method":         utils.GetStringFromMap(geolocMap, "download_method", ""),
					"last_updated_at":         utils.GetStringFromMap(geolocMap, "last_updated_at", ""),
					"last_updated_source":     utils.GetStringFromMap(geolocMap, "last_updated_source", ""),
				}
				if licenseKey := utils.GetStringFromMap(geolocMap, "maxmind_license_key", ""); licenseKey != "" {
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
				"base_dir":     appdir.Base(),
				"db_path":      appdir.GeoLite2(),
				"download_url": "", "download_method": "",
			}
		}
	}

	// Get organization-specific settings from account.settings (webhooks, branding, etc.)
	accountID, err := ResolveAccountID(c, h.userQueries)
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
					if v := utils.GetStringFromMap(geolocAccount, "last_updated_at", ""); v != "" {
						g["last_updated_at"] = v
					}
					if v := utils.GetStringFromMap(geolocAccount, "last_updated_source", ""); v != "" {
						g["last_updated_source"] = v
					}
				}
			}
		}
	}

	return webutil.Response(c, fiber.StatusOK, "settings", safSettings)
}

// maskSecretFirstLast4 returns the first 4 and last 4 characters of a secret.
// Example: "abcd1234WXYZ" -> "abcd…WXYZ"
func maskSecretFirstLast4(secret string) string {
	secret = strings.TrimSpace(secret)
	if len(secret) == 0 {
		return ""
	}
	if len(secret) <= 8 {
		return strings.Repeat("*", len(secret))
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
func (h *SettingsHandler) UpdateEmail(c fiber.Ctx) error {
	var req UpdateEmailRequest
	if err := c.Bind().JSON(&req); err != nil {
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
		provider := utils.GetStringFromMap(currentSettings, "provider", "smtp")
		if provider == "smtp" {
			var port int
			portStr := utils.GetStringFromMap(currentSettings, "smtp_port", "1025")
			_, _ = fmt.Sscanf(portStr, "%d", &port)
			if port == 0 {
				port = 1025
			}
			h.notificationSvc.RegisterProvider(notification.NewEmailProvider(notification.SMTPConfig{
				Host:      utils.GetStringFromMap(currentSettings, "smtp_host", ""),
				Port:      port,
				User:      utils.GetStringFromMap(currentSettings, "smtp_user", ""),
				Password:  utils.GetStringFromMap(currentSettings, "smtp_pass", ""),
				FromEmail: utils.GetStringFromMap(currentSettings, "from_email", ""),
				FromName:  utils.GetStringFromMap(currentSettings, "from_name", ""),
			}))
		}
	}

	log.Info().Msg("Email settings updated in database")

	return webutil.Response(c, fiber.StatusOK, "email_settings", map[string]any{
		"status": "updated",
	})
}

type UpdateSMSRequest struct {
	TwilioEnabled    bool   `json:"twilio_enabled"`
	TwilioAccountSID string `json:"twilio_account_sid,omitempty"`
	TwilioAuthToken  string `json:"twilio_auth_token,omitempty"` // write-only
	TwilioFromNumber string `json:"twilio_from_number,omitempty"`
}

func (h *SettingsHandler) UpdateSMS(c fiber.Ctx) error {
	var req UpdateSMSRequest
	if err := c.Bind().JSON(&req); err != nil {
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
			AccountSID: utils.GetStringFromMap(currentSettings, "twilio_account_sid", ""),
			AuthToken:  utils.GetStringFromMap(currentSettings, "twilio_auth_token", ""),
			FromNumber: utils.GetStringFromMap(currentSettings, "twilio_from_number", ""),
			Enabled:    utils.GetBoolFromMap(currentSettings, "twilio_enabled", false),
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

func (h *SettingsHandler) TestSMS(c fiber.Ctx) error {
	var req TestSMSRequest
	if err := c.Bind().JSON(&req); err != nil {
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
func (h *SettingsHandler) UpdateStorage(c fiber.Ctx) error {
	var req UpdateStorageRequest
	if err := c.Bind().JSON(&req); err != nil {
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

// TestEmailRequest request body for testing email
type TestEmailRequest struct {
	Provider  string `json:"provider"`
	SMTPHost  string `json:"smtp_host" validate:"required"`
	SMTPPort  string `json:"smtp_port" validate:"required"`
	SMTPUser  string `json:"smtp_user" validate:"required"`
	SMTPPass  string `json:"smtp_pass" validate:"required"`
	FromEmail string `json:"from_email" validate:"required,email"`
	FromName  string `json:"from_name"`
	ToEmail   string `json:"to_email" validate:"required,email"`
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
func (h *SettingsHandler) TestEmail(c fiber.Ctx) error {
	var req TestEmailRequest
	if err := c.Bind().JSON(&req); err != nil {
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
func (h *SettingsHandler) TestStorage(c fiber.Ctx) error {
	var req TestStorageRequest
	if err := c.Bind().JSON(&req); err != nil {
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

// UpdateGeolocationRequest is the request body for updating geolocation settings.
type UpdateGeolocationRequest struct {
	MaxMindLicenseKey string `json:"maxmind_license_key,omitempty"` // Optional: MaxMind license key
	DownloadURL       string `json:"download_url,omitempty"`        // Optional: Download URL
	DownloadMethod    string `json:"download_method,omitempty"`     // Optional: "maxmind" or "url"
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
func (h *SettingsHandler) UpdateGeolocation(c fiber.Ctx) error {
	var req UpdateGeolocationRequest
	if err := c.Bind().JSON(&req); err != nil {
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

// DownloadGeoLite2Request request body for downloading the GeoLite2 database.
// Method selects the source: "url" (direct or archive URL) or "maxmind"
// (official MaxMind API, license key from request or saved settings).
// When omitted, "url" is assumed if url is set, otherwise "maxmind".
type DownloadGeoLite2Request struct {
	Method     string `json:"method,omitempty" validate:"omitempty,oneof=url maxmind"`
	URL        string `json:"url,omitempty" validate:"omitempty,url"`
	LicenseKey string `json:"license_key,omitempty"`
	Force      bool   `json:"force,omitempty"`
}

// geoLite2SyncTimeout bounds the worker download RPC (the network download
// itself can take minutes).
const geoLite2SyncTimeout = 6 * time.Minute

// DownloadGeoLite2 downloads the GeoLite2 database from a URL or MaxMind.
// The download is performed by the worker (gRPC); across worker replicas
// only one download runs at a time.
// @Summary Download GeoLite2 database
// @Description Downloads GeoLite2-City.mmdb via the worker from a URL (method=url) or from MaxMind (method=maxmind)
// @Tags settings
// @Accept json
// @Produce json
// @Param body body DownloadGeoLite2Request true "Download request"
// @Success 200 {object} map[string]any
// @Router /v1/settings/geolocation/download [post]
func (h *SettingsHandler) DownloadGeoLite2(c fiber.Ctx) error {
	var req DownloadGeoLite2Request
	if err := c.Bind().JSON(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if err := webutil.ValidateStruct(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if h.worker == nil {
		return webutil.Response(c, fiber.StatusServiceUnavailable, "Worker is not available", nil)
	}

	method := req.Method
	if method == "" {
		if req.URL != "" {
			method = "url"
		} else {
			method = "maxmind"
		}
	}
	if method == "url" && req.URL == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Field 'url' is required for method 'url'", nil)
	}

	accountID := ""
	if id, err := ResolveAccountID(c, h.userQueries); err == nil {
		accountID = id
	}

	// Keep the per-account license key resolution on the API side; the
	// worker falls back to globally saved settings otherwise.
	licenseKey := strings.TrimSpace(req.LicenseKey)
	if method == "maxmind" && licenseKey == "" && h.accountQueries != nil && accountID != "" {
		licenseKey, _ = h.accountQueries.GetAccountGeolocationLicenseKey(c.Context(), accountID)
	}

	ctx, cancel := context.WithTimeout(c.Context(), geoLite2SyncTimeout)
	defer cancel()

	resp, err := h.worker.SyncGeoLite2(ctx, &workerv1.SyncGeoLite2Request{
		Force:      req.Force,
		Method:     method,
		Url:        req.URL,
		LicenseKey: licenseKey,
		AccountId:  accountID,
	})
	if err != nil {
		log.Error().Err(err).Str("method", method).Msg("GeoLite2 download via worker failed")
		return webutil.Response(c, fiber.StatusBadRequest, "Failed to download GeoLite2 database", map[string]any{"error": err.Error()})
	}

	switch resp.GetStatus() {
	case workerv1.SyncStatus_SYNC_STATUS_SKIPPED:
		return webutil.Response(c, fiber.StatusOK, "database_already_exists", map[string]any{
			"status":  "skipped",
			"message": "GeoLite2 database already exists",
			"path":    appdir.GeoLite2(),
		})
	case workerv1.SyncStatus_SYNC_STATUS_BUSY:
		return webutil.Response(c, fiber.StatusConflict, "update_in_progress", map[string]any{
			"status":  "busy",
			"message": "Another worker replica is updating the database right now",
		})
	default:
		log.Info().Str("method", method).Msg("GeoLite2 database downloaded via worker")
		return webutil.Response(c, fiber.StatusOK, "database_downloaded", map[string]any{
			"status": "success",
			"path":   appdir.GeoLite2(),
		})
	}
}

// DeleteGeolocationMaxMindKey removes the saved MaxMind license key from account settings.
func (h *SettingsHandler) DeleteGeolocationMaxMindKey(c fiber.Ctx) error {
	accountID, err := ResolveAccountID(c, h.userQueries)
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

func (h *SettingsHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.Get)
	router.Put("/email", h.UpdateEmail)
	router.Put("/sms", h.UpdateSMS)
	router.Put("/storage", h.UpdateStorage)
	router.Put("/geolocation", h.UpdateGeolocation)
	router.Delete("/geolocation/maxmind-key", h.DeleteGeolocationMaxMindKey)
	router.Post("/email/test", h.TestEmail)
	router.Post("/sms/test", h.TestSMS)
	router.Post("/storage/test", h.TestStorage)
	router.Post("/geolocation/download", h.DownloadGeoLite2)
}
