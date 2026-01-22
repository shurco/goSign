package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// BrandingHandler handles branding-related API requests
type BrandingHandler struct {
	accountQueries *queries.AccountQueries
	userQueries    *queries.UserQueries
	storage        interface{} // TODO: implement storage interface
}

// NewBrandingHandler creates a new branding handler
func NewBrandingHandler(accountQueries *queries.AccountQueries, userQueries *queries.UserQueries, storage interface{}) *BrandingHandler {
	return &BrandingHandler{
		accountQueries: accountQueries,
		userQueries:    userQueries,
		storage:        storage,
	}
}

// GetBranding returns current branding settings
// @Summary Get branding settings
// @Description Returns current account branding configuration
// @Tags branding
// @Produce json
// @Success 200 {object} models.BrandingSettings
// @Router /api/v1/branding [get]
func (h *BrandingHandler) GetBranding(c *fiber.Ctx) error {
	_, err := GetAccountID(c)
	if err != nil {
		return err
	}

	// TODO: Load branding from account settings
	branding := models.BrandingSettings{
		ShowPoweredBy: true,
	}

	return webutil.Response(c, fiber.StatusOK, "Branding retrieved successfully", map[string]interface{}{
		"branding": branding,
	})
}

// UpdateBrandingSettingsRequest request body for updating branding
type UpdateBrandingSettingsRequest struct {
	Branding models.BrandingSettings `json:"branding"`
}

// UpdateBranding updates branding settings
// @Summary Update branding settings
// @Description Updates account branding configuration
// @Tags branding
// @Accept json
// @Produce json
// @Param body body UpdateBrandingSettingsRequest true "Branding settings"
// @Success 200 {object} map[string]any
// @Router /api/v1/branding [put]
func (h *BrandingHandler) UpdateBranding(c *fiber.Ctx) error {
	accountID, err := GetAccountID(c)
	if err != nil {
		return err
	}

	var req UpdateBrandingSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	// TODO: Save to account.settings_jsonb
	log.Info().Str("account_id", accountID).Msg("Branding settings updated")

	return webutil.Response(c, fiber.StatusOK, "Branding updated successfully", map[string]interface{}{
		"branding": req.Branding,
	})
}

// UploadAssetRequest request body for uploading asset
type UploadAssetRequest struct {
	Type string `json:"type" validate:"required,oneof=logo favicon email_header watermark"`
}

// UploadAsset uploads a branding asset
// @Summary Upload branding asset
// @Description Uploads logo, favicon, or other branding assets
// @Tags branding
// @Accept multipart/form-data
// @Produce json
// @Param type formData string true "Asset type (logo, favicon, email_header, watermark)"
// @Param file formData file true "Asset file"
// @Success 200 {object} models.BrandingAsset
// @Router /api/v1/branding/assets [post]
func (h *BrandingHandler) UploadAsset(c *fiber.Ctx) error {
	accountID, err := GetAccountID(c)
	if err != nil {
		return err
	}

	assetType := c.FormValue("type")
	if assetType == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Asset type is required", nil)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "File is required", nil)
	}

	// TODO: Validate file type and size
	// TODO: Save to storage
	// TODO: Create asset record in database

	log.Info().Str("account_id", accountID).Str("type", assetType).Str("filename", file.Filename).Msg("Branding asset uploaded")

	asset := models.BrandingAsset{
		ID:        "", // TODO: generate UUID
		AccountID: accountID,
		Type:      assetType,
		FilePath:  "", // TODO: set from storage
		MimeType:  file.Header.Get("Content-Type"),
	}

	return webutil.Response(c, fiber.StatusOK, "Asset uploaded successfully", asset)
}

// DeleteAsset deletes a branding asset
// @Summary Delete branding asset
// @Description Deletes a branding asset by ID
// @Tags branding
// @Produce json
// @Param id path string true "Asset ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/branding/assets/:id [delete]
func (h *BrandingHandler) DeleteAsset(c *fiber.Ctx) error {
	accountID, err := GetAccountID(c)
	if err != nil {
		return err
	}

	assetID := c.Params("id")
	if assetID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Asset ID is required", nil)
	}

	// TODO: Verify asset belongs to account
	// TODO: Delete from storage
	// TODO: Delete from database

	log.Info().Str("account_id", accountID).Str("asset_id", assetID).Msg("Branding asset deleted")

	return webutil.Response(c, fiber.StatusOK, "Asset deleted successfully", nil)
}

// AddCustomDomainRequest request body for adding custom domain
type AddCustomDomainRequest struct {
	Domain string `json:"domain" validate:"required"`
}

// AddCustomDomain adds a custom domain
// @Summary Add custom domain
// @Description Adds a custom domain for white-label branding
// @Tags branding
// @Accept json
// @Produce json
// @Param body body AddCustomDomainRequest true "Domain configuration"
// @Success 200 {object} models.CustomDomain
// @Router /api/v1/branding/domain [post]
func (h *BrandingHandler) AddCustomDomain(c *fiber.Ctx) error {
	accountID, err := GetAccountID(c)
	if err != nil {
		return err
	}

	var req AddCustomDomainRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	// TODO: Validate domain format
	// TODO: Create domain record
	// TODO: Generate verification token

	log.Info().Str("account_id", accountID).Str("domain", req.Domain).Msg("Custom domain added")

	domain := models.CustomDomain{
		ID:        "", // TODO: generate UUID
		AccountID: accountID,
		Domain:    req.Domain,
		Verified:  false,
	}

	return webutil.Response(c, fiber.StatusOK, "Custom domain added successfully", domain)
}

// VerifyCustomDomain verifies a custom domain
// @Summary Verify custom domain
// @Description Verifies ownership of a custom domain
// @Tags branding
// @Produce json
// @Param domain_id path string true "Domain ID"
// @Success 200 {object} models.CustomDomain
// @Router /api/v1/branding/domain/:domain_id/verify [post]
func (h *BrandingHandler) VerifyCustomDomain(c *fiber.Ctx) error {
	accountID, err := GetAccountID(c)
	if err != nil {
		return err
	}

	domainID := c.Params("domain_id")
	if domainID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Domain ID is required", nil)
	}

	// TODO: Verify domain ownership via DNS TXT record or file upload
	// TODO: Update domain record

	log.Info().Str("account_id", accountID).Str("domain_id", domainID).Msg("Custom domain verified")

	return webutil.Response(c, fiber.StatusOK, "Domain verified successfully", nil)
}

// PreviewBranding returns branding preview
// @Summary Preview branding
// @Description Returns preview of branding settings
// @Tags branding
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/v1/branding/preview [get]
func (h *BrandingHandler) PreviewBranding(c *fiber.Ctx) error {
	_, err := GetAccountID(c)
	if err != nil {
		return err
	}

	// TODO: Load branding and return preview data
	preview := map[string]interface{}{
		"theme": "default",
		"colors": map[string]string{
			"primary": "#4F46E5",
		},
	}

	return webutil.Response(c, fiber.StatusOK, "Preview generated successfully", preview)
}

// RegisterRoutes registers all branding routes
func (h *BrandingHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.GetBranding)
	router.Put("/", h.UpdateBranding)
	router.Post("/assets", h.UploadAsset)
	router.Delete("/assets/:id", h.DeleteAsset)
	router.Post("/domain", h.AddCustomDomain)
	router.Post("/domain/:domain_id/verify", h.VerifyCustomDomain)
	router.Get("/preview", h.PreviewBranding)
}
