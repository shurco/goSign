package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFormFields_invalidPDFReturnsEmpty(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.pdf")
	if err := os.WriteFile(p, []byte("not a pdf"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := ExtractFormFields(ExtractFormFieldsInput{PDFPath: p})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res == nil || len(res.Fields) != 0 {
		t.Fatalf("expected empty fields, got %+v", res)
	}
}
