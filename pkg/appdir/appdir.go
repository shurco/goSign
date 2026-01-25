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

// Init sets the data directory to the directory of the running executable.
// Called once at app startup. If os.Executable() fails, current working directory is used.
func Init() {
	mu.Lock()
	defer mu.Unlock()
	execPath, err := os.Executable()
	if err != nil {
		dataDir = "."
		return
	}
	dataDir = filepath.Dir(execPath)
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
