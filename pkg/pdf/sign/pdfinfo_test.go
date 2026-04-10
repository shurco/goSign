package sign

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/digitorus/pdf"
)

func TestCreateInfoEmpty(t *testing.T) {
	input_file, err := os.Open(testPDFFixturePath(t, "testfile20.pdf"))
	if err != nil {
		t.Errorf("Failed to load test PDF")
		return
	}

	finfo, err := input_file.Stat()
	if err != nil {
		t.Errorf("Failed to load test PDF")
		return
	}
	size := finfo.Size()

	rdr, err := pdf.NewReader(input_file, size)
	if err != nil {
		t.Errorf("Failed to load test PDF")
		return
	}

	sign_data := SignData{
		Signature: SignDataSignature{
			Info: SignDataSignatureInfo{
				Name:        "John Doe",
				Location:    "Somewhere",
				Reason:      "Test",
				ContactInfo: "None",
				Date:        time.Now().Local(),
			},
			CertType:   CertificationSignature,
			DocMDPPerm: AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
	}

	sign_data.ObjectId = uint32(rdr.XrefInformation.ItemCount) + 3

	context := SignContext{
		Filesize:  size + 1,
		PDFReader: rdr,
		InputFile: input_file,
		VisualSignData: VisualSignData{
			ObjectId: uint32(rdr.XrefInformation.ItemCount),
		},
		CatalogData: CatalogData{
			ObjectId: uint32(rdr.XrefInformation.ItemCount) + 1,
		},
		InfoData: InfoData{
			ObjectId: uint32(rdr.XrefInformation.ItemCount) + 2,
		},
		SignData: sign_data,
	}

	info, err := context.createInfo()
	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}

	expectedInfo := fmt.Sprintf("%d 0 obj\n<<>>\nendobj\n", context.InfoData.ObjectId)
	if info != expectedInfo {
		t.Errorf("Info mismatch, expected %s, but got %s", expectedInfo, info)
	}
}

func TestCreateInfo(t *testing.T) {
	input_file, err := os.Open(testPDFFixturePath(t, "testfile20.pdf"))
	if err != nil {
		t.Errorf("Failed to load test PDF")
		return
	}
	defer input_file.Close()

	finfo, err := input_file.Stat()
	if err != nil {
		t.Errorf("Failed to load test PDF")
		return
	}
	size := finfo.Size()

	rdr, err := pdf.NewReader(input_file, size)
	if err != nil {
		t.Errorf("Failed to load test PDF")
		return
	}

	sign_data := SignData{
		Signature: SignDataSignature{
			Info: SignDataSignatureInfo{
				Name:        "John Doe",
				Location:    "Somewhere",
				Reason:      "Test",
				ContactInfo: "None",
				Date:        time.Now().Local(),
			},
			CertType:   CertificationSignature,
			DocMDPPerm: AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
	}

	sign_data.ObjectId = uint32(rdr.XrefInformation.ItemCount) + 3

	context := SignContext{
		Filesize:  size + 1,
		PDFReader: rdr,
		InputFile: input_file,
		VisualSignData: VisualSignData{
			ObjectId: uint32(rdr.XrefInformation.ItemCount),
		},
		CatalogData: CatalogData{
			ObjectId: uint32(rdr.XrefInformation.ItemCount) + 1,
		},
		InfoData: InfoData{
			ObjectId: uint32(rdr.XrefInformation.ItemCount) + 2,
		},
		SignData: sign_data,
	}

	info, err := context.createInfo()
	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}

	if !strings.HasPrefix(info, fmt.Sprintf("%d 0 obj\n<<", context.InfoData.ObjectId)) {
		t.Errorf("unexpected info header: %q", info)
	}
	if !strings.HasSuffix(info, ">>\nendobj\n") {
		t.Errorf("unexpected info trailer: %q", info)
	}

	orig := rdr.Trailer().Key("Info")
	for _, key := range orig.Keys() {
		var want string
		if key == "ModDate" {
			want = pdfDateTime(sign_data.Signature.Info.Date)
		} else {
			want = pdfString(orig.Key(key).RawString())
		}
		needle := "/" + key + want
		if !strings.Contains(info, needle) {
			t.Errorf("info missing %q (full info: %q)", needle, info)
		}
	}
}
