package api

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/pdf"
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

	// Decode file if base64
	var fileData []byte
	var err error
	if req.FileBase64 != "" {
		fileData, err = base64.StdEncoding.DecodeString(req.FileBase64)
		if err != nil {
			return webutil.StatusBadRequest(c, "Invalid base64 data")
		}
	} else {
		// TODO: Implement download from URL
		return webutil.StatusBadRequest(c, "file_url not yet supported")
	}

	// Process based on type
	var template *models.Template
	switch req.Type {
	case "pdf":
		template, err = h.processPDF(req.Name, req.Description, fileData, req.Settings)
		if err != nil {
			log.Error().Err(err).Msg("Failed to process PDF")
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to process PDF", map[string]any{
				"error": err.Error(),
			})
		}
	case "html", "docx":
		return webutil.Response(c, fiber.StatusNotImplemented, fmt.Sprintf("%s conversion not yet supported", req.Type), nil)
	default:
		return webutil.StatusBadRequest(c, "Unsupported file type")
	}

	// Save template
	if err := h.repository.Create(template); err != nil {
		log.Error().Err(err).Msg("Failed to create template from file")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create template", nil)
	}

	log.Info().Str("template_id", template.ID).Str("name", template.Name).Msg("Template created from file")
	return webutil.Response(c, fiber.StatusCreated, "template", template)
}

// processPDF processes PDF file and creates template
func (h *TemplateHandler) processPDF(name, description string, fileData []byte, settings map[string]any) (*models.Template, error) {
	// Save PDF to temporary location
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("template_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	if err := os.WriteFile(tmpFile, fileData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	// 1. Extract pages
	pagesResult, err := pdf.ExtractPages(pdf.ExtractPagesInput{
		PDFPath: tmpFile,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to extract pages: %w", err)
	}

	// 2. Generate preview images
	previewDir := filepath.Join(tmpDir, fmt.Sprintf("previews_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create preview dir: %w", err)
	}
	defer os.RemoveAll(previewDir)

	previewResult, err := pdf.GeneratePreview(pdf.GeneratePreviewInput{
		PDFPath:   tmpFile,
		OutputDir: previewDir,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate previews, continuing without them")
	}

	// 3. Extract form fields (if any)
	formFieldsResult, err := pdf.ExtractFormFields(pdf.ExtractFormFieldsInput{
		PDFPath: tmpFile,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract form fields, continuing without them")
	}

	// 4. Build documents array
	documents := make([]models.Document, pagesResult.PageCount)
	for i := 0; i < pagesResult.PageCount; i++ {
		doc := models.Document{
			ID:        fmt.Sprintf("doc_%d", i),
			FileName:  fmt.Sprintf("page_%d.pdf", i+1),
			CreatedAt: time.Now(),
		}

		// Add preview images if available
		if previewResult != nil && i < len(previewResult.Images) {
			doc.PreviewImages = []models.PreviewImages{
				{
					ID:       fmt.Sprintf("preview_%d", i),
					FileName: fmt.Sprintf("%s_%d.jpg", uuid.New().String(), i),
				},
			}
		}

		documents[i] = doc
	}

	// 5. Build fields array from extracted form fields
	var fields []models.Field
	if formFieldsResult != nil && len(formFieldsResult.Fields) > 0 {
		for _, ff := range formFieldsResult.Fields {
			// Convert string type to FieldType
			fieldType := models.FieldTypeText
			switch ff.Type {
			case "checkbox":
				fieldType = models.FieldTypeCheckbox
			case "radio":
				fieldType = models.FieldTypeRadio
			case "select":
				fieldType = models.FieldTypeSelect
			default:
				fieldType = models.FieldTypeText
			}

			field := models.Field{
				ID:       uuid.New().String(),
				Name:     ff.Name,
				Type:     fieldType,
				Required: ff.Required,
			}
			fields = append(fields, field)
		}
	}

	// 6. Create template
	template := &models.Template{
		ID:          uuid.New().String(),
		Slug:        uuid.New().String(), // Generate unique slug
		Name:        name,
		Description: description,
		Documents:   documents,
		Fields:      fields,
		Submitters:  []models.Submitter{}, // Empty, will be added by user
		Schema:      []models.Schema{},    // Empty, will be added by user
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set settings if provided
	if settings != nil {
		expirationDays := 30 // default
		if days, ok := settings["expiration_days"].(float64); ok {
			expirationDays = int(days)
		} else if days, ok := settings["expiration_days"].(int); ok {
			expirationDays = days
		}

		template.Settings = &models.TemplateSettings{
			EmbeddingEnabled: settings["embedding_enabled"] == true,
			WebhookEnabled:   settings["webhook_enabled"] == true,
			ReminderEnabled:  settings["reminder_enabled"] == true,
			ExpirationDays:   expirationDays,
		}
	}

	return template, nil
}

// RegisterRoutes registers all routes for templates
func (h *TemplateHandler) RegisterRoutes(router fiber.Router) {
	// Generic CRUD routes
	h.ResourceHandler.RegisterRoutes(router)

	// Specific operations
	router.Post("/clone", h.Clone)
	router.Post("/from-file", h.CreateFromType)
}

