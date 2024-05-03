package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/primitives"
	"github.com/skip2/go-qrcode"
)

func main() {
	UUID := uuid.New().String()
	fileUUID := "./img/" + UUID + ".png"

	if err := qrcode.WriteFile("https://github.com/shurco/goSign", qrcode.Medium, 128, fileUUID); err != nil {
		log.Fatal(err)
	}

	model.VersionStr = "goSign"

	certData := primitives.PDF{
		Origin:     "UpperLeft",
		Debug:      false,
		ContentBox: false,
		Guides:     false,
		ImageBoxPool: map[string]*primitives.ImageBox{
			"qrcode": {
				Src:    "./img/" + UUID + ".png",
				Dx:     460,
				Dy:     765,
				Width:  65,
				Height: 65,
				Url:    "https://github.com/shurco/goSign",
			},
			"stamp": {
				Src:    "./img/stamp.png",
				Dx:     80,
				Dy:     765,
				Width:  65,
				Height: 65,
				Url:    "https://github.com/shurco/goSign",
			},
		},
		Fonts: map[string]*primitives.FormFont{
			"helvetica":      {Name: "Helvetica", Size: 8},
			"helvetica-bold": {Name: "Helvetica-Bold", Size: 8},
		},
		Pages: map[string]*primitives.PDFPage{
			"1": {
				Content: &primitives.Content{
					// BackgroundColor: "#C1C1C1",
					TextBoxes: []*primitives.TextBox{
						{
							Value:    "Signature Certificate",
							Position: [2]float64{80, 100},
							Font:     &primitives.FormFont{Name: "$helvetica-bold", Size: 20},
						},
						{
							Value:    "Reference number:",
							Position: [2]float64{80, 118},
							Font:     &primitives.FormFont{Name: "$helvetica", Color: "#6D6D6D"},
						},
						{
							Value:    "Signer",
							Position: [2]float64{80, 163},
							Font:     &primitives.FormFont{Name: "$helvetica-bold"},
						},
						{
							Value:    "Timestamp",
							Position: [2]float64{225, 163},
							Font:     &primitives.FormFont{Name: "$helvetica-bold"},
						},
						{
							Value:    "Signature",
							Position: [2]float64{370, 163},
							Font:     &primitives.FormFont{Name: "$helvetica-bold"},
						},
						{
							Value:    "Signed with goSign",
							Position: [2]float64{160, 725},
							Font:     &primitives.FormFont{Name: "$helvetica-bold"},
						},
						{
							Value:    "goSign is an open-source solution for easy",
							Position: [2]float64{160, 740},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{
							Value:    "and secure document signing with eSignature.",
							Position: [2]float64{160, 750},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},

						// dynamic content
						{
							Value:    UUID, // dynamic
							Position: [2]float64{150, 118},
							Font:     &primitives.FormFont{Name: "$helvetica"},
						},
						{ // dynamic moves with the table
							Value:    "Document completed by all parties on:",
							Position: [2]float64{80, 287}, // '287' - dynamic
							Font:     &primitives.FormFont{Name: "$helvetica", Color: "#6D6D6D"},
						},
						{ // dynamic moves with the table
							Value:    "02 Feb 2023 09:59:25 UTC", // dynamic
							Position: [2]float64{220, 287},       // '287' - dynamic
							Font:     &primitives.FormFont{Name: "$helvetica"},
						},
						// end dynamic content

						// signer section 1
						{
							Value:    "Email:",
							Position: [2]float64{83, 194},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{
							Value:    "Sent:",
							Position: [2]float64{83, 210},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{
							Value:    "Viewed:",
							Position: [2]float64{83, 220},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{
							Value:    "Signed:",
							Position: [2]float64{83, 230},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{
							Value:    "Recipient Verification:",
							Position: [2]float64{83, 248},
							Font:     &primitives.FormFont{Name: "$helvetica-bold"},
						},
						{
							Value:    "Email verified",
							Position: [2]float64{90, 258},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{
							Value:    "IP address:",
							Position: [2]float64{370, 248},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{
							Value:    "Location:",
							Position: [2]float64{370, 258},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},

						// dynamic content
						{
							Value:    "User Name 1", // dynamic
							Position: [2]float64{83, 185},
							Font:     &primitives.FormFont{Name: "$helvetica-bold", Size: 10},
						},
						{
							Value:    "user@mail.com", // dynamic
							Position: [2]float64{105, 194},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{ // Sent
							Value:    "02 Feb 2023 09:59:25 UTC", // dynamic
							Position: [2]float64{225, 210},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{ // Viewed
							Value:    "02 Feb 2023 09:59:25 UTC", // dynamic
							Position: [2]float64{225, 220},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{ // Signed
							Value:    "02 Feb 2023 09:59:25 UTC", // dynamic
							Position: [2]float64{225, 230},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{ // Email verified
							Value:    "02 Feb 2023 09:59:25 UTC", // dynamic
							Position: [2]float64{225, 258},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{ // IP address
							Value:    "79.153.222.202", // dynamic
							Position: [2]float64{408, 248},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						{ // Location
							Value:    "Barcelona, Spain", // dynamic
							Position: [2]float64{400, 258},
							Font:     &primitives.FormFont{Name: "$helvetica", Size: 7},
						},
						// end dynamic content

						// end signer section 1
					},

					SimpleBoxes: []*primitives.SimpleBox{
						// signer section 1
						{Dx: 75, Dy: 265, Width: 450, Height: 98, FillColor: "#F9F9F9"},
						{Dx: 370, Dy: 230, Width: 145, Height: 55, FillColor: "#FFFFFF", Border: &primitives.Border{Color: "#D0D0D0"}},
						{Dx: 75, Dy: 265, Width: 450, FillColor: "#D0D0D0"},
						// end signer section 1

						{Dx: 75, Dy: 167, Width: 450, FillColor: "#D0D0D0"}, // start table line
					},

					ImageBoxes: []*primitives.ImageBox{
						{Name: "$qrcode"},
						{Name: "$stamp"},
					},
				},
			},
		},
		XRefTable: &model.XRefTable{
			Author:   "goSign (https://github.com/shurco/goSign)",
			Creator:  "goSign (https://github.com/shurco/goSign)",
			Producer: "goSign",
		},
	}

	certBytes, err := json.Marshal(certData)
	if err != nil {
		log.Fatal(err)
	}

	substrate, err := os.ReadFile("./img/substrate.pdf")
	if err != nil {
		log.Fatal(err)
	}

	fileName := "cert.pdf"
	certFile, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer certFile.Close()

	if err = api.Create(bytes.NewReader(substrate), bytes.NewReader(certBytes), certFile, nil); err != nil {
		log.Fatal(err)
	}

	if err := os.Remove(fileUUID); err != nil {
		log.Fatal(err)
	}
}
