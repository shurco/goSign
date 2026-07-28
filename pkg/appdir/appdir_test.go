package appdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDataDir(t *testing.T) {
	execDir := "/app/cmd/goSign"
	got := resolveDataDir(filepath.Join(execDir, "..", "..", "bin"))
	want := filepath.Clean("/app/bin")
	if got != want {
		t.Fatalf("resolveDataDir() = %q, want %q", got, want)
	}
}

func TestInitUsesEnvOverride(t *testing.T) {
	t.Setenv("GOSIGN_DATA_DIR", "/tmp/gosign-data")
	Init()
	if got, want := DataDir(), "/tmp/gosign-data"; got != want {
		t.Fatalf("DataDir() = %q, want %q", got, want)
	}
	os.Unsetenv("GOSIGN_DATA_DIR")
}
