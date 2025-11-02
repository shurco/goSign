package api

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/pdf"
	"github.com/shurco/gosign/pkg/utils/webutil"
)


// TemplateHandler handles requests to templates
type TemplateHandler struct {
	*ResourceHandler[models.Template] // embed generic CRUD
	templateQueries *queries.TemplateQueries
}

// NewTemplateHandler creates new handler
func NewTemplateHandler(repo ResourceRepository[models.Template], templateQueries *queries.TemplateQueries) *TemplateHandler {
	return &TemplateHandler{
		ResourceHandler: NewResourceHandler("template", repo),
		templateQueries: templateQueries,
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
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.TemplateID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "template_id is required", nil)
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
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.Name == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "name is required", nil)
	}

	if req.Type == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "type is required", nil)
	}

	if req.FileBase64 == "" && req.FileURL == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "file_base64 or file_url is required", nil)
	}

	// Decode file if base64
	var fileData []byte
	var err error
	if req.FileBase64 != "" {
		fileData, err = base64.StdEncoding.DecodeString(req.FileBase64)
		if err != nil {
			return webutil.Response(c, fiber.StatusBadRequest, "Invalid base64 data", nil)
		}
	} else {
		// TODO: Implement download from URL
		return webutil.Response(c, fiber.StatusBadRequest, "file_url not yet supported", nil)
	}

	// Get organization ID from context
	organizationID := ""
	if orgID := c.Locals("organization_id"); orgID != nil {
		if orgIDStr, ok := orgID.(string); ok {
			organizationID = orgIDStr
		}
	}

	// Process based on type
	var template *models.Template
	switch req.Type {
	case "pdf":
		template, err = h.processPDF(req.Name, req.Description, fileData, req.Settings, organizationID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to process PDF")
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to process PDF", map[string]any{
				"error": err.Error(),
			})
		}
	case "html", "docx":
		return webutil.Response(c, fiber.StatusNotImplemented, fmt.Sprintf("%s conversion not yet supported", req.Type), nil)
	default:
		return webutil.Response(c, fiber.StatusBadRequest, "Unsupported file type", nil)
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
func (h *TemplateHandler) processPDF(name, description string, fileData []byte, settings map[string]any, organizationID string) (*models.Template, error) {
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
		ID:             uuid.New().String(),
		Slug:           uuid.New().String(), // Generate unique slug
		OrganizationID: organizationID,
		Name:           name,
		Description:    description,
		Documents:      documents,
		Fields:         fields,
		Submitters:     []models.Submitter{}, // Empty, will be added by user
		Schema:         []models.Schema{},    // Empty, will be added by user
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
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

// SearchTemplates searches templates with filters
// @Summary Search templates
// @Description Search templates with filters, pagination and sorting
// @Tags templates
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param category query string false "Category filter"
// @Param tags query []string false "Tags filter (comma-separated)"
// @Param favorites query bool false "Show only favorites"
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset for pagination"
// @Param sort_by query string false "Sort by field (name, created_at, updated_at)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Success 200 {object} queries.TemplateSearchResult
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/templates/search [get]
func (h *TemplateHandler) SearchTemplates(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Get organization ID from context (optional)
	organizationID, _ := GetOrganizationID(c)

	// Parse query parameters
	req := queries.TemplateSearchRequest{
		Query:         c.Query("query"),
		Category:      c.Query("category"),
		OrganizationID: organizationID,
		UserID:        userID,
		SortBy:        c.Query("sort_by"),
		SortOrder:     c.Query("sort_order"),
	}

	// Parse tags (comma-separated)
	if tagsStr := c.Query("tags"); tagsStr != "" {
		req.Tags = strings.Split(tagsStr, ",")
		for i, tag := range req.Tags {
			req.Tags[i] = strings.TrimSpace(tag)
		}
	}

	// Parse favorites flag
	if favoritesStr := c.Query("favorites"); favoritesStr != "" {
		req.Favorites = favoritesStr == "true"
	}

	// Parse pagination
	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	req.Limit = limit

	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}
	req.Offset = offset

	// Search templates
	result, err := h.templateQueries.SearchTemplates(c.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to search templates")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to search templates", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "templates", result)
}

// GetUserFavorites returns user's favorite templates
// @Summary Get user favorites
// @Description Get user's favorite templates with pagination
// @Tags templates
// @Produce json
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/templates/favorites [get]
func (h *TemplateHandler) GetUserFavorites(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Parse pagination
	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get favorites
	templates, err := h.templateQueries.GetUserFavoriteTemplates(c.Context(), userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user favorites")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get favorites", nil)
	}

	result := map[string]interface{}{
		"templates": templates,
		"total":     len(templates), // Note: This is approximate since we don't count total
		"limit":     limit,
		"offset":    offset,
	}

	return webutil.Response(c, fiber.StatusOK, "favorites", result)
}

// AddToFavorites adds template to user's favorites
// @Summary Add to favorites
// @Description Add template to user's favorites
// @Tags templates
// @Accept json
// @Produce json
// @Param template_id body string true "Template ID"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Router /api/v1/templates/favorites [post]
func (h *TemplateHandler) AddToFavorites(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Parse request body
	var req struct {
		TemplateID string `json:"template_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.TemplateID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "template_id is required", nil)
	}

	// Add to favorites
	err = h.templateQueries.AddTemplateFavorite(c.Context(), req.TemplateID, userID)
	if err != nil {
		log.Error().Err(err).Str("template_id", req.TemplateID).Str("user_id", userID).Msg("Failed to add to favorites")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to add to favorites", nil)
	}

	return webutil.Response(c, fiber.StatusCreated, "Added to favorites", map[string]string{
		"template_id": req.TemplateID,
		"user_id":     userID,
	})
}

// RemoveFromFavorites removes template from user's favorites
// @Summary Remove from favorites
// @Description Remove template from user's favorites
// @Tags templates
// @Param template_id path string true "Template ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/templates/favorites/{template_id} [delete]
func (h *TemplateHandler) RemoveFromFavorites(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	templateID := c.Params("template_id")
	if templateID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "template_id is required", nil)
	}

	// Remove from favorites
	err = h.templateQueries.RemoveTemplateFavorite(c.Context(), templateID, userID)
	if err != nil {
		log.Error().Err(err).Str("template_id", templateID).Str("user_id", userID).Msg("Failed to remove from favorites")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to remove from favorites", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Removed from favorites", map[string]string{
		"template_id": templateID,
		"user_id":     userID,
	})
}

// CreateFolder creates a new template folder
// @Summary Create template folder
// @Description Create a new folder for organizing templates
// @Tags templates
// @Accept json
// @Produce json
// @Param body body map[string]string true "Folder data (name, parent_id)"
// @Success 201 {object} models.TemplateFolder
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Router /api/v1/templates/folders [post]
func (h *TemplateHandler) CreateFolder(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Parse request body
	var req struct {
		Name     string  `json:"name" validate:"required"`
		ParentID *string `json:"parent_id,omitempty"`
	}
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.Name == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "name is required", nil)
	}

	// Create folder
	folder := &models.TemplateFolder{
		ID:        uuid.New().String(),
		Name:      req.Name,
		ParentID:  req.ParentID, // Will be ignored for now (column doesn't exist)
		UserID:    userID,       // Will be used to get account_id
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = h.templateQueries.CreateTemplateFolder(c.Context(), userID, folder)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create template folder")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create folder", nil)
	}

	return webutil.Response(c, fiber.StatusCreated, "folder", folder)
}

// GetFolders returns user's template folders
// @Summary Get template folders
// @Description Get all folders for organizing templates
// @Tags templates
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Router /api/v1/templates/folders [get]
func (h *TemplateHandler) GetFolders(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	folders, err := h.templateQueries.GetTemplateFolders(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get template folders")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get folders", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "folders", folders)
}

// UpdateFolder updates a template folder
// @Summary Update template folder
// @Description Update folder name or parent folder
// @Tags templates
// @Accept json
// @Produce json
// @Param folder_id path string true "Folder ID"
// @Param body body map[string]string true "Folder data (name, parent_id)"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /api/v1/templates/folders/{folder_id} [put]
func (h *TemplateHandler) UpdateFolder(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	folderID := c.Params("folder_id")
	if folderID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "folder_id is required", nil)
	}

	// Parse request body
	var req struct {
		Name     string  `json:"name" validate:"required"`
		ParentID *string `json:"parent_id,omitempty"`
	}
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.Name == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "name is required", nil)
	}

	err = h.templateQueries.UpdateTemplateFolder(c.Context(), folderID, userID, req.Name, req.ParentID)
	if err != nil {
		log.Error().Err(err).Str("folder_id", folderID).Msg("Failed to update template folder")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to update folder", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Folder updated", map[string]string{
		"folder_id": folderID,
	})
}

// DeleteFolder deletes a template folder
// @Summary Delete template folder
// @Description Delete folder and move templates to root
// @Tags templates
// @Param folder_id path string true "Folder ID"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Router /api/v1/templates/folders/{folder_id} [delete]
func (h *TemplateHandler) DeleteFolder(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	folderID := c.Params("folder_id")
	if folderID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "folder_id is required", nil)
	}

	err = h.templateQueries.DeleteTemplateFolder(c.Context(), folderID, userID)
	if err != nil {
		log.Error().Err(err).Str("folder_id", folderID).Msg("Failed to delete template folder")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to delete folder", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Folder deleted", map[string]string{
		"folder_id": folderID,
	})
}

// MoveTemplate moves a template to a different folder
// @Summary Move template to folder
// @Description Move template to a different folder or root
// @Tags templates
// @Accept json
// @Produce json
// @Param template_id path string true "Template ID"
// @Param body body map[string]string true "Move data (folder_id)"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Router /api/v1/templates/{template_id}/move [put]
func (h *TemplateHandler) MoveTemplate(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	templateID := c.Params("template_id")
	if templateID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "template_id is required", nil)
	}

	// Parse request body
	var req struct {
		FolderID string `json:"folder_id,omitempty"` // Empty string means move to root
	}
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	err = h.templateQueries.MoveTemplateToFolder(c.Context(), templateID, req.FolderID, userID)
	if err != nil {
		log.Error().Err(err).Str("template_id", templateID).Msg("Failed to move template")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to move template", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Template moved", map[string]string{
		"template_id": templateID,
		"folder_id":   req.FolderID,
	})
}


// RegisterRoutes registers all routes for templates
func (h *TemplateHandler) RegisterRoutes(router fiber.Router) {
	// IMPORTANT: Register specific routes BEFORE generic CRUD routes
	// Otherwise, routes like /search will be matched by /:id pattern
	
	// Library operations (must be before /:id)
	router.Get("/search", h.SearchTemplates)
	router.Get("/favorites", h.GetUserFavorites)
	router.Post("/favorites", h.AddToFavorites)
	router.Delete("/favorites/:template_id", h.RemoveFromFavorites)

	// Folder operations (must be before /:id)
	router.Get("/folders", h.GetFolders)
	router.Post("/folders", h.CreateFolder)
	router.Put("/folders/:folder_id", h.UpdateFolder)
	router.Delete("/folders/:folder_id", h.DeleteFolder)

	// Specific operations (must be before /:id)
	router.Post("/clone", h.Clone)
	router.Post("/from-file", h.CreateFromType)

	// Template movement (specific pattern before /:id)
	router.Put("/:template_id/move", h.MoveTemplate)

	// Generic CRUD routes (register LAST, as they use catch-all /:id pattern)
	h.ResourceHandler.RegisterRoutes(router)
}

