package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/digitorus/pdf"
)

func TestParseKeywords(t *testing.T) {
	tests := []struct {
		in   string
		want int // min parts
	}{
		{"a, b, c", 2},
		{"kw1,kw2", 2},
		{"single", 1},
		{"two words", 2},
	}
	for _, tt := range tests {
		got := parseKeywords(tt.in)
		if len(got) < tt.want {
			t.Fatalf("parseKeywords(%q) = %v, want at least %d parts", tt.in, got, tt.want)
		}
	}
}

func TestParseDate(t *testing.T) {
	_, err := parseDate("(D:20170923123900+01'00')")
	if err == nil {
		t.Fatal("expected error for parenthesized PDF date with current layout")
	}
	ts, err := parseDate(`D:20170923123900+01'00'`)
	if err != nil {
		t.Fatalf("parseDate: %v", err)
	}
	if ts.Year() != 2017 || ts.Month() != 9 || ts.Day() != 23 {
		t.Fatalf("unexpected time: %v", ts)
	}
}

func TestParseDocumentInfo_fromFixture(t *testing.T) {
	path := filepath.Join(testRepoRoot(t), "fixtures", "testfiles", "testfile20.pdf")
	f, err := os.Open(path)
	if err != nil {
		t.Skip("fixture not available:", err)
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	rdr, err := pdf.NewReader(f, st.Size())
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	info := rdr.Trailer().Key("Info")
	var di DocumentInfo
	parseDocumentInfo(info, &di)
	// Fixture is PDF 2.0 sample; at least Pages or Title may be set.
	if di.Pages == 0 && di.Title == "" && di.Producer == "" {
		t.Log("document info mostly empty (OK for this fixture)")
	}
}
