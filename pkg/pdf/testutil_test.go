package pdf

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/signintech/gopdf"
)

var errNoTestTTF = errors.New("no TTF font available for gopdf tests")

// buildTestPDFData builds a valid multi-page PDF; returns errNoTestTTF if no system font was found.
func buildTestPDFData(pages int, cellText string) ([]byte, error) {
	if pages < 1 {
		pages = 1
	}
	if cellText == "" {
		cellText = "test"
	}
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	fonts := addStandardFonts(&pdf, "")
	if !fonts.NormalOK {
		return nil, errNoTestTTF
	}
	for range pages {
		pdf.AddPage()
		if err := pdf.SetFont(fonts.NormalName, "", 10); err != nil {
			return nil, err
		}
		pdf.SetXY(50, 50)
		pdf.Cell(nil, cellText)
	}
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// buildTestPDF is like buildTestPDFData but skips or fatals under testing.T.
func buildTestPDF(t *testing.T, pages int, cellText string) []byte {
	t.Helper()
	b, err := buildTestPDFData(pages, cellText)
	if err != nil {
		if errors.Is(err, errNoTestTTF) {
			t.Skip("no TTF fonts available for gopdf")
		}
		t.Fatalf("build test PDF: %v", err)
	}
	return b
}

// writeTempPDFPath writes buildTestPDF output to a temp file and registers cleanup.
func writeTempPDFPath(t *testing.T, pages int) string {
	t.Helper()
	data := buildTestPDF(t, pages, "")
	f, err := os.CreateTemp("", "gosign_pdf_test_*.pdf")
	if err != nil {
		t.Fatalf("create temp pdf: %v", err)
	}
	t.Cleanup(func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	})
	if _, err := f.Write(data); err != nil {
		t.Fatalf("write temp pdf: %v", err)
	}
	return f.Name()
}

func pageCountFromBytes(t *testing.T, b []byte) int {
	t.Helper()
	f, err := os.CreateTemp("", "gosign_pagecount_*.pdf")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(b); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := f.Sync(); err != nil {
		t.Fatalf("sync: %v", err)
	}
	res, err := ExtractPages(ExtractPagesInput{PDFPath: f.Name()})
	if err != nil {
		t.Fatalf("ExtractPages: %v", err)
	}
	return res.PageCount
}

func assertPDFHeader(t *testing.T, b []byte) {
	t.Helper()
	if len(b) == 0 || !bytes.HasPrefix(b, []byte("%PDF")) {
		t.Errorf("expected non-empty PDF with %%PDF prefix, len=%d", len(b))
	}
}

func assertErrorWants(t *testing.T, err error, wantErr bool) bool {
	t.Helper()
	if got := err != nil; got != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
		return false
	}
	return true
}
