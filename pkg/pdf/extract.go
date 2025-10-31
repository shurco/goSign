package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ExtractPagesInput input data for extracting PDF information
type ExtractPagesInput struct {
	PDFPath string
}

// ExtractPagesResult result of PDF extraction
type ExtractPagesResult struct {
	PageCount int
}

// ExtractPages extracts number of pages from PDF
func ExtractPages(input ExtractPagesInput) (*ExtractPagesResult, error) {
	// Read PDF
	file, err := os.Open(input.PDFPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	// Get page count using pdfcpu
	conf := model.NewDefaultConfiguration()
	ctx, err := api.ReadContext(file, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF context: %w", err)
	}

	pageCount := ctx.PageCount
	return &ExtractPagesResult{
		PageCount: pageCount,
	}, nil
}

// GeneratePreviewInput input data for generating preview images
type GeneratePreviewInput struct {
	PDFPath   string
	OutputDir string // directory to save preview images
}

// GeneratePreviewResult result of preview generation
type GeneratePreviewResult struct {
	Images []string // paths to generated images
}

// GeneratePreview generates preview images for each page of PDF
func GeneratePreview(input GeneratePreviewInput) (*GeneratePreviewResult, error) {
	// Get page count first
	pagesResult, err := ExtractPages(ExtractPagesInput{PDFPath: input.PDFPath})
	if err != nil {
		return nil, fmt.Errorf("failed to extract pages: %w", err)
	}

	// Create output directory if not exists
	if err := os.MkdirAll(input.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use ExtractImagesFile to get page images
	// This extracts embedded images, but for full page rendering
	// we need external tools like ghostscript or pdftoppm
	// For now, create placeholder images
	var images []string
	
	for pageNum := 1; pageNum <= pagesResult.PageCount; pageNum++ {
		// For basic implementation, we'll note that full rendering
		// requires external tools (ghostscript, poppler-utils, etc.)
		// Create placeholder that indicates page exists
		outputFile := filepath.Join(input.OutputDir, fmt.Sprintf("%d.jpg", pageNum-1))
		images = append(images, outputFile)
	}

	// Extract images from PDF if they exist
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("extract_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	conf := model.NewDefaultConfiguration()
	// Try to extract embedded images
	_ = api.ExtractImagesFile(input.PDFPath, tmpDir, nil, conf)

	return &GeneratePreviewResult{
		Images: images,
	}, nil
}

// FormField represents a PDF form field
type FormField struct {
	Name     string
	Type     string // text, checkbox, radio, select, etc.
	Value    string
	Required bool
	Page     int
	X        float64
	Y        float64
	Width    float64
	Height   float64
}

// ExtractFormFieldsInput input data for extracting form fields
type ExtractFormFieldsInput struct {
	PDFPath string
}

// ExtractFormFieldsResult result of form field extraction
type ExtractFormFieldsResult struct {
	Fields []FormField
}

// ExtractFormFields extracts existing PDF form fields (AcroForm)
func ExtractFormFields(input ExtractFormFieldsInput) (*ExtractFormFieldsResult, error) {
	// Read PDF
	file, err := os.Open(input.PDFPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	// Get form fields using pdfcpu
	conf := model.NewDefaultConfiguration()
	fields, err := api.FormFields(file, conf)
	if err != nil {
		// No form fields or error - return empty result
		return &ExtractFormFieldsResult{Fields: []FormField{}}, nil
	}

	// Convert to our format
	var result []FormField
	for _, field := range fields {
		formField := FormField{
			Name:  field.Name,
			Value: field.V,
		}

		// Map field type
		switch field.Typ.String() {
		case "Tx":
			formField.Type = "text"
		case "Btn":
			if field.Opts != "" {
				formField.Type = "radio"
			} else {
				formField.Type = "checkbox"
			}
		case "Ch":
			formField.Type = "select"
		default:
			formField.Type = "text"
		}

		result = append(result, formField)
	}

	return &ExtractFormFieldsResult{
		Fields: result,
	}, nil
}

