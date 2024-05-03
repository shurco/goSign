package main

import (
	"github.com/google/uuid"
	"github.com/signintech/gopdf"
	"github.com/skip2/go-qrcode"
)

func main() {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	pdf.AddTTFFont("Arial", "./fonts/Arial.ttf")
	pdf.AddTTFFont("Arial-Bold", "./fonts/Arial-Bold.ttf")

	// ---------
	pdf.AddPage()
	cert := pdf.ImportPage("./img/cert-fon.pdf", 1, "/MediaBox")
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
	refNumber := uuid.New().String()
	pdf.Cell(nil, refNumber) // dynamic

	pdf.SetFont("Arial-Bold", "", 8)
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
	type Signer struct {
		UserName            string `json:"user_name"`
		Email               string `json:"email"`
		Sent                string `json:"sent"`
		Viewed              string `json:"viewed"`
		Signed              string `json:"signed"`
		EmailVerified       string `json:"email_verified"`
		EmailVerifiedStatus string `json:"email_verified_status"`
		IP                  string `json:"ip"`
		Location            string `json:"location"`
		SignerURL           string `json:"signer_url"`
	}

	// max 5 signed on page
	signers := []Signer{
		{
			UserName:            "User Name 1",
			Email:               "user@mail.com",
			Sent:                "02 Feb 2023 09:59:25 UTC",
			Viewed:              "02 Feb 2023 10:05:59 UTC",
			Signed:              "02 Feb 2023 10:06:21 UTC",
			EmailVerified:       "02 Feb 2023 10:05:59 UTC",
			EmailVerifiedStatus: "x",
			IP:                  "79.153.222.202",
			Location:            "Barcelona, Spain",
			SignerURL:           "",
		},
		/*
			{
				UserName:            "User Name 2",
				Email:               "user@mail.com",
				Sent:                "02 Feb 2023 09:59:25 UTC",
				Viewed:              "02 Feb 2023 10:05:59 UTC",
				Signed:              "02 Feb 2023 10:06:21 UTC",
				EmailVerified:       "02 Feb 2023 10:05:59 UTC",
				EmailVerifiedStatus: "x",
				IP:                  "79.153.222.202",
				Location:            "Barcelona, Spain",
				SignerURL:           "",
			},
			{
				UserName:            "User Name 3",
				Email:               "user@mail.com",
				Sent:                "02 Feb 2023 09:59:25 UTC",
				Viewed:              "02 Feb 2023 10:05:59 UTC",
				Signed:              "02 Feb 2023 10:06:21 UTC",
				EmailVerified:       "02 Feb 2023 10:05:59 UTC",
				EmailVerifiedStatus: "x",
				IP:                  "79.153.222.202",
				Location:            "Barcelona, Spain",
				SignerURL:           "",
			},
			{
				UserName:            "User Name 4",
				Email:               "user@mail.com",
				Sent:                "02 Feb 2023 09:59:25 UTC",
				Viewed:              "02 Feb 2023 10:05:59 UTC",
				Signed:              "02 Feb 2023 10:06:21 UTC",
				EmailVerified:       "02 Feb 2023 10:05:59 UTC",
				EmailVerifiedStatus: "x",
				IP:                  "79.153.222.202",
				Location:            "Barcelona, Spain",
				SignerURL:           "",
			},
			{
				UserName:            "User Name 5",
				Email:               "user@mail.com",
				Sent:                "02 Feb 2023 09:59:25 UTC",
				Viewed:              "02 Feb 2023 10:05:59 UTC",
				Signed:              "02 Feb 2023 10:06:21 UTC",
				EmailVerified:       "02 Feb 2023 10:05:59 UTC",
				EmailVerifiedStatus: "x",
				IP:                  "79.153.222.202",
				Location:            "Barcelona, Spain",
				SignerURL:           "",
			},
		*/
	}

	var shiftSignerBlock float64
	for _, signer := range signers {
		shiftSignerBlock = shiftSignerBlock + 98

		pdf.SetLineWidth(0.5)
		pdf.SetLineType("solid")
		pdf.SetStrokeColor(153, 153, 153)
		pdf.Line(75, 167+shiftSignerBlock, 525, 167+shiftSignerBlock)

		pdf.SetFillColor(255, 255, 255)
		pdf.SetTransparency(gopdf.Transparency{
			Alpha:         0.8,
			BlendModeType: gopdf.SoftLight,
		})
		pdf.Rectangle(75, 69+shiftSignerBlock, 525, 167+shiftSignerBlock, "F", 0, 0)
		pdf.ClearTransparency()

		pdf.SetFont("Arial-Bold", "", 10)
		pdf.SetXY(83, 77+shiftSignerBlock)
		pdf.Cell(nil, signer.UserName)
		pdf.SetFont("Arial", "", 7)
		pdf.SetXY(83, 89+shiftSignerBlock)
		pdf.Cell(nil, "Email:")
		pdf.SetXY(105, 89+shiftSignerBlock)
		pdf.Cell(nil, signer.Email)

		pdf.SetXY(83, 105+shiftSignerBlock)
		pdf.Cell(nil, "Sent:")
		pdf.SetXY(225, 105+shiftSignerBlock)
		pdf.Cell(nil, signer.Sent)

		pdf.SetXY(83, 115+shiftSignerBlock)
		pdf.Cell(nil, "Viewed:")
		pdf.SetXY(225, 115+shiftSignerBlock)
		pdf.Cell(nil, signer.Viewed)

		pdf.SetXY(83, 125+shiftSignerBlock)
		pdf.Cell(nil, "Signed:")
		pdf.SetXY(225, 125+shiftSignerBlock)
		pdf.Cell(nil, signer.Signed)

		pdf.SetFont("Arial-Bold", "", 8)
		pdf.SetXY(83, 142+shiftSignerBlock)
		pdf.Cell(nil, "Recipient Verification:")
		pdf.SetFont("Arial", "", 7)
		pdf.SetXY(90, 152+shiftSignerBlock)
		pdf.Cell(nil, "Email verified")
		pdf.SetXY(83, 152+shiftSignerBlock)
		pdf.Cell(nil, signer.EmailVerifiedStatus)
		pdf.SetXY(225, 152+shiftSignerBlock)
		pdf.Cell(nil, signer.EmailVerified)

		// signature
		pdf.Rectangle(370, 77+shiftSignerBlock, 515, 132+shiftSignerBlock, "D", 0, 0)

		pdf.SetXY(370, 142+shiftSignerBlock)
		pdf.Cell(nil, "IP address:")
		pdf.SetXY(407, 142+shiftSignerBlock)
		pdf.Cell(nil, signer.IP)
		pdf.SetXY(370, 152+shiftSignerBlock)
		pdf.Cell(nil, "Location:")
		pdf.SetXY(400, 152+shiftSignerBlock)
		pdf.Cell(nil, signer.Location)
	}

	// ---- END BLOCK ------

	pdf.SetFont("Arial", "", 8)
	pdf.SetXY(80, 167+shiftSignerBlock+15)
	pdf.SetTextColor(109, 109, 109)
	pdf.Cell(nil, "Document completed by all parties on:")
	pdf.SetXY(220, 167+shiftSignerBlock+15)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(nil, "02 Feb 2023 10:11:25 UTC") // dynamic

	// ------ START FOOTER -----
	pdf.SetFillColor(71, 170, 98)
	pdf.Rectangle(80, 700, 145, 765, "F", 0, 0)
	pdf.Image("./img/stamp.png", 85, 705, &gopdf.Rect{W: 55, H: 55})

	pdf.SetFont("Arial-Bold", "", 8)
	pdf.SetXY(160, 717)
	pdf.Cell(nil, "Signed with goSign")
	pdf.SetFont("Arial", "", 7)
	pdf.SetXY(160, 737)
	pdf.Text("goSign is an open-source solution for easy")
	pdf.SetXY(160, 747)
	pdf.Text("and secure document signing with eSignature.")

	// pdf.SetFillColor(255, 255, 255)
	// pdf.Rectangle(465, 705, 525, 765, "F", 0, 0)
	qrCode, _ := qrcode.Encode("https://github.com/shurco/goSign", qrcode.Medium, 256) // dynamic
	imgQRCode, _ := gopdf.ImageHolderByBytes(qrCode)
	pdf.ImageByHolder(imgQRCode, 460, 700, &gopdf.Rect{W: 65, H: 65})
	// ------ END FOOTER -----

	// ---------
	pdf.SetInfo(gopdf.PdfInfo{
		Title:    "Title",
		Author:   "goSign (https://github.com/shurco/goSign)",
		Subject:  "Subject",
		Creator:  "goSign (https://github.com/shurco/goSign)",
		Producer: "goSign (https://github.com/shurco/goSign)",
	})

	pdf.WritePdf("example.pdf")
}
