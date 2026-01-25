package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// EmailTemplateHandler handles email template API requests
type EmailTemplateHandler struct {
	emailTemplateQueries *queries.EmailTemplateQueries
	userQueries          *queries.UserQueries
}

// NewEmailTemplateHandler creates a new email template handler
func NewEmailTemplateHandler(emailTemplateQueries *queries.EmailTemplateQueries, userQueries *queries.UserQueries) *EmailTemplateHandler {
	return &EmailTemplateHandler{
		emailTemplateQueries: emailTemplateQueries,
		userQueries:          userQueries,
	}
}

// verifyTemplateOwnership ensures the template belongs to the current scope (organization or account)
func (h *EmailTemplateHandler) verifyTemplateOwnership(c *fiber.Ctx, existing *models.EmailTemplate) error {
	orgID, _ := GetOrganizationID(c)
	if existing.OrganizationID != nil && *existing.OrganizationID != "" {
		if orgID != *existing.OrganizationID {
			return webutil.Response(c, fiber.StatusForbidden, "Template does not belong to current organization", nil)
		}
		return nil
	}
	if existing.AccountID != nil && *existing.AccountID != "" {
		userID, err := GetUserID(c)
		if err != nil {
			return err
		}
		accountID, err := h.userQueries.GetUserAccountID(c.Context(), userID)
		if err != nil || accountID != *existing.AccountID {
			return webutil.Response(c, fiber.StatusForbidden, "Template does not belong to current account", nil)
		}
		return nil
	}
	return webutil.Response(c, fiber.StatusForbidden, "Cannot modify this template", nil)
}

// getEmailTemplateScope returns accountID and organizationID for the current request.
// When user is in organization context (JWT has organization_id), organizationID is set and accountID is nil for scoping.
func (h *EmailTemplateHandler) getEmailTemplateScope(c *fiber.Ctx) (accountIDPtr *string, organizationIDPtr *string, err error) {
	orgID, _ := GetOrganizationID(c)
	if orgID != "" {
		return nil, &orgID, nil
	}
	userID, err := GetUserID(c)
	if err != nil {
		return nil, nil, err
	}
	accountID, err := h.userQueries.GetUserAccountID(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user account ID")
		return nil, nil, err
	}
	if accountID != "" {
		return &accountID, nil, nil
	}
	return nil, nil, nil
}

// GetAllEmailTemplates retrieves all email templates for the current scope (organization or account)
// @Summary Get all email templates
// @Description Returns all email templates for the current organization or account (and system templates)
// @Tags email-templates
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/email-templates [get]
func (h *EmailTemplateHandler) GetAllEmailTemplates(c *fiber.Ctx) error {
	accountIDPtr, organizationIDPtr, err := h.getEmailTemplateScope(c)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to resolve scope", nil)
	}

	locale := c.Query("locale", "en")
	var localePtr *string
	if locale != "" {
		localePtr = &locale
	}

	templates, err := h.emailTemplateQueries.GetAllEmailTemplates(c.Context(), accountIDPtr, organizationIDPtr, localePtr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get email templates")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get email templates", map[string]interface{}{
			"error": err.Error(),
		})
	}

	log.Info().Int("count", len(templates)).Msg("Retrieved email templates")

	return webutil.Response(c, fiber.StatusOK, "Email templates retrieved successfully", map[string]interface{}{
		"templates": templates,
	})
}

// GetEmailTemplate retrieves a specific email template by name
// @Summary Get email template
// @Description Returns a specific email template by name for current organization or account
// @Tags email-templates
// @Produce json
// @Param name path string true "Template name"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/email-templates/{name} [get]
func (h *EmailTemplateHandler) GetEmailTemplate(c *fiber.Ctx) error {
	templateName := c.Params("name")
	if templateName == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Template name is required", nil)
	}

	accountIDPtr, organizationIDPtr, err := h.getEmailTemplateScope(c)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to resolve scope", nil)
	}

	locale := c.Query("locale", "en")
	template, err := h.emailTemplateQueries.GetEmailTemplate(c.Context(), templateName, locale, accountIDPtr, organizationIDPtr)
	if err != nil {
		log.Debug().Err(err).Str("template_name", templateName).Str("locale", locale).Msg("Email template not found")
		return webutil.Response(c, fiber.StatusNotFound, "Email template not found", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Email template retrieved successfully", map[string]interface{}{
		"template": template,
	})
}

// CreateEmailTemplate creates a new email template for the current organization or account
// @Summary Create email template
// @Description Creates a new email template for the current organization or account
// @Tags email-templates
// @Accept json
// @Produce json
// @Param body body models.EmailTemplateRequest true "Email template data"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/email-templates [post]
func (h *EmailTemplateHandler) CreateEmailTemplate(c *fiber.Ctx) error {
	accountIDPtr, organizationIDPtr, err := h.getEmailTemplateScope(c)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to resolve scope", nil)
	}

	var req models.EmailTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse email template request")
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.Name == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Template name is required", nil)
	}

	if req.Locale == "" {
		req.Locale = "en"
	}

	if req.Content == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Template content is required", nil)
	}

	existing, err := h.emailTemplateQueries.GetEmailTemplate(c.Context(), req.Name, req.Locale, accountIDPtr, organizationIDPtr)
	if err == nil && existing != nil {
		return webutil.Response(c, fiber.StatusConflict, "Template with this name and locale already exists", nil)
	}

	template := &models.EmailTemplate{
		Name:      req.Name,
		Locale:    req.Locale,
		Subject:   req.Subject,
		Content:   req.Content,
		IsSystem:  false,
		AccountID: accountIDPtr,
	}

	if err := h.emailTemplateQueries.CreateEmailTemplate(c.Context(), accountIDPtr, organizationIDPtr, template); err != nil {
		log.Error().Err(err).Msg("Failed to create email template")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create email template", nil)
	}

	return webutil.Response(c, fiber.StatusCreated, "Email template created successfully", map[string]interface{}{
		"template": template,
	})
}

// UpdateEmailTemplate updates an existing email template
// @Summary Update email template
// @Description Updates an existing email template (non-system templates only)
// @Tags email-templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param body body models.EmailTemplateRequest true "Email template data"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/email-templates/{id} [put]
func (h *EmailTemplateHandler) UpdateEmailTemplate(c *fiber.Ctx) error {
	templateID := c.Params("id")
	if templateID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Template ID is required", nil)
	}

	var req models.EmailTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse email template request")
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	// Validate input
	if req.Locale == "" {
		req.Locale = "en" // Default to English
	}
	if req.Content == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Template content is required", nil)
	}

	existing, err := h.emailTemplateQueries.GetEmailTemplateByID(c.Context(), templateID)
	if err != nil {
		log.Error().Err(err).Str("template_id", templateID).Msg("Failed to get email template")
		return webutil.Response(c, fiber.StatusNotFound, "Email template not found", nil)
	}

	if existing.IsSystem {
		return webutil.Response(c, fiber.StatusForbidden, "Cannot update system templates", nil)
	}

	if err := h.verifyTemplateOwnership(c, existing); err != nil {
		return err
	}

	// Prevent changing template name - use existing name from database
	// Changing the name could break email sending functionality
	// If name is provided in request, it must match existing name
	if req.Name != "" && req.Name != existing.Name {
		return webutil.Response(c, fiber.StatusBadRequest, "Cannot change template name", nil)
	}

	// Update template (always use existing name and locale, ignore from request)
	template := &models.EmailTemplate{
		Name:    existing.Name,    // Always use existing name to prevent breaking email sending
		Locale:  existing.Locale,  // Always use existing locale
		Subject: req.Subject,
		Content: req.Content,
	}

	if err := h.emailTemplateQueries.UpdateEmailTemplate(c.Context(), templateID, template); err != nil {
		log.Error().Err(err).Str("template_id", templateID).Msg("Failed to update email template")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to update email template", nil)
	}

	// Get updated template
	updated, err := h.emailTemplateQueries.GetEmailTemplateByID(c.Context(), templateID)
	if err != nil {
		log.Error().Err(err).Str("template_id", templateID).Msg("Failed to get updated email template")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get updated template", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Email template updated successfully", map[string]interface{}{
		"template": updated,
	})
}

// DeleteEmailTemplate deletes an email template
// @Summary Delete email template
// @Description Deletes an email template (non-system templates only)
// @Tags email-templates
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/email-templates/{id} [delete]
func (h *EmailTemplateHandler) DeleteEmailTemplate(c *fiber.Ctx) error {
	templateID := c.Params("id")
	if templateID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Template ID is required", nil)
	}

	existing, err := h.emailTemplateQueries.GetEmailTemplateByID(c.Context(), templateID)
	if err != nil {
		log.Error().Err(err).Str("template_id", templateID).Msg("Failed to get email template")
		return webutil.Response(c, fiber.StatusNotFound, "Email template not found", nil)
	}

	if existing.IsSystem {
		return webutil.Response(c, fiber.StatusForbidden, "Cannot delete system templates", nil)
	}

	if err := h.verifyTemplateOwnership(c, existing); err != nil {
		return err
	}

	if err := h.emailTemplateQueries.DeleteEmailTemplate(c.Context(), templateID); err != nil {
		log.Error().Err(err).Str("template_id", templateID).Msg("Failed to delete email template")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to delete email template", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Email template deleted successfully", nil)
}

// RegisterRoutes registers email template routes
func (h *EmailTemplateHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.GetAllEmailTemplates)
	router.Get("/:name", h.GetEmailTemplate)
	router.Post("/", h.CreateEmailTemplate)
	router.Put("/:id", h.UpdateEmailTemplate)
	router.Delete("/:id", h.DeleteEmailTemplate)
}
