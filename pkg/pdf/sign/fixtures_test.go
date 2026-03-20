package sign

import (
	"path/filepath"
	"runtime"
	"testing"
)

// testRepoRoot returns the repository root (directory that contains go.mod).
// Works for any *_test.go file under pkg/pdf/sign/.
func testRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	signDir := filepath.Dir(file)
	return filepath.Clean(filepath.Join(signDir, "..", "..", ".."))
}

func testPDFFixturesDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(testRepoRoot(t), "fixtures", "testfiles")
}

func testPDFFixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(testPDFFixturesDir(t), name)
}
