package pdf

import (
	"bytes"
	"os"
	"testing"

	"github.com/signintech/gopdf"
	"github.com/shurco/gosign/internal/assets"
)

func makeOnePagePDF(t *testing.T) []byte {
	t.Helper()
	p := gopdf.GoPdf{}
	p.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	p.AddPage()
	fonts := addStandardFonts(&p, "")
	if !fonts.NormalOK {
		t.Skip("no TTF fonts available for gopdf")
	}
	if err := p.SetFont(fonts.NormalName, "", 10); err != nil {
		t.Fatalf("failed to set font: %v", err)
	}
	p.SetXY(50, 50)
	p.Cell(nil, "base")
	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("failed to write base PDF: %v", err)
	}
	return buf.Bytes()
}

func pageCountFromBytes(t *testing.T, b []byte) int {
	t.Helper()
	f, err := os.CreateTemp("", "gosign_test_pages_*.pdf")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(b); err != nil {
		t.Fatalf("failed to write temp pdf: %v", err)
	}
	res, err := ExtractPages(ExtractPagesInput{PDFPath: f.Name()})
	if err != nil {
		t.Fatalf("failed to extract pages: %v", err)
	}
	return res.PageCount
}

func TestGenerateSignatureCertificatePDF(t *testing.T) {
	assetPaths, err := assets.EnsureOnDisk(t.TempDir())
	if err != nil {
		t.Fatalf("failed to prepare assets: %v", err)
	}
	cert, err := GenerateSignatureCertificatePDF(SignatureCertificateInput{
		DocumentName: "Test Document",
		Reference:    "submission_123",
		AssetsDir:    assetPaths.Dir,
		QRURL:        "https://example.com/public/sign/slug/certificate",
		Signers: []SignatureCertificateSigner{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		},
	})
	if err != nil {
		t.Fatalf("GenerateSignatureCertificatePDF() error: %v", err)
	}
	if len(cert) == 0 {
		t.Fatalf("expected non-empty PDF bytes")
	}
	if got := pageCountFromBytes(t, cert); got < 1 {
		t.Fatalf("expected at least 1 page, got %d", got)
	}
}

func TestAppendSignatureCertificate(t *testing.T) {
	assetPaths, err := assets.EnsureOnDisk(t.TempDir())
	if err != nil {
		t.Fatalf("failed to prepare assets: %v", err)
	}
	base := makeOnePagePDF(t)
	cert, err := GenerateSignatureCertificatePDF(SignatureCertificateInput{
		DocumentName: "Test Document",
		Reference:    "submission_123",
		AssetsDir:    assetPaths.Dir,
		QRURL:        "https://example.com/public/sign/slug/certificate",
		Signers: []SignatureCertificateSigner{
			{Name: "Alice", Email: "alice@example.com"},
		},
	})
	if err != nil {
		t.Fatalf("GenerateSignatureCertificatePDF() error: %v", err)
	}

	merged, err := AppendSignatureCertificate(base, cert)
	if err != nil {
		t.Fatalf("AppendSignatureCertificate() error: %v", err)
	}
	if len(merged) <= len(base) {
		t.Fatalf("expected merged PDF to be larger than base PDF")
	}
	gotPages := pageCountFromBytes(t, merged)
	if gotPages < 2 {
		t.Fatalf("expected merged PDF to have at least 2 pages, got %d", gotPages)
	}
}

