package pdf

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shurco/gosign/internal/models"
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
			if _, err := os.Stat(tt.input.PDFPath); os.IsNotExist(err) && !tt.wantErr {
				t.Skip("Test PDF file not found")
			}

			result, err := FillFields(tt.input)
			if !assertErrorWants(t, err, tt.wantErr) {
				return
			}
			if !tt.wantErr && tt.checkFunc != nil && !tt.checkFunc(result) {
				t.Errorf("FillFields() result validation failed")
			}
		})
	}
}

func TestMergeSignatures(t *testing.T) {
	sampleSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	tests := []struct {
		name      string
		input     MergeSignaturesInput
		wantErr   bool
		checkFunc func([]byte) bool
	}{
		{
			name: "empty signatures returns original PDF",
			input: MergeSignaturesInput{
				BasePDF:    buildTestPDF(t, 1, ""),
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
				BasePDF: buildTestPDF(t, 1, ""),
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
			if !assertErrorWants(t, err, tt.wantErr) {
				return
			}
			if !tt.wantErr && tt.checkFunc != nil && !tt.checkFunc(result) {
				t.Errorf("MergeSignatures() result validation failed")
			}
		})
	}
}

func TestAppendAuditTrail(t *testing.T) {
	minimalPDF := buildTestPDF(t, 1, "")

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
			if !assertErrorWants(t, err, tt.wantErr) {
				return
			}
			if !tt.wantErr && tt.checkResult != nil && !tt.checkResult(result) {
				t.Errorf("AppendAuditTrail() result validation failed")
			}
		})
	}
}

func TestGenerateAuditTrail(t *testing.T) {
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	done := ts.Add(time.Hour)
	sub := &models.Submission{ID: "sub-1", CreatedAt: ts}
	signers := []*models.Submitter{
		{Name: "Alice", Email: "a@example.com", CompletedAt: &done},
	}
	events := []*models.Event{
		{Type: models.EventSubmissionCreated, CreatedAt: ts},
	}

	out, err := GenerateAuditTrail(GenerateAuditTrailInput{
		Submission: sub,
		Submitters: signers,
		Events:     events,
	})
	if err != nil {
		t.Fatalf("GenerateAuditTrail: %v", err)
	}
	assertPDFHeader(t, out)
	if n := pageCountFromBytes(t, out); n < 1 {
		t.Fatalf("expected ≥1 page, got %d", n)
	}
}

func TestAssembleDocument(t *testing.T) {
	base := buildTestPDF(t, 1, "")
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	ts := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)

	out, err := AssembleDocument(base, []SignatureInfo{
		{Page: 1, X: 10, Y: 10, Width: 20, Height: 20, ImageData: pngHeader},
	}, GenerateAuditTrailInput{
		Submission: &models.Submission{ID: "s1", CreatedAt: ts},
		Submitters: []*models.Submitter{{Name: "Bob", Email: "b@example.com"}},
		Events:     nil,
	})
	if err != nil {
		t.Fatalf("AssembleDocument: %v", err)
	}
	assertPDFHeader(t, out)
	if n := pageCountFromBytes(t, out); n < 2 {
		t.Fatalf("expected at least base + audit pages, got %d", n)
	}
}

func TestMain(m *testing.M) {
	testDataDir := "testdata"
	if err := os.MkdirAll(testDataDir, 0o755); err != nil {
		panic(err)
	}
	testPDFPath := filepath.Join(testDataDir, "sample.pdf")
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		if b, err := buildTestPDFData(1, "test"); err == nil {
			_ = os.WriteFile(testPDFPath, b, 0o644)
		}
	}
	os.Exit(m.Run())
}
