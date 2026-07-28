package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// BrandingHandler handles branding-related API requests
type BrandingHandler struct {
	accountQueries *queries.AccountQueries
}

// NewBrandingHandler creates a new branding handler
func NewBrandingHandler(accountQueries *queries.AccountQueries) *BrandingHandler {
	return &BrandingHandler{
		accountQueries: accountQueries,
	}
}

// GetBranding returns current branding settings
// @Summary Get branding settings
// @Description Returns current account branding configuration
// @Tags branding
// @Produce json
// @Success 200 {object} models.BrandingSettings
// @Router /v1/settings/branding [get]
func (h *BrandingHandler) GetBranding(c fiber.Ctx) error {
	accountID, err := GetAccountID(c)
	if err != nil {
		return err
	}
	if h.accountQueries == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Account queries not initialized", nil)
	}

	branding := models.BrandingSettings{
		ShowPoweredBy: true,
	}

	settings, err := h.accountQueries.GetAccountSettings(c.Context(), accountID)
	if err != nil {
		log.Error().Err(err).Str("account_id", accountID).Msg("Failed to load branding settings")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to load settings", nil)
	}

	if raw, ok := settings["branding"]; ok {
		if rawJSON, err := json.Marshal(raw); err == nil {
			// Defaults above are kept for keys absent in the stored settings.
			_ = json.Unmarshal(rawJSON, &branding)
		}
	}

	return webutil.Response(c, fiber.StatusOK, "Branding retrieved successfully", map[string]any{
		"branding": branding,
	})
}

// UpdateBrandingSettingsRequest request body for updating branding
type UpdateBrandingSettingsRequest struct {
	// RawMessage keeps only the keys the client actually sent,
	// so partial updates don't reset the remaining settings.
	Branding json.RawMessage `json:"branding"`
}

// UpdateBranding updates branding settings
// @Summary Update branding settings
// @Description Updates account branding configuration, merging provided keys into stored settings
// @Tags branding
// @Accept json
// @Produce json
// @Param body body models.BrandingSettings true "Branding settings"
// @Success 200 {object} map[string]any
// @Router /v1/settings/branding [put]
func (h *BrandingHandler) UpdateBranding(c fiber.Ctx) error {
	accountID, err := GetAccountID(c)
	if err != nil {
		return err
	}

	var req UpdateBrandingSettingsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	var typed models.BrandingSettings
	if err := json.Unmarshal(req.Branding, &typed); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid branding settings", nil)
	}
	var patch map[string]any
	if err := json.Unmarshal(req.Branding, &patch); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid branding settings", nil)
	}

	if h.accountQueries == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Account queries not initialized", nil)
	}

	settings, err := h.accountQueries.GetAccountSettings(c.Context(), accountID)
	if err != nil {
		settings = make(map[string]any)
	}

	branding, ok := settings["branding"].(map[string]any)
	if !ok {
		branding = make(map[string]any)
	}
	for k, v := range patch {
		branding[k] = v
	}
	settings["branding"] = branding

	if err := h.accountQueries.UpdateAccountSettings(c.Context(), accountID, settings); err != nil {
		log.Error().Err(err).Str("account_id", accountID).Msg("Failed to save branding settings")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save settings", nil)
	}

	log.Info().Str("account_id", accountID).Msg("Branding settings updated")

	return webutil.Response(c, fiber.StatusOK, "Branding updated successfully", map[string]any{
		"branding": branding,
	})
}

// RegisterRoutes registers all branding routes
func (h *BrandingHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.GetBranding)
	router.Put("/", h.UpdateBranding)
}
