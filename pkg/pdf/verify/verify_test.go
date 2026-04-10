package verify

import (
	"bytes"
	"testing"
)

func TestReader_notPDF(t *testing.T) {
	r := bytes.NewReader([]byte("not a pdf"))
	_, err := Reader(r, int64(r.Len()))
	if err == nil {
		t.Fatal("expected error for non-PDF input")
	}
}
