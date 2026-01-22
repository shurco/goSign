package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpufont "github.com/pdfcpu/pdfcpu/pkg/font"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/create"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/primitives"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/skip2/go-qrcode"
)

func main() {
	// Parse command line flags
	addWatermark := flag.Bool("wm", false, "Add watermark background from substrate.pdf")
	flag.Parse()

	// Install and load TTF fonts (if fonts directory exists in parent)
	fontDir := "../pdf-cert/fonts"
	if err := installAndLoadFonts(fontDir); err != nil {
		// Try local fonts directory
		fontDir = "./fonts"
		_ = installAndLoadFonts(fontDir)
	}

	// Get the base directory (where main.go is located)
	// When running with 'go run', the working directory should be cmd/pdf-cert2
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// Ensure we're in the right directory by checking for img folder
	if _, err := os.Stat(filepath.Join(baseDir, "img")); err != nil {
		// If img doesn't exist, try relative to current directory
		baseDir, _ = filepath.Abs(".")
	}

	UUID := uuid.New().String()
	imgDir := filepath.Join(baseDir, "img")
	fileUUID := filepath.Join(imgDir, UUID+".png")

	// Ensure img directory exists
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		log.Fatal(err)
	}

	if err := qrcode.WriteFile("https://github.com/shurco/goSign", qrcode.Medium, 128, fileUUID); err != nil {
		log.Fatal(err)
	}

	// Convert to absolute paths for PDF library
	qrcodePath, err := filepath.Abs(fileUUID)
	if err != nil {
		log.Fatal(err)
	}
	stampPath, err := filepath.Abs(filepath.Join(imgDir, "stamp.png"))
	if err != nil {
		log.Fatal(err)
	}

	rootPDF := &primitives.PDF{
		Origin:     "UpperLeft",
		Debug:      false,
		ContentBox: false,
		Guides:     false,
		ImageBoxPool: map[string]*primitives.ImageBox{
			"qrcode": {
				Src:    qrcodePath,
				Dx:     460,
				Dy:     765,
				Width:  65,
				Height: 65,
				Url:    "https://github.com/shurco/goSign",
			},
			"stamp": {
				Src:    stampPath,
				Dx:     80,
				Dy:     765,
				Width:  65,
				Height: 65,
				Url:    "https://github.com/shurco/goSign",
			},
		},
		Fonts: buildFontMap(),
		Pages: map[string]*primitives.PDFPage{
			"1": {
				Content: &primitives.Content{
					TextBoxes: []*primitives.TextBox{
						{
							Name:     "test",
							Value:    "Signature Certificate",
							Position: [2]float64{80, 100},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica-bold", "$arial-bold"), Size: 20},
						},
						{
							Value:    "Reference number:",
							Position: [2]float64{80, 118},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Color: "#6D6D6D"},
						},
						{
							Value:    "Signer",
							Position: [2]float64{80, 163},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica-bold", "$arial-bold")},
						},
						{
							Value:    "Timestamp",
							Position: [2]float64{225, 163},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica-bold", "$arial-bold")},
						},
						{
							Value:    "Signature",
							Position: [2]float64{370, 163},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica-bold", "$arial-bold")},
						},
						{
							Value:    "Signed with goSign",
							Position: [2]float64{160, 725},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica-bold", "$arial-bold")},
						},
						{
							Value:    "goSign is an open-source solution for easy",
							Position: [2]float64{160, 740},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{
							Value:    "and secure document signing with eSignature.",
							Position: [2]float64{160, 750},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},

						// dynamic content
						{
							Value:    UUID, // dynamic
							Position: [2]float64{150, 118},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial")},
						},
						{ // dynamic moves with the table
							Value:    "Document completed by all parties on:",
							Position: [2]float64{80, 287}, // '287' - dynamic
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Color: "#6D6D6D"},
						},
						{ // dynamic moves with the table
							Value:    "02 Feb 2023 09:59:25 UTC", // dynamic
							Position: [2]float64{220, 287},       // '287' - dynamic
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial")},
						},
						// end dynamic content

						// signer section 1
						{
							Value:    "Email:",
							Position: [2]float64{83, 194},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{
							Value:    "Sent:",
							Position: [2]float64{83, 210},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{
							Value:    "Viewed:",
							Position: [2]float64{83, 220},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{
							Value:    "Signed:",
							Position: [2]float64{83, 230},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{
							Value:    "Recipient Verification:",
							Position: [2]float64{83, 248},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica-bold", "$arial-bold")},
						},
						{
							Value:    "Email verified",
							Position: [2]float64{90, 258},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{
							Value:    "IP address:",
							Position: [2]float64{370, 248},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{
							Value:    "Location:",
							Position: [2]float64{370, 258},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},

						// dynamic content
						{
							Value:    "User Name 1", // dynamic
							Position: [2]float64{83, 185},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica-bold", "$arial-bold"), Size: 10},
						},
						{
							Value:    "user@mail.com", // dynamic
							Position: [2]float64{105, 194},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{ // Sent
							Value:    "02 Feb 2023 09:59:25 UTC", // dynamic
							Position: [2]float64{225, 210},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{ // Viewed
							Value:    "02 Feb 2023 09:59:25 UTC", // dynamic
							Position: [2]float64{225, 220},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{ // Signed
							Value:    "02 Feb 2023 09:59:25 UTC", // dynamic
							Position: [2]float64{225, 230},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{ // Email verified
							Value:    "02 Feb 2023 09:59:25 UTC", // dynamic
							Position: [2]float64{225, 258},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{ // IP address
							Value:    "79.153.222.202", // dynamic
							Position: [2]float64{408, 248},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
						},
						{ // Location
							Value:    "Barcelona, Spain", // dynamic
							Position: [2]float64{400, 258},
							Font:     &primitives.FormFont{Name: getFontName("$helvetica", "$arial"), Size: 7},
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

	// Save certificate to temporary file first
	tmpCertFile := "./cert_temp.pdf"
	certFile, err := os.Create(tmpCertFile)
	if err != nil {
		log.Fatal(err)
	}

	model.VersionStr = "goSign"
	if err := api.WriteContext(ctx, certFile); err != nil {
		certFile.Close()
		log.Fatal(err)
	}
	certFile.Close()

	// Final certificate file
	fileName := "cert.pdf"
	
	// Add background from substrate.pdf as watermark if -wm flag is set
	if *addWatermark {
		conf = model.NewDefaultConfiguration()
		
		// Create watermark from substrate.pdf page 1
		// PDFWatermark(fileName, description, onTop, update, unit)
		// onTop=false means background (behind content)
		// description: empty string uses first page, or use format like "page:1"
		substratePath, err := filepath.Abs(filepath.Join(imgDir, "substrate.pdf"))
		if err != nil {
			log.Printf("Warning: Failed to resolve substrate.pdf path: %v\n", err)
			os.Rename(tmpCertFile, fileName)
		} else {
			wm, err := api.PDFWatermark(substratePath, "", false, false, types.POINTS)
			if err != nil {
				log.Printf("Warning: Failed to create watermark from substrate.pdf: %v\n", err)
				// Continue without background if substrate.pdf is not available
				os.Rename(tmpCertFile, fileName)
			} else {
				// Set watermark to cover entire page as background
				wm.Scale = 1.0
				wm.ScaleAbs = true
				wm.Dx = 0
				wm.Dy = 0
				wm.Diagonal = model.NoDiagonal
				
				// Apply watermark to page 1 (background should be behind content)
				pages := []string{"1"}
				
				// Add background watermark
				err = api.AddWatermarksFile(tmpCertFile, fileName, pages, wm, conf)
				if err != nil {
					log.Printf("Warning: Failed to add background watermark: %v\n", err)
					// Use file without background
					os.Rename(tmpCertFile, fileName)
				} else {
					// Remove temporary file
					os.Remove(tmpCertFile)
				}
			}
		}
	} else {
		// No watermark flag - use white background (just rename temp file)
		os.Rename(tmpCertFile, fileName)
	}

	// remove tmp qrcode file
	if err := os.Remove(fileUUID); err != nil {
		log.Fatal(err)
	}

}

// installAndLoadFonts installs and loads TTF fonts from the fonts directory
func installAndLoadFonts(fontDir string) error {
	// Convert to absolute path
	absFontDir, err := filepath.Abs(fontDir)
	if err != nil {
		return err
	}

	// Set user font directory
	pdfcpufont.UserFontDir = absFontDir

	// Try to install Arial fonts if they exist
	arialPath := filepath.Join(absFontDir, "Arial.ttf")
	arialBoldPath := filepath.Join(absFontDir, "Arial-Bold.ttf")

	if _, err := os.Stat(arialPath); err == nil {
		// Read font file and install from bytes
		if fontData, err := os.ReadFile(arialPath); err == nil {
			_ = pdfcpufont.InstallFontFromBytes(absFontDir, "Arial.ttf", fontData)
		}
	}

	if _, err := os.Stat(arialBoldPath); err == nil {
		// Read font file and install from bytes
		if fontData, err := os.ReadFile(arialBoldPath); err == nil {
			_ = pdfcpufont.InstallFontFromBytes(absFontDir, "Arial-Bold.ttf", fontData)
		}
	}

	// Load user fonts
	_ = pdfcpufont.LoadUserFonts()

	return nil
}

// buildFontMap builds font map with TTF fonts if available, otherwise uses default fonts
func buildFontMap() map[string]*primitives.FormFont {
	fontMap := make(map[string]*primitives.FormFont)

	// Check if Arial fonts are available (after loading)
	// pdfcpu loads TTF fonts with names like "ArialMT" and "Arial-BoldMT"
	arialName := "ArialMT"
	arialBoldName := "Arial-BoldMT"
	
	// Always add standard Helvetica fonts as fallback
	fontMap["helvetica"] = &primitives.FormFont{Name: "Helvetica", Size: 8}
	fontMap["helvetica-bold"] = &primitives.FormFont{Name: "Helvetica-Bold", Size: 8}

	if pdfcpufont.SupportedFont(arialName) && pdfcpufont.SupportedFont(arialBoldName) {
		// Use TTF Arial fonts
		fontMap["arial"] = &primitives.FormFont{Name: arialName, Size: 8}
		fontMap["arial-bold"] = &primitives.FormFont{Name: arialBoldName, Size: 8}
	}

	return fontMap
}

// getFontName returns TTF font name if available, otherwise returns default font name
// pdfcpu loads TTF fonts with names like "ArialMT" and "Arial-BoldMT"
func getFontName(defaultFont, ttfFont string) string {
	// Check if TTF fonts are available (with correct names)
	arialName := "ArialMT"
	arialBoldName := "Arial-BoldMT"
	
	if pdfcpufont.SupportedFont(arialName) || pdfcpufont.SupportedFont(arialBoldName) {
		// Map to correct TTF font names
		if ttfFont == "$arial" {
			return "$arial"
		} else if ttfFont == "$arial-bold" {
			return "$arial-bold"
		}
		return ttfFont
	}
	// Use default standard PDF font
	return defaultFont
}
