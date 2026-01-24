package pdf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/signintech/gopdf"
)

func makePDFBytes(t *testing.T, pages int) []byte {
	t.Helper()
	if pages <= 0 {
		pages = 1
	}
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	fonts := addStandardFonts(&pdf, "")
	if !fonts.NormalOK {
		t.Skip("no TTF fonts available for gopdf")
	}
	for i := 0; i < pages; i++ {
		pdf.AddPage()
		if err := pdf.SetFont(fonts.NormalName, "", 10); err != nil {
			t.Fatalf("failed to set font: %v", err)
		}
		pdf.SetXY(50, 50)
		pdf.Cell(nil, "test")
	}
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("failed to build test PDF: %v", err)
	}
	return buf.Bytes()
}

func TestFillFields(t *testing.T) {
	tests := []struct {
		name      string
		input     FillFieldsInput
		wantErr   bool
		checkFunc func([]byte) bool
	}{
		{
			name: "empty fields returns original PDF",
			input: FillFieldsInput{
				PDFPath: "testdata/sample.pdf",
				Fields:  map[string]string{},
			},
			wantErr: false,
			checkFunc: func(result []byte) bool {
				return len(result) > 0
			},
		},
		{
			name: "fill single field",
			input: FillFieldsInput{
				PDFPath: "testdata/sample.pdf",
				Fields: map[string]string{
					"name": "John Doe",
				},
			},
			wantErr: false,
			checkFunc: func(result []byte) bool {
				return len(result) > 0
			},
		},
		{
			name: "non-existent PDF returns error",
			input: FillFieldsInput{
				PDFPath: "testdata/nonexistent.pdf",
				Fields: map[string]string{
					"name": "John Doe",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if test file doesn't exist and test doesn't expect error
			if _, err := os.Stat(tt.input.PDFPath); os.IsNotExist(err) && !tt.wantErr {
				t.Skip("Test PDF file not found")
			}

			result, err := FillFields(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("FillFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				if !tt.checkFunc(result) {
					t.Errorf("FillFields() result validation failed")
				}
			}
		})
	}
}

func TestMergeSignatures(t *testing.T) {
	// Create sample PNG signature for testing
	sampleSignature := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG header
	}

	tests := []struct {
		name      string
		input     MergeSignaturesInput
		wantErr   bool
		checkFunc func([]byte) bool
	}{
		{
			name: "empty signatures returns original PDF",
			input: MergeSignaturesInput{
				BasePDF:    makePDFBytes(t, 1),
				Signatures: []SignatureInfo{},
			},
			wantErr: false,
			checkFunc: func(result []byte) bool {
				return len(result) > 0
			},
		},
		{
			name: "merge single signature",
			input: MergeSignaturesInput{
				BasePDF: makePDFBytes(t, 1),
				Signatures: []SignatureInfo{
					{
						ImageData: sampleSignature,
						X:         100,
						Y:         200,
						Width:     150,
						Height:    50,
						Page:      1,
					},
				},
			},
			wantErr: false,
			checkFunc: func(result []byte) bool {
				return len(result) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeSignatures(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("MergeSignatures() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				if !tt.checkFunc(result) {
					t.Errorf("MergeSignatures() result validation failed")
				}
			}
		})
	}
}

func TestAppendAuditTrail(t *testing.T) {
	minimalPDF := makePDFBytes(t, 1)

	tests := []struct {
		name        string
		basePDF     []byte
		auditTrail  []byte
		wantErr     bool
		checkResult func([]byte) bool
	}{
		{
			name:       "merge two valid PDFs",
			basePDF:    minimalPDF,
			auditTrail: minimalPDF,
			wantErr:    false,
			checkResult: func(result []byte) bool {
				return len(result) > len(minimalPDF)
			},
		},
		{
			name:       "empty base PDF returns error",
			basePDF:    []byte{},
			auditTrail: minimalPDF,
			wantErr:    true,
		},
		{
			name:       "empty audit trail returns error",
			basePDF:    minimalPDF,
			auditTrail: []byte{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AppendAuditTrail(tt.basePDF, tt.auditTrail)

			if (err != nil) != tt.wantErr {
				t.Errorf("AppendAuditTrail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkResult != nil {
				if !tt.checkResult(result) {
					t.Errorf("AppendAuditTrail() result validation failed")
				}
			}
		})
	}
}

func TestMain(m *testing.M) {
	// Setup test data directory
	testDataDir := "testdata"
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		panic(err)
	}

	// Create a minimal test PDF if it doesn't exist
	testPDFPath := filepath.Join(testDataDir, "sample.pdf")
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		pdf := gopdf.GoPdf{}
		pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
		pdf.AddPage()
		fonts := addStandardFonts(&pdf, "")
		if fonts.NormalOK {
			_ = pdf.SetFont(fonts.NormalName, "", 10)
		}
		pdf.SetXY(50, 50)
		pdf.Cell(nil, "test")
		var buf bytes.Buffer
		if err := pdf.Write(&buf); err != nil {
			panic(err)
		}
		pdfBytes := buf.Bytes()
		if err := os.WriteFile(testPDFPath, pdfBytes, 0644); err != nil {
			panic(err)
		}
	}

	// Run tests
	code := m.Run()

	// Cleanup can be added here if needed

	os.Exit(code)
}

