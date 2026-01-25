package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/pkg/appdir"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/services/field"
	"github.com/shurco/gosign/internal/services/formula"
	"github.com/shurco/gosign/pkg/pdf"
	"github.com/shurco/gosign/pkg/utils/webutil"
	"github.com/signintech/gopdf"
)

// TemplateHandler handles requests to templates
type TemplateHandler struct {
	*ResourceHandler[models.Template] // embed generic CRUD
	templateQueries                   *queries.TemplateQueries
}

// NewTemplateHandler creates new handler
func NewTemplateHandler(repo ResourceRepository[models.Template], templateQueries *queries.TemplateQueries) *TemplateHandler {
	return &TemplateHandler{
		ResourceHandler: NewResourceHandler("template", repo),
		templateQueries: templateQueries,
	}
}

// CreateEmptyTemplateRequest request body for creating empty template
type CreateEmptyTemplateRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
}

// CreateEmptyTemplate creates an empty template
// @Summary Create empty template
// @Description Creates a new empty template with default structure
// @Tags templates
// @Accept json
// @Produce json
// @Param body body CreateEmptyTemplateRequest true "Create empty template request"
// @Success 201 {object} models.Template
// @Failure 400 {object} map[string]any
// @Router /api/templates/empty [post]
func (h *TemplateHandler) CreateEmptyTemplate(c *fiber.Ctx) error {
	var req CreateEmptyTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.Name == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "name is required", nil)
	}

	// Get organization ID from context
	organizationID := ""
	if orgID := c.Locals("organization_id"); orgID != nil {
		if orgIDStr, ok := orgID.(string); ok {
			organizationID = orgIDStr
		}
	}

	// Create empty template with all required fields
	template := &models.Template{
		ID:             uuid.New().String(),
		Slug:           uuid.New().String(),
		OrganizationID: organizationID,
		Name:           req.Name,
		Source:         "web", // Required field - source of template creation
		Submitters:     []models.Submitter{},
		Fields:         []models.Field{},
		Schema:         []models.Schema{},
		Documents:      []models.Document{},
		Settings:       &models.TemplateSettings{}, // Required field - initialize with default settings
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Set category if provided
	if req.Category != nil && *req.Category != "" {
		template.Category = *req.Category
	}

	// Save template using templateQueries (repository.Create is not implemented)
	if err := h.templateQueries.CreateTemplate(c.Context(), template); err != nil {
		log.Error().Err(err).Msg("Failed to create empty template")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create template", nil)
	}

	return webutil.Response(c, fiber.StatusCreated, "template", template)
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
	Name        string         `json:"name" validate:"required"`
	Type        string         `json:"type" validate:"required,oneof=pdf html docx"`
	FileBase64  string         `json:"file_base64,omitempty"`
	FileURL     string         `json:"file_url,omitempty"`
	Description string         `json:"description,omitempty"`
	Category    *string        `json:"category,omitempty"`
	Settings    map[string]any `json:"settings,omitempty"`
}

// AttachFileToTemplateRequest request body for attaching a file to an existing template
type AttachFileToTemplateRequest struct {
	Type       string `json:"type" validate:"required,oneof=pdf"`
	FileBase64 string `json:"file_base64" validate:"required"`
	Append     bool   `json:"append,omitempty"`
}

// AttachFileToTemplate attaches a file to an existing template (e.g., import PDF pages).
// Intended for:
// - "empty template" flow (append=false): set initial pages
// - "add pages" flow (append=true): append new pages to existing schema
//
// @Summary Attach file to template
// @Description Attaches a PDF file to an existing template and updates its schema/documents
// @Tags templates
// @Accept json
// @Produce json
// @Param template_id path string true "Template ID"
// @Param body body AttachFileToTemplateRequest true "Attach file request"
// @Success 200 {object} models.Template
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /api/templates/{template_id}/from-file [post]
func (h *TemplateHandler) AttachFileToTemplate(c *fiber.Ctx) error {
	templateID := c.Params("template_id")
	if templateID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "template_id is required", nil)
	}

	var req AttachFileToTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.Type == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "type is required", nil)
	}
	if req.Type != "pdf" {
		return webutil.Response(c, fiber.StatusBadRequest, "Unsupported file type", nil)
	}
	if req.FileBase64 == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "file_base64 is required", nil)
	}

	// Decode file
	fileData, err := base64.StdEncoding.DecodeString(req.FileBase64)
	if err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid base64 data", nil)
	}
	if len(fileData) == 0 {
		return webutil.Response(c, fiber.StatusBadRequest, "file data is empty", nil)
	}

	// Ensure template exists (and get current name)
	existing, err := h.repository.Get(templateID)
	if err != nil || existing == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Template not found", nil)
	}

	// Get organization ID from context (used by storage paths/providers in future)
	organizationID := ""
	if orgID := c.Locals("organization_id"); orgID != nil {
		if orgIDStr, ok := orgID.(string); ok {
			organizationID = orgIDStr
		}
	}

	// Determine base schema (append mode keeps existing pages)
	var baseSchema []models.Schema
	if req.Append {
		// Make a copy to avoid mutating the original slice reference
		baseSchema = append([]models.Schema{}, existing.Schema...)
	}

	// Save PDF pages + previews and update schema
	if err := h.savePDFToStorageWithBaseSchema(c.Context(), templateID, existing.Name, fileData, organizationID, baseSchema); err != nil {
		log.Error().Err(err).Str("template_id", templateID).Msg("Failed to attach PDF to template")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to attach file to template", map[string]any{
			"error": err.Error(),
		})
	}

	// Return updated template with documents populated
	updated, err := h.templateQueries.Template(c.Context(), templateID)
	if err != nil {
		log.Error().Err(err).Str("template_id", templateID).Msg("Failed to load updated template after attach")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to load updated template", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "template", updated)
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
	var pdfFileData []byte // Keep file data for saving to storage after template creation
	switch req.Type {
	case "pdf":
		pdfFileData = fileData // Save file data for later
		template, err = h.processPDF(c.Context(), req.Name, req.Description, fileData, req.Settings, organizationID, req.Category)
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

	// Save template using templateQueries (repository.Create is not implemented)
	if err := h.templateQueries.CreateTemplate(c.Context(), template); err != nil {
		log.Error().Err(err).Msg("Failed to create template from file")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create template from file", nil)
	}

	// Save PDF file to storage and create database records (now that we have template ID)
	if req.Type == "pdf" && len(pdfFileData) > 0 {
		if err := h.savePDFToStorage(c.Context(), template.ID, req.Name, pdfFileData, organizationID); err != nil {
			log.Error().Err(err).Str("template_id", template.ID).Msg("Failed to save PDF to storage")
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save PDF to storage", map[string]any{
				"error": err.Error(),
			})
		}
	}

	return webutil.Response(c, fiber.StatusCreated, "template", template)
}

// processPDF processes a PDF file and creates a template from it.
// It extracts form fields from the PDF, converts them to template fields,
// and creates a template structure. The actual file storage and page extraction
// are handled separately in savePDFToStorage after the template is created.
func (h *TemplateHandler) processPDF(ctx context.Context, name, description string, fileData []byte, settings map[string]any, organizationID string, category *string) (*models.Template, error) {
	// Save PDF to temporary location
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("template_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	if err := os.WriteFile(tmpFile, fileData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	// 1. Extract form fields (if any)
	// Note: Page extraction and preview generation are done in savePDFToStorage
	// after template is created, so we have the template ID for storage_attachment
	formFieldsResult, err := pdf.ExtractFormFields(pdf.ExtractFormFieldsInput{
		PDFPath: tmpFile,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract form fields, continuing without them")
	}

	// 4. Build documents array - will be populated after template is created and file is saved to storage
	// Documents will be saved in savePDFToStorage function
	documents := []models.Document{}
	schema := []models.Schema{}

	// 6. Build fields array from extracted form fields
	fields := []models.Field{} // Initialize as empty slice, not nil
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

	// 7. Create template
	template := &models.Template{
		ID:             uuid.New().String(),
		Slug:           uuid.New().String(), // Generate unique slug
		OrganizationID: organizationID,
		Name:           name,
		Description:    description,
		Source:         "pdf", // Source is PDF file upload
		Documents:      documents,
		Fields:         fields,
		Submitters:     []models.Submitter{}, // Empty, will be added by user
		Schema:         schema,
		Settings:       &models.TemplateSettings{}, // Initialize with default settings
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Set category if provided
	if category != nil && *category != "" {
		template.Category = *category
	}

	// Override settings if provided
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

// savePDFToStorage splits a PDF into individual pages and saves each page to the lc_pages directory.
// For each page, it creates:
// - lc_pages/{attachment_id}/0.pdf - the PDF page file
// - lc_pages/{attachment_id}/0.jpg - the full preview image
// - lc_pages/{attachment_id}/p/0.jpg - the thumbnail preview image
// It also creates storage_attachment and storage_blob records in the database.
func (h *TemplateHandler) savePDFToStorage(ctx context.Context, templateID, name string, fileData []byte, organizationID string) error {
	return h.savePDFToStorageWithBaseSchema(ctx, templateID, name, fileData, organizationID, nil)
}

// savePDFToStorageWithBaseSchema stores pages and sets schema to baseSchema + newPagesSchema.
// If baseSchema is empty/nil, it behaves like "replace schema" for initial upload.
func (h *TemplateHandler) savePDFToStorageWithBaseSchema(
	ctx context.Context,
	templateID, name string,
	fileData []byte,
	organizationID string,
	baseSchema []models.Schema,
) error {
	newSchemaItems, err := h.storePDFPagesToStorage(ctx, templateID, name, fileData, organizationID)
	if err != nil {
		return err
	}

	combined := append([]models.Schema{}, baseSchema...)
	combined = append(combined, newSchemaItems...)

	if err := h.templateQueries.UpdateTemplateSchema(ctx, templateID, combined); err != nil {
		return fmt.Errorf("failed to update template schema: %w", err)
	}

	return nil
}

// storePDFPagesToStorage splits the PDF into pages, writes them to lc_pages, creates storage records,
// and returns schema items for the newly-added pages (does NOT update template.schema).
func (h *TemplateHandler) storePDFPagesToStorage(ctx context.Context, templateID, name string, fileData []byte, organizationID string) ([]models.Schema, error) {
	// Save PDF to temporary location
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("template_%s_%d.pdf", templateID, time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	if err := os.WriteFile(tmpFile, fileData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	// Ensure lc_pages directory exists (next to executable, same as app.go)
	lcPagesDir := appdir.LcPages()
	if err := os.MkdirAll(lcPagesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create lc_pages directory: %w", err)
	}

	var schema []models.Schema

	// Get page count first
	pageResult, err := pdf.ExtractPages(pdf.ExtractPagesInput{PDFPath: tmpFile})
	if err != nil {
		return nil, fmt.Errorf("failed to extract page count: %w", err)
	}
	pageCount := pageResult.PageCount
	if pageCount == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	// Extract all pages to temporary directory using gopdf
	tmpPagesDir := filepath.Join(os.TempDir(), fmt.Sprintf("pages_%s_%d", templateID, time.Now().UnixNano()))
	defer os.RemoveAll(tmpPagesDir)
	if err := os.MkdirAll(tmpPagesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp pages dir: %w", err)
	}

	// Extract each page as a separate PDF file using gopdf
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		// Create new PDF for this single page
		pagePdf := gopdf.GoPdf{}
		pagePdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
		pagePdf.AddPage()

		// Import the specific page from source PDF
		tpl := pagePdf.ImportPage(tmpFile, pageNum, "/MediaBox")
		pagePdf.UseImportedTemplate(tpl, 0, 0, 0, 0)

		// Save to temporary file
		pageFileName := fmt.Sprintf("page_%d.pdf", pageNum)
		pageFilePath := filepath.Join(tmpPagesDir, pageFileName)
		if err := pagePdf.WritePdf(pageFilePath); err != nil {
			log.Error().Err(err).Int("page", pageNum).Msg("Failed to extract page")
			return nil, fmt.Errorf("failed to extract page %d: %w", pageNum, err)
		}
	}

	// Preview images will be generated per-page from extracted PDFs
	previewDir := filepath.Join(os.TempDir(), fmt.Sprintf("previews_%s_%d", templateID, time.Now().UnixNano()))
	defer os.RemoveAll(previewDir)
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create preview dir: %w", err)
	}

	// Process each extracted page
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		attachmentID := uuid.New().String()
		pageDir := filepath.Join(lcPagesDir, attachmentID)
		if err := os.MkdirAll(pageDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create page directory: %w", err)
		}

		// Use extracted page PDF file
		extractedPagePath := filepath.Join(tmpPagesDir, fmt.Sprintf("page_%d.pdf", pageNum))

		// Copy extracted page to lc_pages/{attachment_id}/0.pdf
		pageData, err := os.ReadFile(extractedPagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read extracted page: %w", err)
		}

		pagePDFPath := filepath.Join(pageDir, "0.pdf")
		if err := os.WriteFile(pagePDFPath, pageData, 0644); err != nil {
			absPath, _ := filepath.Abs(pagePDFPath)
			log.Error().Err(err).Str("path", pagePDFPath).Str("abs_path", absPath).Msg("Failed to save page PDF")
			return nil, fmt.Errorf("failed to save page PDF to %s: %w", pagePDFPath, err)
		}

		// Generate preview image for this page from extracted PDF
		previewImagePath := filepath.Join(previewDir, fmt.Sprintf("%d.jpg", pageNum-1))
		var previewBlobID string

		// Generate preview from extracted page PDF
		if err := h.generatePagePreview(extractedPagePath, previewImagePath); err != nil {
			log.Warn().Err(err).Str("page_pdf", extractedPagePath).Int("page", pageNum).Msg("Failed to generate preview, continuing without preview")
		}

		if previewData, err := os.ReadFile(previewImagePath); err == nil {
			// Save full preview as 0.jpg
			previewPath := filepath.Join(pageDir, "0.jpg")
			if err := os.WriteFile(previewPath, previewData, 0644); err == nil {
				// Create storage_blob for preview
				previewBlobID = uuid.New().String()
				previewMetadata := map[string]any{"width": 1400, "height": 1980, "analyzed": true, "identified": true}
				if err := h.templateQueries.CreateStorageBlob(ctx, previewBlobID, "0.jpg", "image/jpeg", int64(len(previewData)), previewMetadata); err != nil {
					log.Warn().Err(err).Int("page", pageNum).Msg("Failed to create preview blob")
				}

				// Create small preview in p/ folder (thumbnail)
				pDir := filepath.Join(pageDir, "p")
				if err := os.MkdirAll(pDir, 0755); err == nil {
					thumbnailPath := filepath.Join(pDir, "0.jpg")
					if thumbnailData, err := createThumbnail(previewData); err == nil {
						_ = os.WriteFile(thumbnailPath, thumbnailData, 0644)
					}
				}
			}
		}

		// Create storage_attachment using preview blob (as in existing template)
		if previewBlobID != "" {
			if err := h.templateQueries.CreateStorageAttachment(ctx, attachmentID, previewBlobID, "Template", templateID, "documents", "disk"); err != nil {
				return nil, fmt.Errorf("failed to create storage_attachment: %w", err)
			}
		}

		// Add to schema
		schema = append(schema, models.Schema{
			AttachmentID: attachmentID,
			Name:         fmt.Sprintf("page_%d", pageNum),
		})
	}

	return schema, nil
}

// generatePagePreview generates a preview image from a PDF page using pdftoppm.
// It renders the first page of the PDF at 150 DPI and saves it as a JPEG image.
// The output image is saved to the specified outputPath.
// Requires pdftoppm utility from poppler-utils package to be installed.
func (h *TemplateHandler) generatePagePreview(pdfPath, outputPath string) error {
	// Check if pdftoppm is available
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return fmt.Errorf("pdftoppm not found: %w (install poppler-utils package)", err)
	}

	// Get page count using digitorus/pdf
	pageResult, err := pdf.ExtractPages(pdf.ExtractPagesInput{PDFPath: pdfPath})
	if err != nil {
		return fmt.Errorf("failed to get page count: %w", err)
	}
	if pageResult.PageCount < 1 {
		return fmt.Errorf("PDF has no pages")
	}

	// Create temporary directory for pdftoppm output
	tmpDir := filepath.Dir(outputPath)
	tmpPrefix := filepath.Join(tmpDir, fmt.Sprintf("preview_%d", time.Now().UnixNano()))

	// Use pdftoppm to convert first page to JPEG at 150 DPI
	// pdftoppm -jpeg -r 150 -f 1 -l 1 -singlefile input.pdf output_prefix
	// With -singlefile and single page, creates output_prefix.jpg (without page number)
	cmd := exec.Command("pdftoppm",
		"-jpeg",     // Output format: JPEG
		"-r", "150", // Resolution: 150 DPI
		"-f", "1", // First page: 1
		"-l", "1", // Last page: 1 (only first page)
		"-singlefile", // Output single file without page number suffix
		pdfPath,       // Input PDF
		tmpPrefix,     // Output prefix (will create tmpPrefix.jpg)
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run pdftoppm: %w (stderr: %s)", err, stderr.String())
	}

	// pdftoppm with -singlefile creates output with .jpg extension (no page number)
	generatedFile := tmpPrefix + ".jpg"

	// Check if file was created
	if _, err := os.Stat(generatedFile); os.IsNotExist(err) {
		// Try alternative naming (some versions might use -1.jpg even with -singlefile)
		altFile := tmpPrefix + "-1.jpg"
		if _, altErr := os.Stat(altFile); altErr == nil {
			generatedFile = altFile
		} else {
			return fmt.Errorf("pdftoppm did not create output file (tried: %s, %s)", generatedFile, altFile)
		}
	}

	// Move/rename to final output path
	if err := os.Rename(generatedFile, outputPath); err != nil {
		// If rename fails (different filesystems), try copy and remove
		data, readErr := os.ReadFile(generatedFile)
		if readErr != nil {
			return fmt.Errorf("failed to read generated file: %w (rename error: %v)", readErr, err)
		}
		if writeErr := os.WriteFile(outputPath, data, 0644); writeErr != nil {
			return fmt.Errorf("failed to write output file: %w (rename error: %v)", writeErr, err)
		}
		os.Remove(generatedFile) // Clean up temp file
	}

	return nil
}

// scaleBilinear scales src from srcRect into dst at dstRect using bilinear interpolation (stdlib only).
func scaleBilinear(dst *image.RGBA, dstRect image.Rectangle, src image.Image, srcRect image.Rectangle) {
	dx, dy := dstRect.Dx(), dstRect.Dy()
	sx0, sy0 := srcRect.Min.X, srcRect.Min.Y
	sw, sh := srcRect.Dx(), srcRect.Dy()
	if sw <= 0 || sh <= 0 {
		return
	}
	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			fx := float64(sx0) + (float64(x)+0.5)/float64(dx)*float64(sw) - 0.5
			fy := float64(sy0) + (float64(y)+0.5)/float64(dy)*float64(sh) - 0.5
			x0 := int(fx)
			y0 := int(fy)
			if x0 < srcRect.Min.X {
				x0 = srcRect.Min.X
			}
			if y0 < srcRect.Min.Y {
				y0 = srcRect.Min.Y
			}
			x1 := x0 + 1
			y1 := y0 + 1
			if x1 >= srcRect.Max.X {
				x1 = srcRect.Max.X - 1
			}
			if y1 >= srcRect.Max.Y {
				y1 = srcRect.Max.Y - 1
			}
			wx := fx - float64(x0)
			wy := fy - float64(y0)
			c00 := toRGBA(src.At(x0, y0))
			c10 := toRGBA(src.At(x1, y0))
			c01 := toRGBA(src.At(x0, y1))
			c11 := toRGBA(src.At(x1, y1))
			r := lerpU8(lerpF(c00.R, c10.R, wx), lerpF(c01.R, c11.R, wx), wy)
			g := lerpU8(lerpF(c00.G, c10.G, wx), lerpF(c01.G, c11.G, wx), wy)
			b := lerpU8(lerpF(c00.B, c10.B, wx), lerpF(c01.B, c11.B, wx), wy)
			a := lerpU8(lerpF(c00.A, c10.A, wx), lerpF(c01.A, c11.A, wx), wy)
			dst.SetRGBA(dstRect.Min.X+x, dstRect.Min.Y+y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
}

func toRGBA(c color.Color) struct{ R, G, B, A float64 } {
	r, g, b, a := c.RGBA()
	return struct{ R, G, B, A float64 }{
		R: float64(r >> 8),
		G: float64(g >> 8),
		B: float64(b >> 8),
		A: float64(a >> 8),
	}
}

func lerpF(a, b, t float64) float64 {
	return a*(1-t) + b*t
}

func lerpU8(a, b, t float64) uint8 {
	x := lerpF(a, b, t)
	if x <= 0 {
		return 0
	}
	if x >= 255 {
		return 255
	}
	return uint8(x + 0.5)
}

// createThumbnail creates a small thumbnail version of the image
func createThumbnail(imageData []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Calculate thumbnail size (max 200px width, maintain aspect ratio)
	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	maxWidth := 200
	thumbnailWidth := maxWidth
	thumbnailHeight := (srcHeight * maxWidth) / srcWidth
	if thumbnailHeight > 200 {
		thumbnailHeight = 200
		thumbnailWidth = (srcWidth * 200) / srcHeight
	}

	// Create thumbnail using stdlib bilinear scaling (no golang.org/x/image)
	thumbnail := image.NewRGBA(image.Rect(0, 0, thumbnailWidth, thumbnailHeight))
	scaleBilinear(thumbnail, thumbnail.Bounds(), img, srcBounds)

	// Encode to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
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
		Query:          c.Query("query"),
		Category:       c.Query("category"),
		OrganizationID: organizationID,
		UserID:         userID,
		SortBy:         c.Query("sort_by"),
		SortOrder:      c.Query("sort_order"),
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

// ValidateConditionsRequest request body for validating conditions
type ValidateConditionsRequest struct {
	Fields []models.Field `json:"fields" validate:"required"`
}

// ValidateConditions validates field conditions
// @Summary Validate conditions
// @Description Validates field conditions for logical errors
// @Tags templates
// @Accept json
// @Produce json
// @Param body body ValidateConditionsRequest true "Fields with conditions"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /api/v1/templates/:id/conditions/validate [post]
func (h *TemplateHandler) ValidateConditions(c *fiber.Ctx) error {
	var req ValidateConditionsRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := field.ValidateConditions(req.Fields); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Conditions valid", nil)
}

// ValidateFormulaRequest request body for validating formula
type ValidateFormulaRequest struct {
	Formula string         `json:"formula" validate:"required"`
	Fields  []models.Field `json:"fields" validate:"required"`
}

// ValidateFormula validates formula syntax
// @Summary Validate formula
// @Description Validates formula syntax and field references
// @Tags templates
// @Accept json
// @Produce json
// @Param body body ValidateFormulaRequest true "Formula validation request"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /api/v1/formulas/validate [post]
func (h *TemplateHandler) ValidateFormula(c *fiber.Ctx) error {
	var req ValidateFormulaRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := formula.ValidateFormula(req.Formula, req.Fields); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Formula valid", nil)
}

// UpdateTemplate handles partial updates for template editor.
// This endpoint exists because the generic CRUD update cannot distinguish
// between "field not provided" and "provided empty/null", which is required
// to safely persist editor changes (schema/fields/submitters) without wiping
// other columns unintentionally.
func (h *TemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	templateID := c.Params("id")
	if templateID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "ID is required", nil)
	}

	if h.templateQueries == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Template queries not initialized", nil)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &raw); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	patch := queries.TemplateUpdatePatch{}

	// name (optional)
	if b, ok := raw["name"]; ok {
		// allow empty string names? editor shouldn't send it, but keep validation in DB constraints
		var name string
		if err := json.Unmarshal(b, &name); err != nil {
			return webutil.Response(c, fiber.StatusBadRequest, "Invalid name", nil)
		}
		patch.Name = &name
	}

	// category (optional, supports null)
	if b, ok := raw["category"]; ok {
		patch.CategoryProvided = true
		if string(b) == "null" {
			patch.Category = nil
		} else {
			var category string
			if err := json.Unmarshal(b, &category); err != nil {
				return webutil.Response(c, fiber.StatusBadRequest, "Invalid category", nil)
			}
			// Empty string treated as NULL (clear)
			if category == "" {
				patch.Category = nil
			} else {
				patch.Category = &category
			}
		}
	}

	// schema (optional)
	if b, ok := raw["schema"]; ok {
		if string(b) == "null" {
			empty := []models.Schema{}
			patch.Schema = &empty
		} else {
			var schema []models.Schema
			if err := json.Unmarshal(b, &schema); err != nil {
				return webutil.Response(c, fiber.StatusBadRequest, "Invalid schema", nil)
			}
			patch.Schema = &schema
		}
	}

	// fields (optional)
	if b, ok := raw["fields"]; ok {
		if string(b) == "null" {
			empty := []models.Field{}
			patch.Fields = &empty
		} else {
			var fields []models.Field
			if err := json.Unmarshal(b, &fields); err != nil {
				return webutil.Response(c, fiber.StatusBadRequest, "Invalid fields", nil)
			}
			patch.Fields = &fields
		}
	}

	// submitters (optional)
	if b, ok := raw["submitters"]; ok {
		if string(b) == "null" {
			empty := []models.Submitter{}
			patch.Submitters = &empty
		} else {
			var submitters []models.Submitter
			if err := json.Unmarshal(b, &submitters); err != nil {
				return webutil.Response(c, fiber.StatusBadRequest, "Invalid submitters", nil)
			}
			patch.Submitters = &submitters
		}
	}

	if err := h.templateQueries.UpdateTemplatePatch(c.Context(), templateID, patch); err != nil {
		log.Error().Err(err).Str("template_id", templateID).Msg("Failed to update template")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to update template", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "template", map[string]any{"id": templateID})
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
	router.Post("/empty", h.CreateEmptyTemplate)
	router.Post("/clone", h.Clone)
	router.Post("/from-file", h.CreateFromType)
	router.Post("/:template_id/from-file", h.AttachFileToTemplate)

	// Condition validation (must be before /:id)
	router.Post("/:template_id/conditions/validate", h.ValidateConditions)

	// Formula validation (public endpoint, no template ID needed)
	router.Post("/formulas/validate", h.ValidateFormula)

	// Template movement (specific pattern before /:id)
	router.Put("/:template_id/move", h.MoveTemplate)

	// Generic CRUD routes:
	// Register explicitly to avoid duplicate PUT /:id (we override update for editor safety).
	router.Get("/", h.ResourceHandler.List)
	router.Get("/:id", h.ResourceHandler.Get)
	router.Post("/", h.ResourceHandler.Create)
	router.Delete("/:id", h.ResourceHandler.Delete)

	// Editor update (must be registered and must not be shadowed).
	router.Put("/:id", h.UpdateTemplate)
}
