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

	ad := strings.TrimSpace(assetsDir)

	normalCandidates := []string{}
	if ad != "" {
		normalCandidates = append(normalCandidates, filepath.Join(ad, "fonts", "Arial.ttf"))
	}
	normalCandidates = append(normalCandidates,
		"/usr/share/fonts/truetype/arial.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/Library/Fonts/Arial.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"./fonts/Arial.ttf",
	)
	for _, p := range normalCandidates {
		if err := pdf.AddTTFFont(normalName, p); err == nil {
			fontSet.NormalOK = true
			break
		}
	}

	boldCandidates := []string{}
	if ad != "" {
		boldCandidates = append(boldCandidates, filepath.Join(ad, "fonts", "Arial-Bold.ttf"))
	}
	boldCandidates = append(boldCandidates,
		"/usr/share/fonts/truetype/arialbd.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
		"/Library/Fonts/Arial Bold.ttf",
		"/System/Library/Fonts/Supplemental/Arial Bold.ttf",
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

// SetNormal applies the registered normal TTF or falls back to core Helvetica.
func (f standardFonts) SetNormal(pdf *gopdf.GoPdf, size int) {
	if f.NormalOK {
		_ = pdf.SetFont(f.NormalName, "", size)
		return
	}
	_ = pdf.SetFont("helvetica", "", size)
}

// SetBold applies the registered bold TTF or falls back to Helvetica bold style.
func (f standardFonts) SetBold(pdf *gopdf.GoPdf, size int) {
	if f.BoldOK {
		_ = pdf.SetFont(f.BoldName, "", size)
		return
	}
	_ = pdf.SetFont("helvetica", "B", size)
}
