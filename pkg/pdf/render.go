package pdf

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/signintech/gopdf"

	"github.com/shurco/gosign/internal/models"
)

// RenderCompletedTemplatePDFInput describes how to render a "completed" PDF
// by overlaying filled values (including signature images) on top of the stored
// per-page PDFs in lc_pages.
//
// Notes:
//   - The current goSign storage model stores each PDF page as its own attachment:
//     lc_pages/{attachment_id}/0.pdf
//   - Field areas are stored as percentages (0..1) relative to an A4 page,
//     measured from the TOP-LEFT corner (same as the web editor).
//   - For signature/initials fields, the frontend stores a PNG data URL in the field value.
type RenderCompletedTemplatePDFInput struct {
	PagesDir string // e.g. "./lc_pages"
	Schema   []models.Schema
	Fields   []models.Field
	Values   map[string]any // field_id -> value (string/bool/[]any/etc.)
}

const defaultFieldFontSize = 10.0

// RenderCompletedTemplatePDF renders a PDF based on stored page PDFs and overlays
// all filled values on the appropriate pages using template field areas.
//
// It intentionally does NOT require PDF form fields; it uses template-defined areas
// (percent-based coordinates) and draws text/images/checkmarks at those coordinates.
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
	fontSet.SetNormal(&pdf, int(defaultFieldFontSize))

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
				// gopdf uses a top-left origin (y grows downward), which matches
				// the editor's area coordinates directly.
				x := clamp01(area.X) * A4WidthPt
				y := clamp01(area.Y) * A4HeightPt
				w := clamp01(area.W) * A4WidthPt
				h := clamp01(area.H) * A4HeightPt
				if h <= 0 {
					// Defensive default: small height so text isn't placed outside the page.
					h = 12
				}

				renderFieldArea(&pdf, fontSet, input.Values, field, area, val, x, y, w, h)
			}
		}
	}

	var buf bytes.Buffer
	if _, err := pdf.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("failed to write PDF: %w", err)
	}
	return buf.Bytes(), nil
}

// renderFieldArea draws a single field area onto the current page.
func renderFieldArea(pdf *gopdf.GoPdf, fontSet standardFonts, values map[string]any, field models.Field, area *models.Areas, val any, x, y, w, h float64) {
	switch field.Type {
	case models.FieldTypeSignature, models.FieldTypeInitials, models.FieldTypeStamp, models.FieldTypeImage:
		imgBytes, err := decodeImageDataURL(val)
		if err != nil || len(imgBytes) == 0 {
			return
		}
		holder, err := gopdf.ImageHolderByBytes(imgBytes)
		if err != nil {
			return
		}
		_ = pdf.ImageByHolder(holder, x, y, &gopdf.Rect{W: w, H: h})

		// If "with signature ID" is enabled, draw the signature ID below the image.
		if field.Preferences != nil && field.Preferences.WithSignatureID {
			if sigIDAny, ok := values[field.ID+"_signature_id"]; ok {
				if sigID, ok := sigIDAny.(string); ok && strings.TrimSpace(sigID) != "" {
					fontSet.SetNormal(pdf, 8)
					idLabel := "ID: " + strings.TrimSpace(sigID)
					textY := y + h + 2
					if textY+10 > A4HeightPt {
						textY = y + h - 10
					}
					pdf.SetXY(x, textY)
					_ = pdf.Cell(nil, idLabel)
					fontSet.SetNormal(pdf, int(defaultFieldFontSize))
				}
			}
		}

	case models.FieldTypeCheckbox:
		if isTruthyValue(val) {
			drawCheckmark(pdf, x, y, w, h)
		}

	case models.FieldTypeRadio, models.FieldTypeSelect, models.FieldTypeMultiple, models.FieldTypeMultiSelect:
		if area.OptionID != nil && *area.OptionID != "" {
			// Per-option area: mark only when this option is selected.
			optValue := optionValueByID(field, *area.OptionID)
			if optValue != "" && valueContains(val, optValue) {
				drawCheckmark(pdf, x, y, w, h)
			}
			return
		}
		drawFieldText(pdf, fontSet, field, stringifyValue(val), x, y, w, h)

	case models.FieldTypeCells:
		drawCells(pdf, fontSet, field, area, stringifyValue(val), x, y, w, h)

	case models.FieldTypeDate:
		text := formatDateValue(stringifyValue(val), fieldFormat(field))
		drawFieldText(pdf, fontSet, field, text, x, y, w, h)

	case models.FieldTypeNumber:
		text := formatNumberValue(stringifyValue(val), fieldFormat(field))
		drawFieldText(pdf, fontSet, field, text, x, y, w, h)

	default:
		drawFieldText(pdf, fontSet, field, stringifyValue(val), x, y, w, h)
	}
}

// drawFieldText renders single/multi-line text inside the area applying
// font size, style, alignment and color preferences.
func drawFieldText(pdf *gopdf.GoPdf, fontSet standardFonts, field models.Field, text string, x, y, w, h float64) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	size := defaultFieldFontSize
	align := ""
	fontType := ""
	colorHex := ""
	if field.Preferences != nil {
		if field.Preferences.FontSize > 0 {
			size = float64(field.Preferences.FontSize)
		}
		align = field.Preferences.Align
		fontType = field.Preferences.FontType
		colorHex = field.Preferences.Color
	}

	if strings.HasPrefix(fontType, "bold") {
		fontSet.SetBold(pdf, int(size))
	} else {
		fontSet.SetNormal(pdf, int(size))
	}

	if r, g, b, ok := parseHexColor(colorHex); ok {
		pdf.SetTextColor(r, g, b)
	} else {
		pdf.SetTextColor(0, 0, 0)
	}

	tw, err := pdf.MeasureTextWidth(text)
	if err != nil {
		tw = 0
	}

	// Vertical center for a single line (top-left origin: y grows downward).
	textY := y + (h-size)/2
	if textY < y {
		textY = y
	}

	if tw > w-4 && w > 20 {
		// Long text: wrap inside the area.
		pdf.SetXY(x+2, y+1)
		_ = pdf.MultiCell(&gopdf.Rect{W: w - 4, H: math.Max(h, size+2)}, text)
	} else {
		textX := x + 2
		switch align {
		case "center":
			textX = x + (w-tw)/2
		case "right":
			textX = x + w - tw - 2
		}
		if textX < x {
			textX = x
		}
		pdf.SetXY(textX, textY)
		_ = pdf.Cell(nil, text)
	}

	// Restore defaults for subsequent draws.
	pdf.SetTextColor(0, 0, 0)
	fontSet.SetNormal(pdf, int(defaultFieldFontSize))
}

// drawCells renders one character per cell across the area (DocuSeal "cells" field).
func drawCells(pdf *gopdf.GoPdf, fontSet standardFonts, field models.Field, area *models.Areas, text string, x, y, w, h float64) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	cellW := 0.0
	if area.CellW != nil && *area.CellW > 0 {
		cellW = *area.CellW * A4WidthPt
	}
	count := 0
	if area.CellCount != nil && *area.CellCount > 0 {
		count = *area.CellCount
	}
	if cellW <= 0 && count > 0 {
		cellW = w / float64(count)
	}
	if cellW <= 0 {
		cellW = h // square cells fallback
	}
	if count <= 0 {
		count = int(math.Max(1, math.Floor(w/cellW+0.5)))
	}

	size := defaultFieldFontSize
	if field.Preferences != nil && field.Preferences.FontSize > 0 {
		size = float64(field.Preferences.FontSize)
	}
	fontSet.SetNormal(pdf, int(size))

	runes := []rune(text)
	textY := y + (h-size)/2
	if textY < y {
		textY = y
	}
	for i := 0; i < len(runes) && i < count; i++ {
		ch := string(runes[i])
		chW, err := pdf.MeasureTextWidth(ch)
		if err != nil {
			chW = size / 2
		}
		cx := x + float64(i)*cellW + (cellW-chW)/2
		pdf.SetXY(cx, textY)
		_ = pdf.Cell(nil, ch)
	}
	fontSet.SetNormal(pdf, int(defaultFieldFontSize))
}

// drawCheckmark draws a vector check mark centered in the area
// (works with any font since no glyphs are involved).
func drawCheckmark(pdf *gopdf.GoPdf, x, y, w, h float64) {
	size := math.Min(w, h)
	if size <= 2 {
		size = 8
	}
	s := size * 0.6
	cx := x + w/2
	cy := y + h/2

	x1 := cx - 0.32*s
	y1 := cy + 0.05*s
	x2 := cx - 0.08*s
	y2 := cy + 0.30*s
	x3 := cx + 0.36*s
	y3 := cy - 0.28*s

	pdf.SetLineWidth(math.Max(1, s*0.14))
	pdf.SetStrokeColor(17, 24, 39)
	pdf.Line(x1, y1, x2, y2)
	pdf.Line(x2, y2, x3, y3)
	// Restore stroke defaults.
	pdf.SetLineWidth(1)
	pdf.SetStrokeColor(0, 0, 0)
}

// fieldFormat returns preferences.format or "".
func fieldFormat(field models.Field) string {
	if field.Preferences == nil {
		return ""
	}
	return field.Preferences.Format
}

// formatDateValue converts an ISO date (yyyy-mm-dd or RFC3339) to the
// editor's display pattern (e.g. DD/MM/YYYY, MMMM D, YYYY).
func formatDateValue(value, pattern string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if pattern == "" {
		pattern = "DD/MM/YYYY"
	}

	var t time.Time
	var err error
	if len(value) >= 10 {
		t, err = time.Parse("2006-01-02", value[:10])
	} else {
		err = fmt.Errorf("short value")
	}
	if err != nil {
		return value
	}

	layout := strings.NewReplacer(
		"YYYY", "2006",
		"YY", "06",
		"MMMM", "January",
		"MMM", "Jan",
		"MM", "01",
		"DD", "02",
		"D", "2",
	).Replace(pattern)
	return t.Format(layout)
}

// formatNumberValue formats a numeric string per the editor's number format
// (comma, dot, space, usd, eur, gbp, percent).
func formatNumberValue(value, format string) string {
	value = strings.TrimSpace(value)
	if value == "" || format == "" || format == "none" {
		return value
	}
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value
	}

	switch format {
	case "comma":
		return groupNumber(num, ",", ".", -1)
	case "dot":
		return groupNumber(num, ".", ",", -1)
	case "space":
		return groupNumber(num, " ", ",", -1)
	case "usd":
		return "$" + groupNumber(num, ",", ".", 2)
	case "eur":
		return "€" + groupNumber(num, ".", ",", 2)
	case "gbp":
		return "£" + groupNumber(num, ",", ".", 2)
	case "percent":
		return groupNumber(num, "", ".", -1) + "%"
	default:
		return value
	}
}

// groupNumber renders num with the given thousands/decimal separators.
// decimals < 0 keeps the value's natural precision.
func groupNumber(num float64, thousandsSep, decimalSep string, decimals int) string {
	s := strconv.FormatFloat(num, 'f', decimals, 64)

	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")

	intPart := s
	fracPart := ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, fracPart = s[:i], s[i+1:]
	}

	if thousandsSep != "" && len(intPart) > 3 {
		var b strings.Builder
		pre := len(intPart) % 3
		if pre > 0 {
			b.WriteString(intPart[:pre])
		}
		for i := pre; i < len(intPart); i += 3 {
			if b.Len() > 0 {
				b.WriteString(thousandsSep)
			}
			b.WriteString(intPart[i : i+3])
		}
		intPart = b.String()
	}

	out := intPart
	if fracPart != "" {
		out += decimalSep + fracPart
	}
	if neg {
		out = "-" + out
	}
	return out
}

// optionValueByID resolves an option's display value by its ID.
// Empty stored values fall back to "Option N" (same as the editor placeholder).
func optionValueByID(field models.Field, optionID string) string {
	for i, opt := range field.Options {
		if opt.ID == optionID {
			if strings.TrimSpace(opt.Value) != "" {
				return opt.Value
			}
			return fmt.Sprintf("Option %d", i+1)
		}
	}
	return ""
}

// valueContains reports whether the submitted value selects the given option:
// exact match for radio/select strings, membership for multi-choice arrays.
func valueContains(val any, option string) bool {
	switch t := val.(type) {
	case string:
		return t == option
	case []any:
		for _, it := range t {
			if fmt.Sprint(it) == option {
				return true
			}
		}
		return false
	case []string:
		for _, it := range t {
			if it == option {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func isTruthyValue(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "1" || t == "yes"
	default:
		return false
	}
}

func parseHexColor(s string) (r, g, b uint8, ok bool) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "#"))
	if len(s) != 6 {
		return 0, 0, 0, false
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return uint8(v >> 16), uint8(v >> 8), uint8(v), true //nolint:gosec // masked to 8 bits
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
		// Fallback to URL-safe alphabet (- and _ instead of + and /)
		b, err = base64.URLEncoding.DecodeString(s)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}
