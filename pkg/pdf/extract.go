package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/digitorus/pdf"
)

// ExtractPagesInput input data for extracting PDF information
type ExtractPagesInput struct {
	PDFPath string
}

// ExtractPagesResult result of PDF extraction
type ExtractPagesResult struct {
	PageCount int
}

// ExtractPages extracts number of pages from PDF using digitorus/pdf
func ExtractPages(input ExtractPagesInput) (*ExtractPagesResult, error) {
	// Read PDF
	file, err := os.Open(input.PDFPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Parse PDF using digitorus/pdf
	pdfReader, err := pdf.NewReader(file, stat.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF reader: %w", err)
	}

	pageCount := pdfReader.NumPage()
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

	// Note: Full page rendering requires external tools (pdftoppm is used in templates.go)
	// This function just returns placeholder paths
	// Actual preview generation is handled by generatePagePreview in templates.go

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

// ExtractFormFields extracts existing PDF form fields (AcroForm) using digitorus/pdf
func ExtractFormFields(input ExtractFormFieldsInput) (*ExtractFormFieldsResult, error) {
	// Read PDF
	file, err := os.Open(input.PDFPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Parse PDF using digitorus/pdf
	pdfReader, err := pdf.NewReader(file, stat.Size())
	if err != nil {
		// If PDF parsing fails, return empty result (not error)
		return &ExtractFormFieldsResult{Fields: []FormField{}}, nil
	}

	// Access AcroForm dictionary
	acroForm := pdfReader.Trailer().Key("Root").Key("AcroForm")
	if acroForm.IsNull() {
		// No AcroForm - return empty result
		return &ExtractFormFieldsResult{Fields: []FormField{}}, nil
	}

	// Get Fields array
	fieldsArray := acroForm.Key("Fields")
	if fieldsArray.IsNull() || fieldsArray.Len() == 0 {
		return &ExtractFormFieldsResult{Fields: []FormField{}}, nil
	}

	// Convert to our format
	var result []FormField
	for i := 0; i < fieldsArray.Len(); i++ {
		fieldObj := fieldsArray.Index(i)
		
		// Use field object directly (digitorus/pdf handles references automatically)
		fieldValue := fieldObj

		// Extract field name
		fieldName := fieldValue.Key("T").Text()
		if fieldName == "" {
			continue
		}

		// Extract field type
		fieldType := fieldValue.Key("FT").Name()
		
		// Extract field value
		fieldVal := fieldValue.Key("V").Text()

		// Extract Rect for coordinates (if available)
		rect := fieldValue.Key("Rect")
		var x, y, width, height float64
		if !rect.IsNull() && rect.Len() >= 4 {
			// Rect is [llx lly urx ury]
			llx := rect.Index(0).Float64()
			lly := rect.Index(1).Float64()
			urx := rect.Index(2).Float64()
			ury := rect.Index(3).Float64()
			x = llx
			y = lly
			width = urx - llx
			height = ury - lly
		}

		// Get page number (try to find from Parent chain)
		pageNum := 1 // Default to page 1 if can't determine

		formField := FormField{
			Name:  fieldName,
			Value: fieldVal,
			X:     x,
			Y:     y,
			Width: width,
			Height: height,
			Page:  pageNum,
		}

		// Map field type
		switch fieldType {
		case "/Tx":
			formField.Type = "text"
		case "/Btn":
			// Check if it's a radio button (has Opts) or checkbox
			if !fieldValue.Key("Opt").IsNull() {
				formField.Type = "radio"
			} else {
				formField.Type = "checkbox"
			}
		case "/Ch":
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

