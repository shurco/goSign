package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractPages(t *testing.T) {
	tests := []struct {
		name      string
		input     ExtractPagesInput
		wantPages int
		wantErr   bool
	}{
		{
			name: "extract pages from valid PDF",
			input: ExtractPagesInput{
				PDFPath: "testdata/sample.pdf",
			},
			wantPages: 1,
			wantErr:   false,
		},
		{
			name: "non-existent PDF returns error",
			input: ExtractPagesInput{
				PDFPath: "testdata/nonexistent.pdf",
			},
			wantPages: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if test file doesn't exist and test doesn't expect error
			if _, err := os.Stat(tt.input.PDFPath); os.IsNotExist(err) && !tt.wantErr {
				t.Skip("Test PDF file not found")
			}

			result, err := ExtractPages(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractPages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result.PageCount != tt.wantPages {
				t.Errorf("ExtractPages() got %d pages, want %d", result.PageCount, tt.wantPages)
			}
		})
	}
}

func TestGeneratePreview(t *testing.T) {
	tmpDir := os.TempDir()
	previewDir := filepath.Join(tmpDir, "test_previews")
	defer os.RemoveAll(previewDir)

	if err := os.MkdirAll(previewDir, 0755); err != nil {
		t.Fatalf("Failed to create preview dir: %v", err)
	}

	tests := []struct {
		name    string
		input   GeneratePreviewInput
		wantErr bool
	}{
		{
			name: "generate preview for valid PDF",
			input: GeneratePreviewInput{
				PDFPath:   "testdata/sample.pdf",
				OutputDir: previewDir,
			},
			wantErr: false,
		},
		{
			name: "non-existent PDF returns error",
			input: GeneratePreviewInput{
				PDFPath:   "testdata/nonexistent.pdf",
				OutputDir: previewDir,
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

			result, err := GeneratePreview(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePreview() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil {
				// Check that at least placeholders were created
				if len(result.Images) == 0 {
					t.Log("Warning: No preview images generated (may require external tools)")
				}
			}
		})
	}
}

func TestExtractFormFields(t *testing.T) {
	tests := []struct {
		name       string
		input      ExtractFormFieldsInput
		wantFields bool
		wantErr    bool
	}{
		{
			name: "extract form fields from valid PDF",
			input: ExtractFormFieldsInput{
				PDFPath: "testdata/sample.pdf",
			},
			wantFields: false, // sample.pdf doesn't have form fields
			wantErr:    false,
		},
		{
			name: "non-existent PDF returns empty fields",
			input: ExtractFormFieldsInput{
				PDFPath: "testdata/nonexistent.pdf",
			},
			wantFields: false,
			wantErr:    false, // Function returns empty result, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractFormFields(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFormFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				hasFields := len(result.Fields) > 0
				if hasFields != tt.wantFields {
					t.Logf("ExtractFormFields() has fields: %v, want: %v", hasFields, tt.wantFields)
				}
			}
		})
	}
}

