package pdf

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestClamp01(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{-1, 0},
		{0, 0},
		{0.5, 0.5},
		{1, 1},
		{2, 1},
	}
	for _, tt := range tests {
		if got := clamp01(tt.in); got != tt.want {
			t.Errorf("clamp01(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestStringifyValue(t *testing.T) {
	tests := []struct {
		v    any
		want string
	}{
		{"x", "x"},
		{true, "Yes"},
		{false, "No"},
		{42, "42"},
		{[]any{"a", "", "b"}, "a, b"},
	}
	for _, tt := range tests {
		if got := stringifyValue(tt.v); got != tt.want {
			t.Errorf("stringifyValue(%#v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestDecodeImageDataURL(t *testing.T) {
	raw := []byte{1, 2, 3}
	b64 := base64.StdEncoding.EncodeToString(raw)

	tests := []struct {
		name    string
		v       any
		want    []byte
		wantErr bool
	}{
		{name: "data url", v: "data:image/png;base64," + b64, want: raw},
		{name: "raw base64", v: b64, want: raw},
		{name: "empty string", v: "", wantErr: true},
		{name: "not string", v: 1, wantErr: true},
		{name: "bad data url", v: "data:image/png", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeImageDataURL(tt.v)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeImageDataURL_whitespaceTrim(t *testing.T) {
	b64 := base64.StdEncoding.EncodeToString([]byte{9})
	got, err := decodeImageDataURL("  " + b64 + "  ")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != 9 {
		t.Fatalf("got %v", got)
	}
}

func TestStringifyValue_defaultFmtSprint(t *testing.T) {
	type custom struct{ n int }
	if got := stringifyValue(custom{n: 7}); !strings.Contains(got, "7") {
		t.Fatalf("expected fmt.Sprint fallback, got %q", got)
	}
}
