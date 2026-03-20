package pdf

import (
	"strings"
	"testing"
	"time"
)

func TestFormatCertTime(t *testing.T) {
	if got := formatCertTime(nil); got != "-" {
		t.Fatalf("nil: got %q", got)
	}
	var zero time.Time
	if got := formatCertTime(&zero); got != "-" {
		t.Fatalf("zero: got %q", got)
	}
	ts := time.Date(2020, 3, 15, 14, 30, 0, 0, time.FixedZone("CET", 3600))
	got := formatCertTime(&ts)
	if !strings.Contains(got, "2020") || !strings.Contains(got, "UTC") {
		t.Fatalf("unexpected formatted time: %q", got)
	}
}
