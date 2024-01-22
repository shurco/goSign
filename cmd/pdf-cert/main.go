package main

import (
	"github.com/signintech/gopdf"
)

func main() {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddTTFFont("Arial", "./fonts/Arial.ttf")
	pdf.AddTTFFont("Arial-Bold", "./fonts/Arial-Bold.ttf")

	// ---------
	pdf.AddPage()
	// pdf.Image("./fon/cert-fon.jpg", 0, 0, gopdf.PageSizeA4)
	cert := pdf.ImportPage("./fon/cert-fon.pdf", 1, "/MediaBox")
	pdf.UseImportedTemplate(cert, 0, 0, 0, 0)

	// ------ START HEADER -----
	pdf.SetFont("Arial-Bold", "", 20)
	pdf.SetXY(80, 80)
	pdf.Cell(nil, "Signature Certificate")

	pdf.SetFont("Arial", "", 8)
	pdf.SetXY(80, 110)
	pdf.SetTextColor(109, 109, 109)
	pdf.Cell(nil, "Reference number:")

	pdf.SetXY(150, 110)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(nil, "JAQCX-RV9YG-VYTGG-KQBIU") // dynamic

	pdf.SetFont("Arial-Bold", "", 8)
	pdf.SetXY(80, 155)
	pdf.Cell(nil, "Signer")
	pdf.SetXY(225, 155)
	pdf.Cell(nil, "Timestamp")
	pdf.SetXY(370, 155)
	pdf.Cell(nil, "Signature")
	// ------ END HEADER -----

	// ---- START BLOCK ------
	pdf.SetLineWidth(0.5)
	pdf.SetLineType("solid")
	pdf.SetStrokeColor(153, 153, 153)
	pdf.Line(75, 167, 525, 167)

	pdf.SetFillColor(255, 255, 255)
	pdf.SetTransparency(gopdf.Transparency{
		Alpha:         0.7,
		BlendModeType: gopdf.SoftLight,
	})
	pdf.Rectangle(75, 167, 525, 265, "F", 0, 0)
	pdf.ClearTransparency()

	pdf.SetFont("Arial-Bold", "", 10)
	pdf.SetXY(83, 175)
	pdf.Cell(nil, "User Name") // dynamic
	pdf.SetFont("Arial", "", 7)
	pdf.SetXY(83, 187)
	pdf.Cell(nil, "Email:")
	pdf.SetXY(105, 187)
	pdf.Cell(nil, "user@mail.com") // dynamic

	pdf.SetXY(83, 203)
	pdf.Cell(nil, "Sent:")
	pdf.SetXY(225, 203)
	pdf.Cell(nil, "02 Feb 2023 09:59:25 UTC") // dynamic

	pdf.SetXY(83, 213)
	pdf.Cell(nil, "Viewed:")
	pdf.SetXY(225, 213)
	pdf.Cell(nil, "02 Feb 2023 10:05:59 UTC") // dynamic

	pdf.SetXY(83, 223)
	pdf.Cell(nil, "Signed:")
	pdf.SetXY(225, 223)
	pdf.Cell(nil, "02 Feb 2023 10:06:21 UTC") // dynamic

	pdf.SetFont("Arial-Bold", "", 8)
	pdf.SetXY(83, 240)
	pdf.Cell(nil, "Recipient Verification:")
	pdf.SetFont("Arial", "", 7)
	pdf.SetXY(90, 250)
	pdf.Cell(nil, "Email verified")
	pdf.SetXY(83, 250)
	pdf.Cell(nil, "x") // dynamic
	pdf.SetXY(225, 250)
	pdf.Cell(nil, "02 Feb 2023 10:05:59 UTC") // dynamic

	// signature
	pdf.Rectangle(370, 175, 515, 230, "D", 0, 0)

	pdf.SetXY(370, 240)
	pdf.Cell(nil, "IP address:")
	pdf.SetXY(407, 240)
	pdf.Cell(nil, "79.153.222.202") // dynamic
	pdf.SetXY(370, 250)
	pdf.Cell(nil, "Location:")
	pdf.SetXY(400, 250)
	pdf.Cell(nil, "Barcelona, Spain") // dynamic
	// ---- END BLOCK ------

	pdf.SetFont("Arial", "", 8)
	pdf.SetXY(80, 285)
	pdf.SetTextColor(109, 109, 109)
	pdf.Cell(nil, "Document completed by all parties on:")
	pdf.SetXY(220, 285)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(nil, "02 Feb 2023 10:11:25 UTC") // dynamic
	pdf.SetXY(80, 300)
	pdf.Cell(nil, "Page 1 of 1") // dynamic

	// ------ START FOOTER -----
	pdf.SetFillColor(71, 170, 98)
	pdf.Rectangle(80, 700, 145, 765, "F", 0, 0)
	pdf.Image("./stamp.png", 85, 705, &gopdf.Rect{W: 55, H: 55})

	pdf.SetFont("Arial-Bold", "", 8)
	pdf.SetXY(160, 717)
	pdf.Cell(nil, "Signed with goSign")
	pdf.SetFont("Arial", "", 7)
	pdf.SetXY(160, 737)
	pdf.Text("goSign is an open-source solution for easy")
	pdf.SetXY(160, 747)
	pdf.Text("and secure document signing with eSignature.")

	pdf.SetFillColor(255, 255, 255)
	pdf.Rectangle(460, 700, 525, 765, "F", 0, 0)
	pdf.Image("./qr-code.png", 465, 705, &gopdf.Rect{W: 55, H: 55}) // dynamic
	// ------ END FOOTER -----

	// ---------
	pdf.SetInfo(gopdf.PdfInfo{
		Title:   "Title",
		Author:  "goSign (https://github.com/shurco/goSign)",
		Subject: "Subject",
		Creator: "goSign (https://github.com/shurco/goSign)",
		// Producer: "Producer",
	})
	pdf.WritePdf("example.pdf")
}
