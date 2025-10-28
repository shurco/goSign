package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// TemplateHandler handles requests to templates
type TemplateHandler struct {
	*ResourceHandler[models.Template] // embed generic CRUD
	// TODO: add templateRepository for specific operations
}

// NewTemplateHandler creates new handler
func NewTemplateHandler(repo ResourceRepository[models.Template]) *TemplateHandler {
	return &TemplateHandler{
		ResourceHandler: NewResourceHandler("template", repo),
	}
}

// CloneRequest request body for cloning template
type CloneRequest struct {
	TemplateID string `json:"template_id" validate:"required"`
	Name       string `json:"name,omitempty"`
}

// Clone clones existing template
// @Summary Clone template
// @Description Creates a copy of an existing template
// @Tags templates
// @Accept json
// @Produce json
// @Param body body CloneRequest true "Clone request"
// @Success 201 {object} models.Template
// @Failure 400 {object} map[string]any
// @Router /api/templates/clone [post]
func (h *TemplateHandler) Clone(c *fiber.Ctx) error {
	var req CloneRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, "Invalid request body")
	}

	if req.TemplateID == "" {
		return webutil.StatusBadRequest(c, "template_id is required")
	}

	// Get source template
	original, err := h.repository.Get(req.TemplateID)
	if err != nil {
		log.Error().Err(err).Str("template_id", req.TemplateID).Msg("Failed to get template")
		return webutil.Response(c, fiber.StatusNotFound, "Template not found", nil)
	}

	// Create copy
	cloned := *original
	cloned.ID = uuid.New().String()
	if req.Name != "" {
		cloned.Name = req.Name
	} else {
		cloned.Name = original.Name + " (Copy)"
	}

	// Save clone
	if err := h.repository.Create(&cloned); err != nil {
		log.Error().Err(err).Msg("Failed to clone template")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to clone template", nil)
	}

	return webutil.Response(c, fiber.StatusCreated, "template", cloned)
}

// CreateFromTypeRequest request body for creating template from file
type CreateFromTypeRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Type        string                 `json:"type" validate:"required,oneof=pdf html docx"`
	FileBase64  string                 `json:"file_base64,omitempty"`
	FileURL     string                 `json:"file_url,omitempty"`
	Description string                 `json:"description,omitempty"`
	Settings    map[string]any `json:"settings,omitempty"`
}

// CreateFromType creates template from file of specific type
// @Summary Create template from file
// @Description Creates a template from PDF, HTML, or DOCX file
// @Tags templates
// @Accept json
// @Produce json
// @Param body body CreateFromTypeRequest true "Create from type request"
// @Success 201 {object} models.Template
// @Failure 400 {object} map[string]any
// @Router /api/templates/from-file [post]
func (h *TemplateHandler) CreateFromType(c *fiber.Ctx) error {
	var req CreateFromTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, "Invalid request body")
	}

	if req.Name == "" {
		return webutil.StatusBadRequest(c, "name is required")
	}

	if req.Type == "" {
		return webutil.StatusBadRequest(c, "type is required")
	}

	if req.FileBase64 == "" && req.FileURL == "" {
		return webutil.StatusBadRequest(c, "file_base64 or file_url is required")
	}

	// TODO: Process file based on type
	// - PDF: extract fields and images
	// - HTML: convert to PDF
	// - DOCX: convert to PDF

	// Create template
	template := &models.Template{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		// TODO: set remaining fields after file processing
	}

	if err := h.repository.Create(template); err != nil {
		log.Error().Err(err).Msg("Failed to create template from file")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create template", nil)
	}

	return webutil.Response(c, fiber.StatusCreated, "template", template)
}

// RegisterRoutes registers all routes for templates
func (h *TemplateHandler) RegisterRoutes(router fiber.Router) {
	// Generic CRUD routes
	h.ResourceHandler.RegisterRoutes(router)

	// Specific operations
	router.Post("/clone", h.Clone)
	router.Post("/from-file", h.CreateFromType)
}

