package appdir

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	mu      sync.RWMutex
	dataDir string
)

// Init sets the data directory for lc_pages, lc_signed, etc.
// Priority: GOSIGN_DATA_DIR env → dev fallback (cmd/server → repo bin/) → executable directory.
func Init() {
	mu.Lock()
	defer mu.Unlock()

	if dir := os.Getenv("GOSIGN_DATA_DIR"); dir != "" {
		dataDir = resolveDataDir(dir)
		return
	}

	execPath, err := os.Executable()
	if err != nil {
		dataDir = "."
		return
	}
	execDir := filepath.Dir(execPath)
	dataDir = execDir

	// IDE/debug runs from cmd/server but built binary stores data in repo bin/.
	if filepath.Base(execDir) == "server" {
		binDir := filepath.Clean(filepath.Join(execDir, "..", "..", "bin"))
		if _, err := os.Stat(filepath.Join(binDir, "lc_pages")); err == nil {
			dataDir = binDir
		}
	}
}

func resolveDataDir(dir string) string {
	if filepath.IsAbs(dir) {
		return dir
	}
	execPath, err := os.Executable()
	if err != nil {
		return dir
	}
	return filepath.Clean(filepath.Join(filepath.Dir(execPath), dir))
}

// DataDir returns the directory where app data (lc_uploads, lc_signed, etc.) should live.
// Defaults to the executable's directory after Init(); returns "." if Init was not called or failed.
func DataDir() string {
	mu.RLock()
	defer mu.RUnlock()
	if dataDir == "" {
		return "."
	}
	return dataDir
}

// LcUploads returns path to local uploads directory (e.g. {DataDir}/lc_uploads).
func LcUploads() string {
	return filepath.Join(DataDir(), "lc_uploads")
}

// LcSigned returns path to signed documents directory (e.g. {DataDir}/lc_signed).
func LcSigned() string {
	return filepath.Join(DataDir(), "lc_signed")
}

// LcPages returns path to pages cache directory (e.g. {DataDir}/lc_pages).
func LcPages() string {
	return filepath.Join(DataDir(), "lc_pages")
}

// LcTmp returns path to temporary files directory (e.g. {DataDir}/lc_tmp).
func LcTmp() string {
	return filepath.Join(DataDir(), "lc_tmp")
}

// Base returns path to base data directory (e.g. {DataDir}/base), used for GeoLite2 etc.
func Base() string {
	return filepath.Join(DataDir(), "base")
}

// GeoLite2 returns path to the GeoLite2 city database file (e.g. {DataDir}/base/GeoLite2-City.mmdb).
func GeoLite2() string {
	return filepath.Join(Base(), "GeoLite2-City.mmdb")
}
