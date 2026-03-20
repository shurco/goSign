package sign

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/digitorus/pdf"
)

func TestFindFirstPage(t *testing.T) {
	input_file, reader := loadHelpersTestPDF(t)
	if input_file == nil || reader == nil {
		t.Errorf("Failed to load test PDF")
		return
	}

	root := reader.Trailer().Key("Root")
	root_keys := root.Keys()
	found_pages := false
	for _, key := range root_keys {
		if key == "Pages" {
			found_pages = true
			break
		}
	}

	if !found_pages {
		input_file.Close()
		t.Errorf("Could not find pages element")
		return
	}

	first_page, err := findFirstPage(root.Key("Pages"))
	if err != nil {
		input_file.Close()
		t.Errorf("Could not find first page")
		return
	}

	first_page_ptr := first_page.GetPtr()
	if first_page_ptr.GetGen() != 0 {
		input_file.Close()
		t.Errorf("First page gen mismatch")
	}

	if first_page_ptr.GetID() != 4 {
		t.Errorf("First page ID mismatch")
	}

	input_file.Close()
}

func TestPDFString(t *testing.T) {
	string_compare := map[string]string{
		"Test":    "(Test)",
		"((Test)": "(\\(\\(Test\\))",
		"\\TEst":  "(\\\\TEst)",
		"\rnew":   "(\\rnew)",
	}

	for text, expected := range string_compare {
		if pdfString(text) != expected {
			t.Errorf("Error while escaping %s. Expected %s, got %s.", text, expected, pdfString(text))
		}
	}
}

func TestPdfDateTime(t *testing.T) {
	timezone, _ := time.LoadLocation("Europe/Tallinn")
	timezone_1, _ := time.LoadLocation("Africa/Casablanca")
	timezone_2, _ := time.LoadLocation("America/New_York")
	timezone_3, _ := time.LoadLocation("Asia/Jerusalem")
	timezone_4, _ := time.LoadLocation("Europe/Amsterdam")
	timezone_5, _ := time.LoadLocation("Pacific/Honolulu")

	now := time.Date(2017, 9, 23, 14, 39, 0, 0, timezone)

	date_compare := map[time.Time]string{
		now.In(timezone_1): "(D:20170923123900+01'00')",
		now.In(timezone_2): "(D:20170923073900-04'00')",
		now.In(timezone_3): "(D:20170923143900+03'00')",
		now.In(timezone_4): "(D:20170923133900+02'00')",
		now.In(timezone_5): "(D:20170923013900-10'00')",
	}

	for date, expected := range date_compare {
		if pdfDateTime(date) != expected {
			t.Errorf("Error while converting date %s to string. Expected %s, got %s.", date.String(), expected, pdfDateTime(date))
		}
	}
}

func TestLeftPad(t *testing.T) {
	string_compare := map[string]string{
		"123456789": "123456789",
		"12345678":  "12345678",
		"1234567":   "_1234567",
		"123456":    "__123456",
		"12345":     "___12345",
		"1234":      "____1234",
		"123":       "_____123",
		"12":        "______12",
		"1":         "_______1",
		"":          "________",
	}

	for text, expected := range string_compare {
		if leftPad(text, "_", 8-len(text)) != expected {
			t.Errorf("Error while left padding %s. Expected %s, got %s.", text, expected, leftPad(text, "_", 8-len(text)))
		}
	}
}

func TestWritePartFromSourceFileToTargetFile(t *testing.T) {
	input_file, err := os.Open(testPDFFixturePath(t, "testfile20.pdf"))
	if err != nil {
		t.Errorf("Failed to load test PDF")
		return
	}
	defer input_file.Close()

	write := func(offset, length int64) string {
		var b bytes.Buffer
		_ = writePartFromSourceFileToTargetFile(input_file, &b, offset, length)
		return b.String()
	}

	if n := len(write(0, 0)); n != 0 {
		t.Errorf("Expected 0 bytes for length=0, got %d", n)
	}

	if n := len(write(0, -20)); n != 0 {
		t.Errorf("Expected 0 bytes for length=-20, got %d", n)
	}

	if got := write(0, 8); got != "%PDF-2.0" {
		t.Errorf("Wrong content at offset 0 len 8: got %q, want %q", got, "%PDF-2.0")
	}

	if got := write(33, 8); got != "/Catalog" {
		t.Errorf("Wrong content at offset 33 len 8: got %q, want %q", got, "/Catalog")
	}

	if n := len(write(0, 1200)); n != 1200 {
		t.Errorf("Requested 1200 bytes but only got %d", n)
	}
}

func loadHelpersTestPDF(t *testing.T) (*os.File, *pdf.Reader) {
	t.Helper()
	input_file, err := os.Open(testPDFFixturePath(t, "testfile20.pdf"))
	if err != nil {
		return nil, nil
	}

	finfo, err := input_file.Stat()
	if err != nil {
		input_file.Close()
		return nil, nil
	}
	size := finfo.Size()

	rdr, err := pdf.NewReader(input_file, size)
	if err != nil {
		input_file.Close()
		return nil, nil
	}

	return input_file, rdr
}
