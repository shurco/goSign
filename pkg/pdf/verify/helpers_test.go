package verify

import (
	"path/filepath"
	"runtime"
	"testing"
)

func testRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// .../pkg/pdf/verify/*.go -> repository root
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}
