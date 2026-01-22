package main

import (
	"log"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/create"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/primitives"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func main() {
	// Create PDF with multiple pages using pdfcpu primitives API
	rootPDF := &primitives.PDF{
		Origin:     "UpperLeft",
		Debug:      false,
		ContentBox: false,
		Guides:     false,
		Fonts: map[string]*primitives.FormFont{
			"helvetica": {Name: "Helvetica", Size: 8},
		},
		Pages: map[string]*primitives.PDFPage{
			// Page 1: Import existing PDF page (if example.pdf exists)
			// TODO: For importing existing PDF pages, use api.MergeCreateFile instead
			// This example creates a new page with text indicating import functionality
			"1": {
				Content: &primitives.Content{
					TextBoxes: []*primitives.TextBox{
						{
							Value:    "Import existing PDF into GoPDF Document",
							Position: [2]float64{70, 50},
							Font:     &primitives.FormFont{Name: "Helvetica", Size: 20},
						},
						{
							Value:    "Note: To import existing PDF pages, use api.MergeCreateFile",
							Position: [2]float64{70, 80},
							Font:     &primitives.FormFont{Name: "Helvetica", Size: 10},
						},
					},
				},
			},
			// Page 2: Rectangle and image
			"2": {
				Content: &primitives.Content{
					SimpleBoxes: []*primitives.SimpleBox{
						{
							Dx:        50,
							Dy:        100,
							Width:     400,
							Height:    600,
							FillColor: "#7CFC00", // lime green (124, 252, 0)
							Border: &primitives.Border{
								Width: 1,
								Color: "#000000",
							},
						},
					},
					ImageBoxes: []*primitives.ImageBox{
						{
							Src:    "./image.png",
							Dx:     100,
							Dy:     150,
							Width:  200,
							Height: 200,
						},
					},
					TextBoxes: []*primitives.TextBox{
						{
							Value:    "Page with rectangle and image",
							Position: [2]float64{70, 50},
							Font:     &primitives.FormFont{Name: "Helvetica", Size: 12},
						},
					},
				},
			},
			// Page 3: Text content
			"3": {
				Content: &primitives.Content{
					TextBoxes: []*primitives.TextBox{
						{
							Value:    "page 1 content",
							Position: [2]float64{50, 400},
							Font:     &primitives.FormFont{Name: "Helvetica", Size: 10},
						},
					},
				},
			},
			// Page 4: More text content
			"4": {
				Content: &primitives.Content{
					TextBoxes: []*primitives.TextBox{
						{
							Value:    "page 2 content",
							Position: [2]float64{50, 400},
							Font:     &primitives.FormFont{Name: "Helvetica", Size: 10},
						},
					},
				},
			},
		},
	}

	conf := model.NewDefaultConfiguration()
	conf.Cmd = model.CREATE

	ctx, err := pdfcpu.CreateContextWithXRefTable(conf, types.PaperSize["A4"])
	if err != nil {
		log.Fatal(err)
	}

	rootPDF.Conf = ctx.Configuration
	rootPDF.XRefTable = ctx.XRefTable
	rootPDF.Optimize = ctx.Optimize

	if err := rootPDF.Validate(); err != nil {
		log.Fatal(err)
	}

	pages, fontMap, err := rootPDF.RenderPages()
	if err != nil {
		log.Fatal(err)
	}

	if _, _, err := create.UpdatePageTree(ctx, pages, fontMap); err != nil {
		log.Fatal(err)
	}

	if err = api.ValidateContext(ctx); err != nil {
		log.Fatal(err)
	}

	// Set PDF metadata
	ctx.Author = "goSign (https://github.com/shurco/goSign)"
	ctx.Title = "Title"
	ctx.Subject = "Subject"
	ctx.Creator = "goSign (https://github.com/shurco/goSign)"
	ctx.Producer = "goSign (https://github.com/shurco/goSign)"

	// Save PDF
	fileName := "new.pdf"
	pdfFile, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer pdfFile.Close()

	model.VersionStr = "goSign"
	if err := api.WriteContext(ctx, pdfFile); err != nil {
		log.Fatal(err)
	}

	log.Printf("PDF created successfully: %s\n", fileName)
}
