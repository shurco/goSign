package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/signintech/gopdf"
)

// AppendPDF appends all pages from extraPDF to the end of basePDF.
//
// Implementation detail: gopdf's import APIs are file-path based, so we write both
// PDFs to temporary files and then re-compose a new PDF by importing pages.
func AppendPDF(basePDF []byte, extraPDF []byte) ([]byte, error) {
	if len(basePDF) == 0 {
		return nil, fmt.Errorf("base PDF is empty")
	}
	if len(extraPDF) == 0 {
		return nil, fmt.Errorf("extra PDF is empty")
	}

	tmpDir := os.TempDir()

	// Save base PDF.
	tmpBase := filepath.Join(tmpDir, fmt.Sprintf("gosign_base_%d.pdf", time.Now().UnixNano()))
	if err := os.WriteFile(tmpBase, basePDF, 0644); err != nil {
		return nil, fmt.Errorf("failed to write base PDF: %w", err)
	}
	defer func() { _ = os.Remove(tmpBase) }()

	// Save extra PDF.
	tmpExtra := filepath.Join(tmpDir, fmt.Sprintf("gosign_extra_%d.pdf", time.Now().UnixNano()))
	if err := os.WriteFile(tmpExtra, extraPDF, 0644); err != nil {
		return nil, fmt.Errorf("failed to write extra PDF: %w", err)
	}
	defer func() { _ = os.Remove(tmpExtra) }()

	// Count pages.
	basePages, err := ExtractPages(ExtractPagesInput{PDFPath: tmpBase})
	if err != nil {
		return nil, fmt.Errorf("failed to get base PDF page count: %w", err)
	}
	extraPages, err := ExtractPages(ExtractPagesInput{PDFPath: tmpExtra})
	if err != nil {
		return nil, fmt.Errorf("failed to get extra PDF page count: %w", err)
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// Import all base pages.
	for pageNum := 1; pageNum <= basePages.PageCount; pageNum++ {
		pdf.AddPage()
		tpl := pdf.ImportPage(tmpBase, pageNum, "/MediaBox")
		pdf.UseImportedTemplate(tpl, 0, 0, 0, 0)
	}

	// Import all extra pages.
	for pageNum := 1; pageNum <= extraPages.PageCount; pageNum++ {
		pdf.AddPage()
		tpl := pdf.ImportPage(tmpExtra, pageNum, "/MediaBox")
		pdf.UseImportedTemplate(tpl, 0, 0, 0, 0)
	}

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write merged PDF: %w", err)
	}
	return buf.Bytes(), nil
}

