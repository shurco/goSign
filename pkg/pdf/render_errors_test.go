package pdf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shurco/gosign/internal/models"
)

func TestRenderCompletedTemplatePDF_validationErrors(t *testing.T) {
	_, err := RenderCompletedTemplatePDF(RenderCompletedTemplatePDFInput{})
	if err == nil {
		t.Fatal("expected error for empty PagesDir")
	}

	_, err = RenderCompletedTemplatePDF(RenderCompletedTemplatePDFInput{
		PagesDir: t.TempDir(),
		Schema:   nil,
	})
	if err == nil {
		t.Fatal("expected error for empty schema")
	}
}

func TestRenderCompletedTemplatePDF_skipsEmptyAttachmentID(t *testing.T) {
	tmp := t.TempDir()
	att := "a1"
	pageDir := filepath.Join(tmp, att)
	if err := os.MkdirAll(pageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pagePath := filepath.Join(pageDir, "0.pdf")
	if err := os.WriteFile(pagePath, buildTestPDF(t, 1, "p"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := RenderCompletedTemplatePDF(RenderCompletedTemplatePDFInput{
		PagesDir: tmp,
		Schema: []models.Schema{
			{AttachmentID: ""},
			{AttachmentID: att, Name: "p1"},
		},
		Fields: nil,
		Values: nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("expected PDF bytes")
	}
}

func TestRenderCompletedTemplatePDF_missingPageFile(t *testing.T) {
	tmp := t.TempDir()
	_, err := RenderCompletedTemplatePDF(RenderCompletedTemplatePDFInput{
		PagesDir: tmp,
		Schema:   []models.Schema{{AttachmentID: "missing", Name: "x"}},
	})
	if err == nil {
		t.Fatal("expected error for missing page PDF")
	}
}
