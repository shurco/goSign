package pdf

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/signintech/gopdf"

	"github.com/shurco/gosign/internal/models"
)

func TestRenderCompletedTemplatePDF_smoke(t *testing.T) {
	tmp := t.TempDir()
	pagesDir := filepath.Join(tmp, "lc_pages")
	attID := "att-1"
	pageDir := filepath.Join(pagesDir, attID)
	if err := os.MkdirAll(pageDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a minimal one-page PDF to act as stored page (lc_pages/{att}/0.pdf).
	pagePath := filepath.Join(pageDir, "0.pdf")
	{
		p := gopdf.GoPdf{}
		p.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
		p.AddPage()
		// Create a valid PDF that gofpdi can re-import.
		if err := p.AddTTFFont("DejaVuSans", "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"); err != nil {
			t.Fatalf("failed to add font: %v", err)
		}
		if err := p.SetFont("DejaVuSans", "", 12); err != nil {
			t.Fatalf("failed to set font: %v", err)
		}
		p.SetXY(50, 50)
		p.Cell(nil, "base")
		var buf bytes.Buffer
		if err := p.Write(&buf); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(pagePath, buf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Small PNG for signature field.
	var sigDataURL string
	{
		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		img.Set(0, 0, color.RGBA{R: 255, A: 255})
		var b bytes.Buffer
		if err := png.Encode(&b, img); err != nil {
			t.Fatal(err)
		}
		sigDataURL = "data:image/png;base64," + base64.StdEncoding.EncodeToString(b.Bytes())
	}

	fieldTextID := "field-text"
	fieldSigID := "field-sig"

	out, err := RenderCompletedTemplatePDF(RenderCompletedTemplatePDFInput{
		PagesDir: pagesDir,
		Schema: []models.Schema{
			{AttachmentID: attID, Name: "page_1"},
		},
		Fields: []models.Field{
			{
				ID:   fieldTextID,
				Type: models.FieldTypeText,
				Areas: []*models.Areas{
					{AttachmentID: attID, X: 0.1, Y: 0.1, W: 0.3, H: 0.05},
				},
			},
			{
				ID:   fieldSigID,
				Type: models.FieldTypeSignature,
				Areas: []*models.Areas{
					{AttachmentID: attID, X: 0.1, Y: 0.2, W: 0.3, H: 0.1},
				},
			},
		},
		Values: map[string]any{
			fieldTextID: "Hello",
			fieldSigID:  sigDataURL,
		},
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("expected non-empty PDF output")
	}
	// Quick sanity check: PDF header.
	if !bytes.HasPrefix(out, []byte("%PDF")) {
		t.Fatalf("expected PDF header, got %q", string(out[:min(16, len(out))]))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

