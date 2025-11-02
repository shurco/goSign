package utils

import (
	"os"
	"path"
)

// IsFile reports whether the named file exists
func IsFile(path string) bool {
	if path == "" || len(path) > 468 {
		return false
	}

	if fi, err := os.Stat(path); err == nil {
		return !fi.IsDir()
	}
	return false
}

// IsDir reports whether the named directory exists
func IsDir(path string) bool {
	if path == "" || len(path) > 468 {
		return false
	}

	if fi, err := os.Stat(path); err == nil {
		return fi.IsDir()
	}
	return false
}

// ExtName returns file extension without dot
func ExtName(fpath string) string {
	if ext := path.Ext(fpath); len(ext) > 0 {
		return ext[1:]
	}
	return ""
}

// MkDirs batch create multiple directories at once
func MkDirs(perm os.FileMode, dirPaths ...string) error {
	for _, dirPath := range dirPaths {
		if !IsDir(dirPath) {
			if err := os.MkdirAll(dirPath, perm); err != nil {
				return err
			}
		}
	}
	return nil
}
