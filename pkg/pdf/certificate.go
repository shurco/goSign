package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"github.com/skip2/go-qrcode"
)

type SignatureCertificateSigner struct {
	Name       string
	Email      string
	IP         string
	SentAt     *time.Time
	OpenedAt   *time.Time
	CompletedAt *time.Time
	Location   string

	// SignatureValue is expected to be a PNG data URL (or raw base64) as stored by the frontend
	// for signature/initials/stamp/image fields.
	SignatureValue any
}

type SignatureCertificateInput struct {
	// DocumentName is shown in the certificate header.
	DocumentName string
	// Reference is typically the submission id.
	Reference string
	// CompletedAt is the "document completed" timestamp (usually max(completed_at) among signers).
	CompletedAt *time.Time
	// AssetsDir is a directory that contains:
	// - fonts/Arial.ttf, fonts/Arial-Bold.ttf
	// - img/cert-fon.pdf, img/stamp.png
	//
	// Required to render the certificate in the exact "pdf-cert" design.
	AssetsDir string
	// QRURL is encoded into the QR code on the certificate.
	// Must be an absolute URL to work reliably on mobile QR scanners.
	QRURL string
	Signers     []SignatureCertificateSigner
}

// GenerateSignatureCertificatePDF renders a certificate page using the exact same design
// as cmd/pdf-cert (same background, fonts, coordinates, footer, QR, etc.).
func GenerateSignatureCertificatePDF(input SignatureCertificateInput) ([]byte, error) {
	assetsDir := strings.TrimSpace(input.AssetsDir)
	if assetsDir == "" {
		return nil, fmt.Errorf("assets dir is required for certificate rendering")
	}
	qrURL := strings.TrimSpace(input.QRURL)
	if qrURL == "" {
		return nil, fmt.Errorf("qr url is required for certificate rendering")
	}

	arial := filepath.Join(assetsDir, "fonts", "Arial.ttf")
	arialBold := filepath.Join(assetsDir, "fonts", "Arial-Bold.ttf")
	bgPDF := filepath.Join(assetsDir, "img", "cert-fon.pdf")
	stampPNG := filepath.Join(assetsDir, "img", "stamp.png")

	// Fail fast if any asset is missing (we want 1:1 design).
	for _, p := range []string{arial, arialBold, bgPDF, stampPNG} {
		if _, err := os.Stat(p); err != nil {
			return nil, fmt.Errorf("missing certificate asset %s: %w", p, err)
		}
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	if err := pdf.AddTTFFont("Arial", arial); err != nil {
		return nil, fmt.Errorf("failed to add Arial font: %w", err)
	}
	if err := pdf.AddTTFFont("Arial-Bold", arialBold); err != nil {
		return nil, fmt.Errorf("failed to add Arial-Bold font: %w", err)
	}

	// Render in chunks (max 5 signers per page, same as the example).
	for start := 0; start < len(input.Signers); start += 5 {
		end := start + 5
		if end > len(input.Signers) {
			end = len(input.Signers)
		}
		signers := input.Signers[start:end]

		pdf.AddPage()
		certTpl := pdf.ImportPage(bgPDF, 1, "/MediaBox")
		pdf.UseImportedTemplate(certTpl, 0, 0, 0, 0)

		// ------ START HEADER -----
		pdf.SetFillColor(0, 0, 0)
		_ = pdf.SetFont("Arial-Bold", "", 20)
		pdf.SetXY(80, 80)
		pdf.Cell(nil, "Signature Certificate")

		_ = pdf.SetFont("Arial", "", 8)
		pdf.SetXY(80, 110)
		pdf.SetTextColor(109, 109, 109)
		pdf.Cell(nil, "Reference number:")

		pdf.SetXY(150, 110)
		pdf.SetTextColor(0, 0, 0)
		pdf.Cell(nil, input.Reference)

		_ = pdf.SetFont("Arial-Bold", "", 8)
		pdf.SetXY(80, 155)
		pdf.Cell(nil, "Signer")
		pdf.SetXY(225, 155)
		pdf.Cell(nil, "Timestamp")
		pdf.SetXY(370, 155)
		pdf.Cell(nil, "Signature")

		pdf.SetLineWidth(0.5)
		pdf.SetLineType("solid")
		pdf.SetStrokeColor(153, 153, 153)
		pdf.Line(75, 167, 525, 167)
		// ------ END HEADER -----

		// ---- START BLOCK ------
		var shiftSignerBlock float64
		for _, signer := range signers {
			shiftSignerBlock = shiftSignerBlock + 98

			pdf.SetLineWidth(0.5)
			pdf.SetLineType("solid")
			pdf.SetStrokeColor(153, 153, 153)
			pdf.Line(75, 167+shiftSignerBlock, 525, 167+shiftSignerBlock)

			pdf.SetFillColor(255, 255, 255)
			pdf.SetTransparency(gopdf.Transparency{
				Alpha:         0.6,
				BlendModeType: gopdf.Overlay,
			})
			pdf.Rectangle(75, 69+shiftSignerBlock, 525, 167+shiftSignerBlock, "F", 0, 0)
			pdf.ClearTransparency()

			_ = pdf.SetFont("Arial-Bold", "", 10)
			pdf.SetXY(83, 77+shiftSignerBlock)
			pdf.Cell(nil, signer.Name)
			_ = pdf.SetFont("Arial", "", 7)
			pdf.SetXY(83, 89+shiftSignerBlock)
			pdf.Cell(nil, "Email:")
			pdf.SetXY(105, 89+shiftSignerBlock)
			pdf.Cell(nil, signer.Email)

			pdf.SetXY(83, 105+shiftSignerBlock)
			pdf.Cell(nil, "Sent:")
			pdf.SetXY(225, 105+shiftSignerBlock)
			pdf.Cell(nil, formatCertTime(signer.SentAt))

			pdf.SetXY(83, 115+shiftSignerBlock)
			pdf.Cell(nil, "Viewed:")
			pdf.SetXY(225, 115+shiftSignerBlock)
			pdf.Cell(nil, formatCertTime(signer.OpenedAt))

			pdf.SetXY(83, 125+shiftSignerBlock)
			pdf.Cell(nil, "Signed:")
			pdf.SetXY(225, 125+shiftSignerBlock)
			pdf.Cell(nil, formatCertTime(signer.CompletedAt))

			_ = pdf.SetFont("Arial-Bold", "", 8)
			pdf.SetXY(83, 142+shiftSignerBlock)
			pdf.Cell(nil, "Recipient Verification:")
			_ = pdf.SetFont("Arial", "", 7)
			pdf.SetXY(90, 153+shiftSignerBlock)
			pdf.Cell(nil, "Email verified")
			pdf.SetXY(83, 153+shiftSignerBlock)
			pdf.Cell(nil, "x")
			pdf.SetXY(225, 153+shiftSignerBlock)
			pdf.Cell(nil, formatCertTime(signer.OpenedAt))

			// signature
			pdf.SetFillColor(255, 255, 255)
			pdf.SetLineWidth(0.5)
			pdf.SetLineType("dotted")
			pdf.Rectangle(370, 77+shiftSignerBlock, 515, 132+shiftSignerBlock, "FD", 0, 0)

			// Put example signature image inside the rectangle when available.
			if imgBytes, err := decodeImageDataURL(signer.SignatureValue); err == nil && len(imgBytes) > 0 {
				if holder, err := gopdf.ImageHolderByBytes(imgBytes); err == nil {
					_ = pdf.ImageByHolder(holder, 373, 80+shiftSignerBlock, &gopdf.Rect{W: 139, H: 49})
				}
			}

			pdf.SetXY(370, 143+shiftSignerBlock)
			pdf.Cell(nil, "IP address:")
			pdf.SetXY(408, 143+shiftSignerBlock)
			pdf.Cell(nil, signer.IP)
			pdf.SetXY(370, 153+shiftSignerBlock)
			pdf.Cell(nil, "Location:")
			pdf.SetXY(400, 153+shiftSignerBlock)
			pdf.Cell(nil, signer.Location)
		}
		// ---- END BLOCK ------

		// Completion line (same position logic as example).
		completedAt := formatCertTime(input.CompletedAt)
		_ = pdf.SetFont("Arial", "", 8)
		pdf.SetXY(80, 167+shiftSignerBlock+15)
		pdf.SetTextColor(109, 109, 109)
		pdf.Cell(nil, "Document completed by all parties on:")
		pdf.SetXY(220, 167+shiftSignerBlock+15)
		pdf.SetTextColor(0, 0, 0)
		pdf.Cell(nil, completedAt)

		// ------ START FOOTER -----
		pdf.SetFillColor(71, 170, 98)
		pdf.Rectangle(80, 700, 145, 765, "F", 0, 0)
		_ = pdf.Image(stampPNG, 85, 705, &gopdf.Rect{W: 55, H: 55})

		_ = pdf.SetFont("Arial-Bold", "", 8)
		pdf.SetXY(160, 717)
		pdf.Cell(nil, "Signed with goSign")
		_ = pdf.SetFont("Arial", "", 7)
		pdf.SetXY(160, 737)
		pdf.Text("goSign is an open-source solution for easy")
		pdf.SetXY(160, 747)
		pdf.Text("and secure document signing with eSignature.")

		qrCode, _ := qrcode.Encode(qrURL, qrcode.Medium, 256)
		imgQRCode, _ := gopdf.ImageHolderByBytes(qrCode)
		_ = pdf.ImageByHolder(imgQRCode, 460, 700, &gopdf.Rect{W: 65, H: 65})
		// ------ END FOOTER -----
	}

	// ---------
	pdf.SetInfo(gopdf.PdfInfo{
		Title:    input.DocumentName,
		Author:   "goSign (https://github.com/shurco/goSign)",
		Subject:  "Signature Certificate",
		Creator:  "goSign (https://github.com/shurco/goSign)",
		Producer: "goSign (https://github.com/shurco/goSign)",
	})

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write certificate PDF: %w", err)
	}
	return buf.Bytes(), nil
}

// AppendSignatureCertificate appends the generated certificate PDF to basePDF.
func AppendSignatureCertificate(basePDF []byte, certificatePDF []byte) ([]byte, error) {
	return AppendPDF(basePDF, certificatePDF)
}

func formatTimeUTC(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "-"
	}
	return t.UTC().Format("02 Jan 2006 15:04:05 UTC")
}

func formatCertTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "-"
	}
	return t.UTC().Format("02 Jan 2006 15:04:05 UTC")
}

