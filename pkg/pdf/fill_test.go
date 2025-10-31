package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

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
				BasePDF:    []byte("%PDF-1.4\n"),
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
				BasePDF: []byte("%PDF-1.4\n"),
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
	// Minimal valid PDF content
	minimalPDF := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Count 0\n/Kids []\n>>\nendobj\nxref\n0 3\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \ntrailer\n<<\n/Size 3\n/Root 1 0 R\n>>\nstartxref\n109\n%%EOF\n")

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
		minimalPDF := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Count 1\n/Kids [3 0 R]\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n>>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \ntrailer\n<<\n/Size 4\n/Root 1 0 R\n>>\nstartxref\n189\n%%EOF\n")
		if err := os.WriteFile(testPDFPath, minimalPDF, 0644); err != nil {
			panic(err)
		}
	}

	// Run tests
	code := m.Run()

	// Cleanup can be added here if needed

	os.Exit(code)
}

