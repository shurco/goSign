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

func writePartFromSourceFileToTargetFile(input_file io.ReadSeeker, output_file io.Writer, offset int64, length int64) error {
	_, err := input_file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Create a small buffer for proper IO handling.
	max_chunk_length := int64(1024)

	// If the target length is smaller than our chunk size, use that as chunk size.
	if length < max_chunk_length {
		max_chunk_length = length
	}

	// Track read/written bytes so we know when we're done.
	read_bytes := int64(0)

	if length <= 0 {
		return nil
	}

	// Create a buffer for the chunks.
	buf := make([]byte, max_chunk_length)
	for {
		// Read the chunk from the input file.
		n, err := input_file.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		// If we got to the end of the file, break.
		if err == io.EOF {
			break
		}

		// If nothing was read, break.
		if n == 0 {
			break
		}

		// Write the chunk to the output file.
		if _, err := output_file.Write(buf[:n]); err != nil {
			return err
		}

		read_bytes += int64(n)

		// If we read enough bytes, break.
		if read_bytes >= length {
			break
		}

		// If our next chunk will be too big, make a smaller buffer.
		// If we won't do this, we might end up with more data than we want.
		if length-read_bytes < max_chunk_length {
			buf = make([]byte, length-read_bytes)
		}
	}

	return nil
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
