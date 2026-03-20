package pdf

import (
	"testing"

	"github.com/signintech/gopdf"
)

func TestStandardFonts_SetNormalSetBold_fallback(t *testing.T) {
	var p gopdf.GoPdf
	p.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	// No TTF registered: exercise helvetica fallback branches.
	f := standardFonts{NormalName: "Arial", BoldName: "Arial-Bold"}
	f.SetNormal(&p, 10)
	f.SetBold(&p, 12)
}

func TestStandardFonts_SetNormalSetBold_withTTF(t *testing.T) {
	var p gopdf.GoPdf
	p.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	fs := addStandardFonts(&p, "")
	if !fs.NormalOK {
		t.Skip("no TTF on this system")
	}
	fs.SetNormal(&p, 10)
	fs.SetBold(&p, 11)
}
