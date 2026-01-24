package pdf

import (
	"path/filepath"
	"strings"

	"github.com/signintech/gopdf"
)

type standardFonts struct {
	NormalName string
	BoldName   string
	NormalOK   bool
	BoldOK     bool
}

// addStandardFonts tries to register Arial/Arial-Bold from assetsDir (if provided),
// otherwise falls back to common system fonts.
//
// This is used by non-certificate PDF generation code and tests to avoid relying
// on gopdf "core fonts" which can panic in some environments.
func addStandardFonts(pdf *gopdf.GoPdf, assetsDir string) standardFonts {
	const normalName = "Arial"
	const boldName = "Arial-Bold"

	fontSet := standardFonts{NormalName: normalName, BoldName: boldName}

	normalCandidates := []string{}
	if ad := strings.TrimSpace(assetsDir); ad != "" {
		normalCandidates = append(normalCandidates, filepath.Join(ad, "fonts", "Arial.ttf"))
	}
	normalCandidates = append(normalCandidates,
		"/usr/share/fonts/truetype/arial.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"./fonts/Arial.ttf",
	)
	for _, p := range normalCandidates {
		if err := pdf.AddTTFFont(normalName, p); err == nil {
			fontSet.NormalOK = true
			break
		}
	}

	boldCandidates := []string{}
	if ad := strings.TrimSpace(assetsDir); ad != "" {
		boldCandidates = append(boldCandidates, filepath.Join(ad, "fonts", "Arial-Bold.ttf"))
	}
	boldCandidates = append(boldCandidates,
		"/usr/share/fonts/truetype/arialbd.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
		"./fonts/Arial-Bold.ttf",
	)
	for _, p := range boldCandidates {
		if err := pdf.AddTTFFont(boldName, p); err == nil {
			fontSet.BoldOK = true
			break
		}
	}

	return fontSet
}

