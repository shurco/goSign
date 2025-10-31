package pdf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/signintech/gopdf"

	"github.com/shurco/gosign/internal/models"
)

// FillFieldsInput input data for filling fields
type FillFieldsInput struct {
	PDFPath string
	Fields  map[string]string // field_name -> value
}

// FillFields fills PDF fields with values using pdfcpu
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

	// Create temporary files
	tmpDir := os.TempDir()
	tmpJSON := filepath.Join(tmpDir, fmt.Sprintf("fields_%d.json", time.Now().UnixNano()))
	tmpOutput := filepath.Join(tmpDir, fmt.Sprintf("filled_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(tmpJSON)
	defer os.Remove(tmpOutput)

	// Convert fields map to JSON
	jsonData, err := json.Marshal(input.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields: %w", err)
	}
	
	if err := os.WriteFile(tmpJSON, jsonData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write JSON: %w", err)
	}

	// Fill form fields using pdfcpu
	conf := model.NewDefaultConfiguration()
	err = api.FillFormFile(input.PDFPath, tmpJSON, tmpOutput, conf)
	if err != nil {
		// If form filling fails (no AcroForm fields), return original
		return data, nil
	}

	// Read filled PDF
	filledData, err := os.ReadFile(tmpOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to read filled PDF: %w", err)
	}

	return filledData, nil
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

// MergeSignatures merges signatures into PDF using pdfcpu
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

	currentPDF := tmpInput

	// Process each signature
	for i, sig := range input.Signatures {
		// Save signature image to temporary file
		tmpImage := filepath.Join(tmpDir, fmt.Sprintf("sig_%d_%d.png", time.Now().UnixNano(), i))
		if err := os.WriteFile(tmpImage, sig.ImageData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write signature image: %w", err)
		}
		defer os.Remove(tmpImage)

		// Create output for this step
		tmpOutput := filepath.Join(tmpDir, fmt.Sprintf("merged_%d_%d.pdf", time.Now().UnixNano(), i))
		defer os.Remove(tmpOutput)

		// Add image to PDF at specified position
		conf := model.NewDefaultConfiguration()
		
		// Create watermark for image placement
		wm, err := api.ImageWatermark(tmpImage, "", false, false, types.POINTS)
		if err != nil {
			return nil, fmt.Errorf("failed to create watermark: %w", err)
		}

		// Set position and size (convert float64 to points)
		wm.Dx = sig.X
		wm.Dy = sig.Y
		wm.Scale = 1.0
		wm.ScaleAbs = true
		
		// Calculate scale to fit signature in specified width/height
		if sig.Width > 0 && sig.Height > 0 {
			wm.Width = int(sig.Width)
			wm.Height = int(sig.Height)
		}

		// Apply to specific pages
		pages := []string{fmt.Sprintf("%d", sig.Page)}

		// Add watermark (image) to PDF
		err = api.AddWatermarksFile(currentPDF, tmpOutput, pages, wm, conf)
		if err != nil {
			return nil, fmt.Errorf("failed to add signature to PDF: %w", err)
		}

		currentPDF = tmpOutput
	}

	// Read final PDF
	finalData, err := os.ReadFile(currentPDF)
	if err != nil {
		return nil, fmt.Errorf("failed to read final PDF: %w", err)
	}

	return finalData, nil
}

// GenerateAuditTrailInput input data for generating audit trail
type GenerateAuditTrailInput struct {
	Submission *models.Submission
	Submitters []*models.Submitter
	Events     []*models.Event
}

// GenerateAuditTrail generates audit trail page
func GenerateAuditTrail(input GenerateAuditTrailInput) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	// TODO: Add font (font file required)
	// pdf.AddTTFFont("Arial", "./fonts/arial.ttf")
	// pdf.SetFont("Arial", "", 16)

	// Simple version without fonts
	// Full implementation should:
	// 1. Add TTF fonts
	// 2. Draw header
	// 3. Add submission information
	// 4. Add list of signers with dates
	// 5. Add events timeline

	// Document information (without fonts - structure only)
	text := fmt.Sprintf("Audit Trail\n\nSubmission ID: %s\nCreated: %s\n\n",
		input.Submission.ID,
		input.Submission.CreatedAt.Format("2006-01-02 15:04:05"))

	// Signers
	text += "Signers:\n"
	for _, submitter := range input.Submitters {
		text += fmt.Sprintf("- %s (%s)\n", submitter.Name, submitter.Email)
		if submitter.CompletedAt != nil {
			text += fmt.Sprintf("  Signed at: %s\n", submitter.CompletedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// Events
	if len(input.Events) > 0 {
		text += "\nTimeline:\n"
		for _, event := range input.Events {
			text += fmt.Sprintf("%s - %s\n", event.CreatedAt.Format("2006-01-02 15:04:05"), event.Type)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate audit trail: %w", err)
	}

	return buf.Bytes(), nil
}

// AppendAuditTrail appends audit trail page to PDF using pdfcpu
func AppendAuditTrail(basePDF []byte, auditTrailPDF []byte) ([]byte, error) {
	// Create temporary directory
	tmpDir := os.TempDir()
	
	// Save base PDF
	tmpBase := filepath.Join(tmpDir, fmt.Sprintf("base_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(tmpBase)
	if err := os.WriteFile(tmpBase, basePDF, 0644); err != nil {
		return nil, fmt.Errorf("failed to write base PDF: %w", err)
	}

	// Save audit trail PDF
	tmpAudit := filepath.Join(tmpDir, fmt.Sprintf("audit_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(tmpAudit)
	if err := os.WriteFile(tmpAudit, auditTrailPDF, 0644); err != nil {
		return nil, fmt.Errorf("failed to write audit trail PDF: %w", err)
	}

	// Output file
	tmpOutput := filepath.Join(tmpDir, fmt.Sprintf("merged_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(tmpOutput)

	// Merge PDFs using pdfcpu
	conf := model.NewDefaultConfiguration()
	err := api.MergeCreateFile([]string{tmpBase, tmpAudit}, tmpOutput, false, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to merge PDFs: %w", err)
	}

	// Read merged PDF
	mergedData, err := os.ReadFile(tmpOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to read merged PDF: %w", err)
	}

	return mergedData, nil
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

