package pdf

import (
	"testing"
)

func TestAppendPDF(t *testing.T) {
	base := buildTestPDF(t, 1, "")
	extra := buildTestPDF(t, 2, "")

	tests := []struct {
		name    string
		base    []byte
		extra   []byte
		wantErr bool
		check   func(t *testing.T, out []byte)
	}{
		{
			name:  "append extra pages to base",
			base:  base,
			extra: extra,
			check: func(t *testing.T, out []byte) {
				t.Helper()
				assertPDFHeader(t, out)
				if n := pageCountFromBytes(t, out); n < 3 {
					t.Fatalf("expected at least 3 pages, got %d", n)
				}
			},
		},
		{
			name:    "empty base",
			base:    nil,
			extra:   extra,
			wantErr: true,
		},
		{
			name:    "empty extra",
			base:    base,
			extra:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := AppendPDF(tt.base, tt.extra)
			if !assertErrorWants(t, err, tt.wantErr) {
				return
			}
			if tt.wantErr {
				return
			}
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}
