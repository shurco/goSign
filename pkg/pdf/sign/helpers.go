package sign

import (
	"crypto"
	"encoding/asn1"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/digitorus/pdf"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func findFirstPage(parent pdf.Value) (pdf.Value, error) {
	valueType := parent.Key("Type").String()
	switch valueType {
	case "/Pages":
		kids := parent.Key("Kids")
		for i := 0; i < kids.Len(); i++ {
			page, err := findFirstPage(kids.Index(i))
			if err == nil {
				return page, nil
			}
		}
		return parent, errors.New("could not find first page")

	case "/Page":
		return parent, nil
	}

	return parent, errors.New("could not find first page")
}

func pdfString(text string) string {
	if !isASCII(text) {
		// UTF-16BE with Byte Order Mark (BOM)
		encoder := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewEncoder()
		encodedText, _, err := transform.String(encoder, text)
		if err != nil {
			panic(err) // Consider handling the error without panic if appropriate
		}
		return "(" + encodedText + ")"
	}

	// PDFDocEncoded
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		")", "\\)",
		"(", "\\(",
		"\r", "\\r",
	)
	escapedText := replacer.Replace(text)
	return "(" + escapedText + ")"
}

func pdfDateTime(date time.Time) string {
	// Calculate timezone offset from GMT.
	_, originalOffset := date.Zone()
	sign := "+"
	if originalOffset < 0 {
		sign = "-"
		originalOffset = -originalOffset // Make offset positive for formatting
	}

	// Calculate hours and minutes from offset
	offsetHours := originalOffset / 3600
	offsetMinutes := (originalOffset / 60) % 60

	// Format date and time with timezone offset
	dateString := fmt.Sprintf("D:%s%s%02d'%02d'", date.Format("20060102150405"), sign, offsetHours, offsetMinutes)

	// Escape characters for PDF string and return
	return pdfString(dateString)
}

func leftPad(s string, padStr string, pLen int) string {
	if pLen <= 0 {
		return s
	}
	return strings.Repeat(padStr, pLen) + s
}

func writePartFromSourceFileToTargetFile(inputFile io.ReadSeeker, outputFile io.Writer, offset int64, length int64) error {
	if length <= 0 {
		return nil
	}
	if _, err := inputFile.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	_, err := io.CopyN(outputFile, inputFile, length)
	return err
}

var hashOIDs = map[crypto.Hash]asn1.ObjectIdentifier{
	crypto.SHA1:   asn1.ObjectIdentifier([]int{1, 3, 14, 3, 2, 26}),
	crypto.SHA256: asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 1}),
	crypto.SHA384: asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 2}),
	crypto.SHA512: asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 3}),
}

// func getHashAlgorithmFromOID(target asn1.ObjectIdentifier) crypto.Hash {
// 	for hash, oid := range hashOIDs {
// 		if oid.Equal(target) {
// 			return hash
// 		}
// 	}
// 	return crypto.Hash(0)
// }

func getOIDFromHashAlgorithm(target crypto.Hash) asn1.ObjectIdentifier {
	if oid, found := hashOIDs[target]; found {
		return oid
	}
	return nil
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > '\u007F' {
			return false
		}
	}
	return true
}
