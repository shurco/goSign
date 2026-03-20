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

	tmpInput, removeTmp, err := tempPDFFile(data, "input")
	if err != nil {
		return nil, fmt.Errorf("failed to write temp PDF: %w", err)
	}
	defer removeTmp()

	// Get page count
	pageResult, err := ExtractPages(ExtractPagesInput{PDFPath: tmpInput})
	if err != nil {
		return data, nil
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	fonts := addStandardFonts(&pdf, "")
	fonts.SetNormal(&pdf, 10)

	// Import all pages and overlay field values
	for pageNum := 1; pageNum <= pageResult.PageCount; pageNum++ {
		pdf.AddPage()
		tpl := pdf.ImportPage(tmpInput, pageNum, "/MediaBox")
		pdf.UseImportedTemplate(tpl, 0, 0, 0, 0)

		// Overlay field values on this page
		for fieldName, value := range input.Fields {
			if field, exists := fieldMap[fieldName]; exists && field.Page == pageNum {
				// gopdf uses bottom-left origin, PDF form fields use top-left.
				yPos := A4HeightPt - field.Y - field.Height
				pdf.SetXY(field.X, yPos)
				fonts.SetNormal(&pdf, 10)
				pdf.Cell(nil, value)
			}
		}
	}

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write filled PDF: %w", err)
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
	ImageData  []byte // signature image (PNG/JPEG)
	SignerName string
	Date       time.Time
}

// MergeSignatures merges signatures into PDF using gopdf
func MergeSignatures(input MergeSignaturesInput) ([]byte, error) {
	// If no signatures, return original
	if len(input.Signatures) == 0 {
		return input.BasePDF, nil
	}

	tmpInput, removeBase, err := tempPDFFile(input.BasePDF, "base")
	if err != nil {
		return nil, fmt.Errorf("failed to write base PDF: %w", err)
	}
	defer removeBase()
	tmpDir := filepath.Dir(tmpInput)

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
				tmpImage := filepath.Join(tmpDir, fmt.Sprintf("sig_%d_%d.png", time.Now().UnixNano(), i))
				if err := os.WriteFile(tmpImage, sig.ImageData, 0644); err != nil {
					continue
				}
				yPos := A4HeightPt - sig.Y - sig.Height
				rect := &gopdf.Rect{}
				if sig.Width > 0 && sig.Height > 0 {
					rect.W = sig.Width
					rect.H = sig.Height
				}
				pdf.Image(tmpImage, sig.X, yPos, rect)
				_ = os.Remove(tmpImage)
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
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	fonts := addStandardFonts(&pdf, "")

	pdf.AddPage()
	yPos := 50.0
	const lineHeight = 12.0

	fonts.SetBold(&pdf, 16)
	pdf.SetXY(50, yPos)
	pdf.Cell(nil, "Audit Trail")
	yPos += lineHeight * 2

	fonts.SetNormal(&pdf, 10)
	pdf.SetXY(50, yPos)
	pdf.Cell(nil, fmt.Sprintf("Submission ID: %s", input.Submission.ID))
	yPos += lineHeight

	pdf.SetXY(50, yPos)
	pdf.Cell(nil, fmt.Sprintf("Created: %s", input.Submission.CreatedAt.Format("2006-01-02 15:04:05")))
	yPos += lineHeight * 2

	fonts.SetBold(&pdf, 10)
	pdf.SetXY(50, yPos)
	pdf.Cell(nil, "Signers:")
	yPos += lineHeight

	fonts.SetNormal(&pdf, 10)
	for _, submitter := range input.Submitters {
		pdf.SetXY(60, yPos)
		pdf.Cell(nil, fmt.Sprintf("- %s (%s)", submitter.Name, submitter.Email))
		yPos += lineHeight

		if submitter.CompletedAt != nil {
			fonts.SetNormal(&pdf, 9)
			pdf.SetXY(60, yPos)
			pdf.Cell(nil, fmt.Sprintf("  Signed at: %s", submitter.CompletedAt.Format("2006-01-02 15:04:05")))
			yPos += lineHeight
			fonts.SetNormal(&pdf, 10)
		}
	}

	if len(input.Events) > 0 {
		yPos += lineHeight
		fonts.SetBold(&pdf, 10)
		pdf.SetXY(50, yPos)
		pdf.Cell(nil, "Timeline:")
		yPos += lineHeight

		fonts.SetNormal(&pdf, 9)
		for _, event := range input.Events {
			pdf.SetXY(60, yPos)
			pdf.Cell(nil, fmt.Sprintf("%s - %s", event.CreatedAt.Format("2006-01-02 15:04:05"), event.Type))
			yPos += lineHeight
		}
	}

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

// tempPDFFile writes data to a unique file under os.TempDir(); cleanup removes the file (no-op on write error).
func tempPDFFile(data []byte, prefix string) (path string, cleanup func(), err error) {
	cleanup = func() {}
	p := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d.pdf", prefix, time.Now().UnixNano()))
	if err = os.WriteFile(p, data, 0o644); err != nil {
		return "", cleanup, err
	}
	return p, func() { _ = os.Remove(p) }, nil
}
