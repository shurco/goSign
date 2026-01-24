package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/signintech/gopdf"

	"github.com/shurco/gosign/internal/models"
)

// FillFieldsInput input data for filling fields
type FillFieldsInput struct {
	PDFPath string
	Fields  map[string]string // field_name -> value
}

// FillFields fills PDF fields with values using gopdf
func FillFields(input FillFieldsInput) ([]byte, error) {
	// Read original PDF
	data, err := os.ReadFile(input.PDFPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	// If no fields to fill, return original
	if len(input.Fields) == 0 {
		return data, nil
	}

	// Parse form fields from PDF to get their positions
	fields, err := ExtractFormFields(ExtractFormFieldsInput{PDFPath: input.PDFPath})
	if err != nil || len(fields.Fields) == 0 {
		// If no form fields found, return original
		return data, nil
	}

	// Create field map for quick lookup
	fieldMap := make(map[string]*FormField)
	for i := range fields.Fields {
		fieldMap[fields.Fields[i].Name] = &fields.Fields[i]
	}

	// Save PDF to temp file for gopdf
	tmpDir := os.TempDir()
	tmpInput := filepath.Join(tmpDir, fmt.Sprintf("input_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(tmpInput)
	if err := os.WriteFile(tmpInput, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp PDF: %w", err)
	}

	// Get page count
	pageResult, err := ExtractPages(ExtractPagesInput{PDFPath: tmpInput})
	if err != nil {
		return data, nil
	}

	// Create new PDF with gopdf
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// Try to add Arial font (fallback to built-in if not available)
	fontAdded := false
	fontPaths := []string{
		"/usr/share/fonts/truetype/arial.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"./fonts/Arial.ttf",
	}
	for _, fontPath := range fontPaths {
		if err := pdf.AddTTFFont("Arial", fontPath); err == nil {
			fontAdded = true
			break
		}
	}
	if !fontAdded {
		// Use built-in Helvetica if TTF not available
		pdf.SetFont("helvetica", "", 10)
	} else {
		pdf.SetFont("Arial", "", 10)
	}

	// Import all pages and overlay field values
	for pageNum := 1; pageNum <= pageResult.PageCount; pageNum++ {
		pdf.AddPage()
		tpl := pdf.ImportPage(tmpInput, pageNum, "/MediaBox")
		pdf.UseImportedTemplate(tpl, 0, 0, 0, 0)

		// Overlay field values on this page
		for fieldName, value := range input.Fields {
			if field, exists := fieldMap[fieldName]; exists && field.Page == pageNum {
				// gopdf uses bottom-left origin, PDF form fields use top-left
				// Convert Y coordinate: pageHeight - fieldY - fieldHeight
				pageHeight := 792.0 // A4 height in points
				yPos := pageHeight - field.Y - field.Height

				pdf.SetXY(field.X, yPos)
				if fontAdded {
					pdf.SetFont("Arial", "", 10)
				}
				pdf.Cell(nil, value)
			}
		}
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return data, nil // Return original on error
	}

	return buf.Bytes(), nil
}

// MergeSignaturesInput input data for merging signatures
type MergeSignaturesInput struct {
	BasePDF    []byte
	Signatures []SignatureInfo
}

// SignatureInfo signature information
type SignatureInfo struct {
	Page       int     // page number (starting from 1)
	X          float64 // X position
	Y          float64 // Y position
	Width      float64
	Height     float64
	ImageData  []byte  // signature image (PNG/JPEG)
	SignerName string
	Date       time.Time
}

// MergeSignatures merges signatures into PDF using gopdf
func MergeSignatures(input MergeSignaturesInput) ([]byte, error) {
	// If no signatures, return original
	if len(input.Signatures) == 0 {
		return input.BasePDF, nil
	}

	// Save base PDF to temporary file
	tmpDir := os.TempDir()
	tmpInput := filepath.Join(tmpDir, fmt.Sprintf("base_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(tmpInput)

	if err := os.WriteFile(tmpInput, input.BasePDF, 0644); err != nil {
		return nil, fmt.Errorf("failed to write base PDF: %w", err)
	}

	// Get page count
	pageResult, err := ExtractPages(ExtractPagesInput{PDFPath: tmpInput})
	if err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}

	// Group signatures by page
	signaturesByPage := make(map[int][]SignatureInfo)
	for _, sig := range input.Signatures {
		if sig.Page > 0 && sig.Page <= pageResult.PageCount {
			signaturesByPage[sig.Page] = append(signaturesByPage[sig.Page], sig)
		}
	}

	// Create new PDF with gopdf
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// Import all pages and add signatures
	for pageNum := 1; pageNum <= pageResult.PageCount; pageNum++ {
		pdf.AddPage()
		tpl := pdf.ImportPage(tmpInput, pageNum, "/MediaBox")
		pdf.UseImportedTemplate(tpl, 0, 0, 0, 0)

		// Add signatures for this page
		if sigs, exists := signaturesByPage[pageNum]; exists {
			for i, sig := range sigs {
				// Save signature image to temporary file
				tmpImage := filepath.Join(tmpDir, fmt.Sprintf("sig_%d_%d.png", time.Now().UnixNano(), i))
				if err := os.WriteFile(tmpImage, sig.ImageData, 0644); err != nil {
					continue
				}
				defer os.Remove(tmpImage)

				// gopdf uses bottom-left origin, convert Y coordinate
				pageHeight := 792.0 // A4 height in points
				yPos := pageHeight - sig.Y - sig.Height

				// Place image at specified position
				rect := &gopdf.Rect{}
				if sig.Width > 0 && sig.Height > 0 {
					rect.W = sig.Width
					rect.H = sig.Height
				}
				pdf.Image(tmpImage, sig.X, yPos, rect)
			}
		}
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateAuditTrailInput input data for generating audit trail
type GenerateAuditTrailInput struct {
	Submission *models.Submission
	Submitters []*models.Submitter
	Events     []*models.Event
}

// GenerateAuditTrail generates audit trail page using gopdf
func GenerateAuditTrail(input GenerateAuditTrailInput) ([]byte, error) {
	// Create new PDF with gopdf
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// Try to add fonts (fallback to built-in if not available)
	fontAdded := false
	boldFontAdded := false
	fontPaths := []string{
		"/usr/share/fonts/truetype/arial.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"./fonts/Arial.ttf",
	}
	for _, fontPath := range fontPaths {
		if !fontAdded {
			if err := pdf.AddTTFFont("Arial", fontPath); err == nil {
				fontAdded = true
			}
		}
		if !boldFontAdded {
			boldPath := fontPath
			if fontPath == "/usr/share/fonts/truetype/arial.ttf" {
				boldPath = "/usr/share/fonts/truetype/arialbd.ttf"
			} else if fontPath == "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf" {
				boldPath = "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"
			} else if fontPath == "./fonts/Arial.ttf" {
				boldPath = "./fonts/Arial-Bold.ttf"
			}
			if err := pdf.AddTTFFont("Arial-Bold", boldPath); err == nil {
				boldFontAdded = true
			}
		}
	}

	pdf.AddPage()
	yPos := 50.0
	lineHeight := 12.0

	// Header
	if boldFontAdded {
		pdf.SetFont("Arial-Bold", "", 16)
	} else {
		pdf.SetFont("helvetica", "B", 16)
	}
	pdf.SetXY(50, yPos)
	pdf.Cell(nil, "Audit Trail")
	yPos += lineHeight * 2

	// Submission information
	if fontAdded {
		pdf.SetFont("Arial", "", 10)
	} else {
		pdf.SetFont("helvetica", "", 10)
	}
	pdf.SetXY(50, yPos)
	pdf.Cell(nil, fmt.Sprintf("Submission ID: %s", input.Submission.ID))
	yPos += lineHeight

	pdf.SetXY(50, yPos)
	pdf.Cell(nil, fmt.Sprintf("Created: %s", input.Submission.CreatedAt.Format("2006-01-02 15:04:05")))
	yPos += lineHeight * 2

	// Signers
	if boldFontAdded {
		pdf.SetFont("Arial-Bold", "", 10)
	} else {
		pdf.SetFont("helvetica", "B", 10)
	}
	pdf.SetXY(50, yPos)
	pdf.Cell(nil, "Signers:")
	yPos += lineHeight

	if fontAdded {
		pdf.SetFont("Arial", "", 10)
	} else {
		pdf.SetFont("helvetica", "", 10)
	}
	for _, submitter := range input.Submitters {
		signerText := fmt.Sprintf("- %s (%s)", submitter.Name, submitter.Email)
		pdf.SetXY(60, yPos)
		pdf.Cell(nil, signerText)
		yPos += lineHeight

		if submitter.CompletedAt != nil {
			signedText := fmt.Sprintf("  Signed at: %s", submitter.CompletedAt.Format("2006-01-02 15:04:05"))
			if fontAdded {
				pdf.SetFont("Arial", "", 9)
			} else {
				pdf.SetFont("helvetica", "", 9)
			}
			pdf.SetXY(60, yPos)
			pdf.Cell(nil, signedText)
			yPos += lineHeight
			if fontAdded {
				pdf.SetFont("Arial", "", 10)
			} else {
				pdf.SetFont("helvetica", "", 10)
			}
		}
	}

	// Events timeline
	if len(input.Events) > 0 {
		yPos += lineHeight
		if boldFontAdded {
			pdf.SetFont("Arial-Bold", "", 10)
		} else {
			pdf.SetFont("helvetica", "B", 10)
		}
		pdf.SetXY(50, yPos)
		pdf.Cell(nil, "Timeline:")
		yPos += lineHeight

		if fontAdded {
			pdf.SetFont("Arial", "", 9)
		} else {
			pdf.SetFont("helvetica", "", 9)
		}
		for _, event := range input.Events {
			eventText := fmt.Sprintf("%s - %s", event.CreatedAt.Format("2006-01-02 15:04:05"), event.Type)
			pdf.SetXY(60, yPos)
			pdf.Cell(nil, eventText)
			yPos += lineHeight
		}
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// AppendAuditTrail appends audit trail page to PDF using gopdf
func AppendAuditTrail(basePDF []byte, auditTrailPDF []byte) ([]byte, error) {
	return AppendPDF(basePDF, auditTrailPDF)
}

// AssembleDocument assembles final document with all signatures and audit trail
func AssembleDocument(basePDF []byte, signatures []SignatureInfo, auditInput GenerateAuditTrailInput) ([]byte, error) {
	// 1. Merge signatures
	pdfWithSignatures, err := MergeSignatures(MergeSignaturesInput{
		BasePDF:    basePDF,
		Signatures: signatures,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to merge signatures: %w", err)
	}

	// 2. Generate audit trail
	auditTrail, err := GenerateAuditTrail(auditInput)
	if err != nil {
		return nil, fmt.Errorf("failed to generate audit trail: %w", err)
	}

	// 3. Append audit trail to document
	finalPDF, err := AppendAuditTrail(pdfWithSignatures, auditTrail)
	if err != nil {
		return nil, fmt.Errorf("failed to append audit trail: %w", err)
	}

	return finalPDF, nil
}

