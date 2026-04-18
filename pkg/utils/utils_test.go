package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetStringFromMap(t *testing.T) {
	tests := []struct {
		name string
		in   map[string]any
		key  string
		def  string
		want string
	}{
		{"nil map returns default", nil, "k", "d", "d"},
		{"existing string", map[string]any{"k": "v"}, "k", "d", "v"},
		{"missing key returns default", map[string]any{"other": "v"}, "k", "d", "d"},
		{"non-string returns default", map[string]any{"k": 42}, "k", "d", "d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStringFromMap(tt.in, tt.key, tt.def); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetBoolFromMap(t *testing.T) {
	tests := []struct {
		name string
		in   map[string]any
		key  string
		def  bool
		want bool
	}{
		{"nil map returns default", nil, "k", true, true},
		{"native bool", map[string]any{"k": true}, "k", false, true},
		{"string true", map[string]any{"k": "true"}, "k", false, true},
		{"string false", map[string]any{"k": "false"}, "k", true, false},
		{"non-bool non-string returns default", map[string]any{"k": 1}, "k", true, true},
		{"missing key returns default", map[string]any{}, "k", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBoolFromMap(tt.in, tt.key, tt.def); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDaysBetween(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)

	if got := DaysBetween(start, end); got != 10 {
		t.Errorf("expected 10 days, got %d", got)
	}
	if got := DaysBetween(end, start); got != -10 {
		t.Errorf("expected -10 days for reversed range, got %d", got)
	}
	if got := DaysBetween(start, start); got != 0 {
		t.Errorf("expected 0 for identical times, got %d", got)
	}
}

func TestExtName(t *testing.T) {
	cases := map[string]string{
		"file.pdf":                 "pdf",
		"/tmp/dir/name.TAR.GZ":     "GZ",
		"no-ext":                   "",
		"":                         "",
		"/path/with.dots/file.txt": "txt",
	}
	for in, want := range cases {
		if got := ExtName(in); got != want {
			t.Errorf("ExtName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsFileAndIsDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(file, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}

	if !IsDir(dir) {
		t.Error("expected IsDir to return true for temp dir")
	}
	if IsDir(file) {
		t.Error("expected IsDir to return false for a file")
	}
	if !IsFile(file) {
		t.Error("expected IsFile to return true for an existing file")
	}
	if IsFile(dir) {
		t.Error("expected IsFile to return false for a directory")
	}
	if IsFile("") || IsDir("") {
		t.Error("empty path must return false")
	}
	long := make([]byte, 500)
	for i := range long {
		long[i] = 'a'
	}
	if IsFile(string(long)) || IsDir(string(long)) {
		t.Error("overly long paths must return false")
	}
}

func TestMkDirs(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b", "c")

	if err := MkDirs(0o755, a, b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !IsDir(a) || !IsDir(b) {
		t.Fatal("MkDirs did not create both directories")
	}
	// Calling MkDirs again must be a no-op.
	if err := MkDirs(0o755, a, b); err != nil {
		t.Fatalf("repeated MkDirs must be idempotent, got %v", err)
	}
}
