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
		// Create a valid PDF that gofpdi can re-import (TTF from common system paths).
		fs := addStandardFonts(&p, "")
		if !fs.NormalOK {
			t.Skip("no suitable TTF font found for smoke test (install Arial or DejaVu Sans)")
		}
		if err := p.SetFont(fs.NormalName, "", 12); err != nil {
			t.Fatalf("failed to set font: %v", err)
		}
		p.SetXY(50, 50)
		p.Cell(nil, "base")
		var buf bytes.Buffer
		if _, err := p.WriteTo(&buf); err != nil {
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
	fieldTextZeroHID := "field-text-zero-h"
	fieldSigID := "field-sig"
	withID := true

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
				ID:   fieldTextZeroHID,
				Type: models.FieldTypeText,
				Areas: []*models.Areas{
					{AttachmentID: attID, X: 0.5, Y: 0.1, W: 0.2, H: 0}, // triggers default height
				},
			},
			{
				ID:   fieldSigID,
				Type: models.FieldTypeSignature,
				Preferences: &models.FieldPreferences{
					WithSignatureID: withID,
				},
				Areas: []*models.Areas{
					{AttachmentID: attID, X: 0.1, Y: 0.2, W: 0.3, H: 0.1},
				},
			},
		},
		Values: map[string]any{
			fieldTextID:                  "Hello",
			fieldTextZeroHID:             "Tall",
			fieldSigID:                   sigDataURL,
			fieldSigID + "_signature_id": "SIG-42",
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
		prefixLen := min(16, len(out))
		t.Fatalf("expected PDF header, got %q", string(out[:prefixLen]))
	}
}

func TestFormatDateValue(t *testing.T) {
	tests := []struct {
		value, pattern, want string
	}{
		{"2026-07-28", "DD/MM/YYYY", "28/07/2026"},
		{"2026-07-28", "MM/DD/YYYY", "07/28/2026"},
		{"2026-07-28", "YYYY-MM-DD", "2026-07-28"},
		{"2026-07-28", "MMM D, YYYY", "Jul 28, 2026"},
		{"2026-07-28", "MMMM D, YYYY", "July 28, 2026"},
		{"2026-07-28", "D MMM YYYY", "28 Jul 2026"},
		{"2026-07-28T10:00:00Z", "DD.MM.YYYY", "28.07.2026"},
		{"2026-07-28", "", "28/07/2026"}, // default pattern
		{"not-a-date", "DD/MM/YYYY", "not-a-date"},
		{"", "DD/MM/YYYY", ""},
	}
	for _, tt := range tests {
		if got := formatDateValue(tt.value, tt.pattern); got != tt.want {
			t.Errorf("formatDateValue(%q, %q) = %q, want %q", tt.value, tt.pattern, got, tt.want)
		}
	}
}

func TestFormatNumberValue(t *testing.T) {
	tests := []struct {
		value, format, want string
	}{
		{"1000.5", "comma", "1,000.5"},
		{"1000.5", "dot", "1.000,5"},
		{"1000.5", "space", "1 000,5"},
		{"1000.5", "usd", "$1,000.50"},
		{"1000.5", "eur", "€1.000,50"},
		{"1000.5", "gbp", "£1,000.50"},
		{"12.5", "percent", "12.5%"},
		{"1234567", "comma", "1,234,567"},
		{"-1234.56", "comma", "-1,234.56"},
		{"1000", "", "1000"},
		{"1000", "none", "1000"},
		{"abc", "comma", "abc"},
	}
	for _, tt := range tests {
		if got := formatNumberValue(tt.value, tt.format); got != tt.want {
			t.Errorf("formatNumberValue(%q, %q) = %q, want %q", tt.value, tt.format, got, tt.want)
		}
	}
}

func TestValueContains(t *testing.T) {
	if !valueContains("Red", "Red") {
		t.Error("string match failed")
	}
	if valueContains("Red", "Blue") {
		t.Error("string mismatch should be false")
	}
	if !valueContains([]any{"Red", "Blue"}, "Blue") {
		t.Error("[]any membership failed")
	}
	if valueContains([]any{"Red"}, "Green") {
		t.Error("[]any non-member should be false")
	}
	if !valueContains([]string{"A", "B"}, "A") {
		t.Error("[]string membership failed")
	}
	if valueContains(42, "42") {
		t.Error("non-string type should be false")
	}
}

func TestOptionValueByID(t *testing.T) {
	field := models.Field{
		Options: models.FieldOptions{
			{ID: "opt-1", Value: "Yes"},
			{ID: "opt-2", Value: ""},
		},
	}
	if got := optionValueByID(field, "opt-1"); got != "Yes" {
		t.Errorf("expected Yes, got %q", got)
	}
	if got := optionValueByID(field, "opt-2"); got != "Option 2" {
		t.Errorf("expected fallback Option 2, got %q", got)
	}
	if got := optionValueByID(field, "missing"); got != "" {
		t.Errorf("expected empty for unknown id, got %q", got)
	}
}

func TestIsTruthyValue(t *testing.T) {
	for _, v := range []any{true, "true", "1", "yes"} {
		if !isTruthyValue(v) {
			t.Errorf("expected truthy for %v", v)
		}
	}
	for _, v := range []any{false, "false", "", nil, 1} {
		if isTruthyValue(v) {
			t.Errorf("expected falsy for %v", v)
		}
	}
}

func TestParseHexColor(t *testing.T) {
	r, g, b, ok := parseHexColor("#FF8000")
	if !ok || r != 255 || g != 128 || b != 0 {
		t.Errorf("parseHexColor(#FF8000) = %d,%d,%d,%v", r, g, b, ok)
	}
	if _, _, _, ok := parseHexColor(""); ok {
		t.Error("empty string should not parse")
	}
	if _, _, _, ok := parseHexColor("#FFF"); ok {
		t.Error("short hex should not parse")
	}
}
