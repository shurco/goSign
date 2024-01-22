package main

import (
	"github.com/signintech/gopdf"
)

func main() {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddTTFFont("arial", "arial.ttf")
	pdf.SetFont("arial", "", 8)

	pdf.AddFooter(func() {
		pdf.SetY(825)
		pdf.Cell(nil, "footer")
	})

	// ----
	pdf.AddPage()
	tpl1 := pdf.ImportPage("example-pdf.pdf", 1, "/MediaBox")
	pdf.UseImportedTemplate(tpl1, 0, 0, 0, 0)

	// ----
	pdf.AddPage()

	pdf.SetLineWidth(0.1)
	pdf.SetFillColor(124, 252, 0)
	pdf.RectFromUpperLeftWithStyle(50, 100, 400, 600, "FD")
	pdf.SetFillColor(0, 0, 0)

	pdf.Image("./image.png", 100, 150, nil)

	pdf.SetFont("arial", "", 20)
	pdf.SetXY(70, 50)
	pdf.Cell(nil, "Import existing PDF into GoPDF Document")
	pdf.Close()

	pdf.AddPage()
	pdf.SetY(400)
	pdf.Text("page 1 content")
	pdf.AddPage()
	pdf.SetY(400)
	pdf.Text("page 2 content")

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
