package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Embedded contains the packaged assets (fonts, images, backgrounds).
//
//go:embed fonts/* img/*
var Embedded embed.FS

// DefaultOutputDir returns a sensible default directory for extracting embedded assets.
//
// We always use "./assets" (next to ./lc_pages, ./lc_signed, ./lc_uploads).
func DefaultOutputDir() string {
	return filepath.FromSlash("assets")
}

type Paths struct {
	Dir string
	// Fonts
	ArialTTF     string
	ArialBoldTTF string
	// Images
	CertificateBackgroundPDF string
	StampPNG                 string
}

// EnsureOnDisk extracts all embedded assets to dir and returns on-disk paths.
//
// Files are written only when missing or when size differs to keep startup cheap.
func EnsureOnDisk(dir string) (Paths, error) {
	if strings.TrimSpace(dir) == "" {
		return Paths{}, fmt.Errorf("assets dir is empty")
	}
	dir = filepath.Clean(dir)

	// Ensure base directories.
	if err := os.MkdirAll(filepath.Join(dir, "fonts"), 0755); err != nil {
		return Paths{}, fmt.Errorf("failed to create fonts dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "img"), 0755); err != nil {
		return Paths{}, fmt.Errorf("failed to create img dir: %w", err)
	}

	// Extract everything we embed, preserving relative paths.
	if err := fs.WalkDir(Embedded, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Only allow our two top-level directories.
		if !strings.HasPrefix(p, "fonts/") && !strings.HasPrefix(p, "img/") {
			return nil
		}
		dst := filepath.Join(dir, filepath.FromSlash(p))
		if err := writeFileIfDifferent(Embedded, p, dst); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return Paths{}, fmt.Errorf("failed to extract embedded assets: %w", err)
	}

	paths := Paths{Dir: dir}
	paths.ArialTTF = filepath.Join(dir, "fonts", "Arial.ttf")
	paths.ArialBoldTTF = filepath.Join(dir, "fonts", "Arial-Bold.ttf")
	paths.CertificateBackgroundPDF = filepath.Join(dir, "img", "cert-fon.pdf")
	paths.StampPNG = filepath.Join(dir, "img", "stamp.png")

	return paths, nil
}

func writeFileIfDifferent(efs fs.FS, src string, dst string) error {
	b, err := fs.ReadFile(efs, src)
	if err != nil {
		return fmt.Errorf("failed to read embedded %s: %w", src, err)
	}

	// Skip write when size matches (cheap check).
	if st, err := os.Stat(dst); err == nil && st.Size() == int64(len(b)) {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create dir for %s: %w", dst, err)
	}
	if err := os.WriteFile(dst, b, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", dst, err)
	}
	return nil
}

