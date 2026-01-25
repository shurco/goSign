package pdf

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/signintech/gopdf"

	"github.com/shurco/gosign/internal/models"
)

// RenderCompletedTemplatePDFInput describes how to render a "completed" PDF
// by overlaying filled values (including signature images) on top of the stored
// per-page PDFs in lc_pages.
//
// Notes:
// - The current goSign storage model stores each PDF page as its own attachment:
//   lc_pages/{attachment_id}/0.pdf
// - Field areas are stored as percentages (0..1) relative to an A4 page.
// - For signature/initials fields, the frontend stores a PNG data URL in the field value.
type RenderCompletedTemplatePDFInput struct {
	PagesDir string // e.g. "./lc_pages"
	Schema   []models.Schema
	Fields   []models.Field
	Values   map[string]any // field_id -> value (string/bool/[]any/etc.)
}

const (
	// A4 size in points (1/72 inch). Must match how we generate/extract pages.
	a4WidthPt  = 595.28
	a4HeightPt = 841.89
)

// RenderCompletedTemplatePDF renders a PDF based on stored page PDFs and overlays
// all filled values on the appropriate pages using template field areas.
//
// It intentionally does NOT require PDF form fields; it uses template-defined areas
// (percent-based coordinates) and draws text/images at those coordinates.
func RenderCompletedTemplatePDF(input RenderCompletedTemplatePDFInput) ([]byte, error) {
	if input.PagesDir == "" {
		return nil, fmt.Errorf("pages dir is required")
	}
	if len(input.Schema) == 0 {
		return nil, fmt.Errorf("template schema is empty")
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	fontSet := addStandardFonts(&pdf, "")
	if fontSet.NormalOK {
		_ = pdf.SetFont(fontSet.NormalName, "", 10)
	} else {
		_ = pdf.SetFont("helvetica", "", 10)
	}

	// Render each stored page and overlay values whose areas target this attachment.
	for _, schemaItem := range input.Schema {
		if schemaItem.AttachmentID == "" {
			continue
		}

		pagePath := filepath.Join(input.PagesDir, schemaItem.AttachmentID, "0.pdf")
		if _, err := os.Stat(pagePath); err != nil {
			return nil, fmt.Errorf("missing page PDF for attachment %s: %w", schemaItem.AttachmentID, err)
		}

		pdf.AddPage()
		tpl := pdf.ImportPage(pagePath, 1, "/MediaBox")
		pdf.UseImportedTemplate(tpl, 0, 0, 0, 0)

		// Overlay all fields that have at least one area on this page attachment.
		for _, field := range input.Fields {
			val, ok := input.Values[field.ID]
			if !ok {
				continue
			}

			for _, area := range field.Areas {
				if area == nil || area.AttachmentID != schemaItem.AttachmentID {
					continue
				}

				// Convert percent-based coordinates to points.
				x := clamp01(area.X) * a4WidthPt
				yTop := clamp01(area.Y) * a4HeightPt
				w := clamp01(area.W) * a4WidthPt
				h := clamp01(area.H) * a4HeightPt
				if h <= 0 {
					// Defensive default: small height so text isn't placed outside the page.
					h = 12
				}

				// gopdf uses bottom-left origin. Convert from top-left.
				y := a4HeightPt - yTop - h

				switch field.Type {
				case models.FieldTypeSignature, models.FieldTypeInitials, models.FieldTypeStamp, models.FieldTypeImage:
					imgBytes, err := decodeImageDataURL(val)
					if err != nil || len(imgBytes) == 0 {
						continue
					}
					rect := &gopdf.Rect{W: w, H: h}
					// Draw image inside the rectangle area (no temp files).
					holder, err := gopdf.ImageHolderByBytes(imgBytes)
					if err != nil {
						continue
					}
					_ = pdf.ImageByHolder(holder, x, y, rect)

					// If "with signature ID" is enabled, draw the signature ID below the image.
					if field.Preferences != nil && field.Preferences.WithSignatureID {
						if sigIDAny, ok := input.Values[field.ID+"_signature_id"]; ok {
							if sigID, ok := sigIDAny.(string); ok && strings.TrimSpace(sigID) != "" {
								_ = pdf.SetFont("helvetica", "", 8)
								idLabel := "ID: " + strings.TrimSpace(sigID)
								// Place text just below the image (smaller y in gopdf = lower on page).
								textY := y - 10
								if textY < 0 {
									textY = y + 2
								}
								pdf.SetXY(x, textY)
								pdf.Cell(nil, idLabel)
							}
						}
						// Restore default font for subsequent fields.
						if fontSet.NormalOK {
							_ = pdf.SetFont(fontSet.NormalName, "", 10)
						} else {
							_ = pdf.SetFont("helvetica", "", 10)
						}
					}

				default:
					text := stringifyValue(val)
					if strings.TrimSpace(text) == "" {
						continue
					}
					// Place text slightly inside the area.
					pdf.SetXY(x+1, y+h-11)
					pdf.Cell(nil, text)
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write PDF: %w", err)
	}
	return buf.Bytes(), nil
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func stringifyValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "Yes"
		}
		return "No"
	case float64, float32, int, int64, int32, uint, uint64, uint32:
		return fmt.Sprint(t)
	case []any:
		parts := make([]string, 0, len(t))
		for _, it := range t {
			s := strings.TrimSpace(fmt.Sprint(it))
			if s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(v)
	}
}

func decodeImageDataURL(v any) ([]byte, error) {
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("image value is not a string")
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("image value is empty")
	}

	// Accept both raw base64 and data URLs.
	if strings.HasPrefix(s, "data:") {
		// Format: data:image/png;base64,AAAA...
		comma := strings.IndexByte(s, ',')
		if comma < 0 {
			return nil, fmt.Errorf("invalid data url")
		}
		s = s[comma+1:]
	}

	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return b, nil
}

