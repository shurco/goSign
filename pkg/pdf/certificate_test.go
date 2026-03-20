package pdf

import (
	"testing"

	"github.com/shurco/gosign/internal/assets"
)

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
	base := buildTestPDF(t, 1, "base")
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
