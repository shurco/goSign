package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// I18nHandler handles i18n-related API requests
type I18nHandler struct {
	userQueries    *queries.UserQueries
	accountQueries *queries.AccountQueries
}

// NewI18nHandler creates a new i18n handler
func NewI18nHandler(userQueries *queries.UserQueries, accountQueries *queries.AccountQueries) *I18nHandler {
	return &I18nHandler{
		userQueries:    userQueries,
		accountQueries: accountQueries,
	}
}

// AvailableLocales represents available locales
type AvailableLocales struct {
	UILocales      map[string]string `json:"ui_locales"`      // 7 UI languages
	SigningLocales map[string]string `json:"signing_locales"` // 14 signing languages
}

// GetLocales returns available locales
// @Summary Get available locales
// @Description Returns list of available UI and signing portal locales
// @Tags i18n
// @Produce json
// @Success 200 {object} AvailableLocales
// @Router /api/v1/i18n/locales [get]
func (h *I18nHandler) GetLocales(c *fiber.Ctx) error {
	uiLocales := map[string]string{
		"en": "English",
		"ru": "Русский",
		"es": "Español",
		"fr": "Français",
		"de": "Deutsch",
		"it": "Italiano",
		"pt": "Português",
	}

	signingLocales := map[string]string{
		"en": "English",
		"ru": "Русский",
		"es": "Español",
		"fr": "Français",
		"de": "Deutsch",
		"it": "Italiano",
		"pt": "Português",
		"zh": "中文",
		"ja": "日本語",
		"ko": "한국어",
		"ar": "العربية",
		"hi": "हिन्दी",
		"pl": "Polski",
		"nl": "Nederlands",
	}

	return webutil.Response(c, fiber.StatusOK, "Locales retrieved successfully", AvailableLocales{
		UILocales:      uiLocales,
		SigningLocales: signingLocales,
	})
}

// UpdateUserLocaleRequest request body for updating user locale
type UpdateUserLocaleRequest struct {
	Locale string `json:"locale" validate:"required,oneof=en ru es fr de it pt"`
}

// UpdateUserLocale updates user's preferred locale
// @Summary Update user locale
// @Description Updates the preferred locale for the current user
// @Tags i18n
// @Accept json
// @Produce json
// @Param body body UpdateUserLocaleRequest true "Locale update request"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/user/locale [put]
func (h *I18nHandler) UpdateUserLocale(c *fiber.Ctx) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	var req UpdateUserLocaleRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	// Validate locale
	validLocales := map[string]bool{
		"en": true, "ru": true, "es": true, "fr": true,
		"de": true, "it": true, "pt": true,
	}
	if !validLocales[req.Locale] {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid locale", nil)
	}

	// Update user locale
	if err := h.userQueries.UpdateUserLocale(c.Context(), userID, req.Locale); err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("locale", req.Locale).Msg("Failed to update user locale")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to update locale", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Locale updated successfully", map[string]interface{}{
		"locale": req.Locale,
	})
}

// UpdateAccountLocaleRequest request body for updating account locale
type UpdateAccountLocaleRequest struct {
	Locale string `json:"locale" validate:"required,oneof=en ru es fr de it pt"`
}

// UpdateAccountLocale updates account locale
// @Summary Update account locale
// @Description Updates the locale for the current account
// @Tags i18n
// @Accept json
// @Produce json
// @Param body body UpdateAccountLocaleRequest true "Locale update request"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/account/locale [put]
func (h *I18nHandler) UpdateAccountLocale(c *fiber.Ctx) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Get account ID from user
	accountID, err := h.userQueries.GetUserAccountID(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to get user account")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get account", nil)
	}

	var req UpdateAccountLocaleRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	// Validate locale
	validLocales := map[string]bool{
		"en": true, "ru": true, "es": true, "fr": true,
		"de": true, "it": true, "pt": true,
	}
	if !validLocales[req.Locale] {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid locale", nil)
	}

	// Update account locale
	if err := h.accountQueries.UpdateAccountLocale(c.Context(), accountID, req.Locale); err != nil {
		log.Error().Err(err).Str("account_id", accountID).Str("locale", req.Locale).Msg("Failed to update account locale")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to update locale", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Locale updated successfully", map[string]interface{}{
		"locale": req.Locale,
	})
}

// RegisterRoutes registers i18n routes
func (h *I18nHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/locales", h.GetLocales)
	router.Put("/user/locale", h.UpdateUserLocale)
	router.Put("/account/locale", h.UpdateAccountLocale)
}
