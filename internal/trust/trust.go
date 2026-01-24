package trust

import (
	"bytes"
	"compress/zlib"
	"context"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/digitorus/pdf"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/utils"
)

// TrustList is ...
var TrustList = map[string]TrustEntry{
	"eutl12": {
		Name: "Europe",
		Url:  "https://trustlist.adobe.com/eutl12.acrobatsecuritysettings",
	},
	"tl12": {
		Name: "Adobe",
		Url:  "https://trustlist.adobe.com/tl12.acrobatsecuritysettings",
	},
}

// TrustEntry ...
type TrustEntry struct {
	Name string
	Url  string
}

// Config is ...
type Config struct {
	List   []string `toml:"list"`
	Update int      `toml:"update-frequency"`
}

// SecuritySettings is ...
type SecuritySettings struct {
	TrustedIdentities struct {
		Identity []struct {
			ImportAction int    `xml:"ImportAction"`
			Certificate  string `xml:"Certificate"`
		} `xml:"Identity"`
	} `xml:"TrustedIdentities"`
}

// Update is ...
func Update(cfg Config) error {
	dateUp, err := queries.DB.TimeUpdateAdobeTL(context.Background())
	if err != nil {
		return err
	}

	daysDiff := 999
	if dateUp != nil {
		daysDiff = utils.DaysBetween(*dateUp, time.Now())
	}

	if daysDiff < cfg.Update {
		return nil
	}

	fmt.Printf("├─[🌍] Updating the list of trusted Adobe certificates\n")
	for _, v := range cfg.List {
		trust, found := TrustList[v]
		if !found {
			return errors.New("not found trust list")
		}
		err := parseTL(v, trust.Url)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseTL(list, url string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return errors.New("error fetching security settings")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("received non-OK HTTP status code")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	securityFile := "SecuritySettings.xml"
	
	// Extract embedded file using digitorus/pdf
	fileContents, err := extractEmbeddedFile(bytes.NewReader(data), int64(len(data)), securityFile)
	if err != nil {
		return fmt.Errorf("error extracting attachments: %w", err)
	}

	// Validate that extracted data looks like XML
	if len(fileContents) == 0 {
		return errors.New("extracted file is empty")
	}
	
	// Check if data starts with XML declaration or root element
	if !bytes.HasPrefix(fileContents, []byte("<?xml")) && !bytes.HasPrefix(fileContents, []byte("<")) {
		return fmt.Errorf("extracted file does not appear to be XML (starts with: %q)", string(fileContents[:min(50, len(fileContents))]))
	}

	// Save to file
	if err := os.WriteFile("./"+securityFile, fileContents, 0644); err != nil {
		return errors.New("error writing file")
	}

	var securitySettings SecuritySettings
	err = xml.Unmarshal(fileContents, &securitySettings)
	if err != nil {
		return fmt.Errorf("error decoding XML: %w", err)
	}

	trustList := models.TrustCerts{}
	for _, identity := range securitySettings.TrustedIdentities.Identity {
		decodedData, err := base64.StdEncoding.DecodeString(identity.Certificate)
		if err != nil {
			return errors.New("error decoding certificate")
		}

		cert, err := x509.ParseCertificate(decodedData)
		if err != nil {
			//if err.Error() == "x509: unsupported elliptic curve" {
			//	fmt.Printf("elliptic\n")
			//	continue
			//}
			//fmt.Printf("error parsing certificate: %s\n", err)
			//log.Warn().Err(err).Send()
			continue
		}

		akiHash := sha1.Sum(decodedData)
		trustCert := models.TrustCert{
			List: list,
			Name: cert.Subject.CommonName,
			AKI:  strings.ToUpper(hex.EncodeToString(akiHash[:])),
			SKI:  strings.ToUpper(hex.EncodeToString(cert.SubjectKeyId[:])),
		}

		trustList.Certs = append(trustList.Certs, trustCert)
	}

	if err := queries.DB.DeleteAdobeTL(context.Background(), list); err != nil {
		return err
	}

	if err := queries.DB.AddAdobeTL(context.Background(), trustList); err != nil {
		return err
	}

	if err := os.Remove("./" + securityFile); err != nil {
		return errors.New("error removing file")
	}
	return nil
}

// extractEmbeddedFile extracts an embedded file from PDF using digitorus/pdf
func extractEmbeddedFile(reader io.ReaderAt, size int64, filename string) ([]byte, error) {
	// Create PDF reader
	pdfReader, err := pdf.NewReader(reader, size)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF reader: %w", err)
	}

	// Access Names dictionary: Root -> Names -> EmbeddedFiles -> Names
	root := pdfReader.Trailer().Key("Root")
	if root.IsNull() {
		return nil, errors.New("PDF has no root catalog")
	}

	names := root.Key("Names")
	if names.IsNull() {
		return nil, errors.New("PDF has no Names dictionary")
	}

	embeddedFiles := names.Key("EmbeddedFiles")
	if embeddedFiles.IsNull() {
		return nil, errors.New("PDF has no EmbeddedFiles")
	}

	// Get Names array (alternates between filename strings and file spec objects)
	namesArray := embeddedFiles.Key("Names")
	if namesArray.IsNull() || namesArray.Len() == 0 {
		return nil, errors.New("EmbeddedFiles has no Names array")
	}

	// Iterate through Names array (pairs of filename and file spec)
	for i := 0; i < namesArray.Len(); i += 2 {
		// Get filename (string entry)
		if i >= namesArray.Len() {
			break
		}
		fileNameValue := namesArray.Index(i)
		fileName := fileNameValue.Text()
		
		// Check if this is the file we're looking for
		if fileName != filename {
			continue
		}

		// Get file specification (next entry)
		if i+1 >= namesArray.Len() {
			return nil, errors.New("file specification missing for " + filename)
		}
		fileSpecValue := namesArray.Index(i + 1)

		// Use file spec directly (digitorus/pdf handles references automatically)
		fileSpec := fileSpecValue

		// Access EF (Embedded File) dictionary
		ef := fileSpec.Key("EF")
		if ef.IsNull() {
			return nil, errors.New("file spec has no EF dictionary")
		}

		// Access F (File) stream reference
		f := ef.Key("F")
		if f.IsNull() {
			return nil, errors.New("EF has no F stream")
		}

		// digitorus/pdf should automatically resolve references
		// Try to use f directly as the stream object
		fileStream := f
		
		// Extract object number from F reference for file reading fallback
		// F reference format is typically "X Y R" where X is object number
		fStr := f.String()
		var targetObjNum int64 = -1
		if strings.Contains(fStr, "R") {
			parts := strings.Fields(fStr)
			if len(parts) >= 1 {
				if objNum, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					targetObjNum = objNum
				}
			}
		}
		
		// If f doesn't have Length (not resolved), try to get object directly
		if fileStream.Key("Length").IsNull() && targetObjNum >= 0 {
			// Try to find the stream object by object number in xref
			for _, x := range pdfReader.Xref() {
				ptr := x.Ptr()
				objID := ptr.GetID()
				objIDStr := fmt.Sprintf("%v", objID)
				
				// Match by object number
				if strings.Contains(objIDStr, fmt.Sprintf("%d", targetObjNum)) {
					obj, err := pdfReader.GetObject(objID)
					if err == nil {
						// Check if this object is a stream (has Length field)
						if !obj.Key("Length").IsNull() {
							fileStream = obj
							break
						}
					}
				}
			}
		}

		// Get stream Length
		lengthVal := fileStream.Key("Length")
		if lengthVal.IsNull() {
			// Try to get Length from the object itself if it's a direct value
			lengthVal = f.Key("Length")
		}
		
		// Try to get stream data using RawString()
		// In digitorus/pdf, stream data might be accessible this way
		rawData := fileStream.RawString()
		
		// If RawString() returns empty, try to read stream data from file
		// Stream data in PDF is located after the stream dictionary in the file
		if len(rawData) == 0 {
			// Get stream length
			streamLength := int64(0)
			if !lengthVal.IsNull() {
				streamLength = lengthVal.Int64()
			}
			
			if streamLength > 0 {
				// Try to read stream data from file using ReaderAt
				// Stream data follows "stream\n" marker after the dictionary
				if readerAt, ok := reader.(io.ReaderAt); ok {
					// Read entire file to search for stream
					fileContent := make([]byte, size)
					if n, err := readerAt.ReadAt(fileContent, 0); err == nil && n > 0 {
						// Try to find stream by searching for object number pattern near stream markers
						// Format: "X Y obj\n<<...>>\nstream\n"
						if targetObjNum >= 0 {
							// Search for object declaration with our target number
							// Try different formats: "X 0 obj", "X Y obj"
							objPatterns := [][]byte{
								[]byte(fmt.Sprintf("%d 0 obj", targetObjNum)),
								[]byte(fmt.Sprintf("%d 0 obj\n", targetObjNum)),
								[]byte(fmt.Sprintf("%d 0 obj\r\n", targetObjNum)),
							}
							
							var objIdx int = -1
							for _, pattern := range objPatterns {
								idx := bytes.LastIndex(fileContent[:n], pattern)
								if idx >= 0 {
									objIdx = idx
									break
								}
							}
							
							if objIdx >= 0 {
								// Found object declaration, search for "stream" after it
								// Look for stream marker within reasonable distance (up to 4KB)
								searchStart := objIdx
								searchEnd := searchStart + 4096
								if searchEnd > n {
									searchEnd = n
								}
								
								// Try different stream marker formats
								streamMarkers := [][]byte{
									[]byte("\nstream\n"),
									[]byte("\r\nstream\r\n"),
									[]byte("stream\n"),
									[]byte("stream\r\n"),
									[]byte("\nstream\r\n"),
								}
								
								for _, streamMarker := range streamMarkers {
									streamIdx := bytes.Index(fileContent[searchStart:searchEnd], streamMarker)
									if streamIdx >= 0 {
										streamIdx += searchStart
										dataStart := streamIdx + len(streamMarker)
										dataEnd := dataStart + int(streamLength)
										
										// Verify that "endstream" comes after our data
										if dataEnd <= n {
											// Check for endstream marker after data
											endStreamMarker := []byte("endstream")
											endIdx := bytes.Index(fileContent[dataEnd:min(dataEnd+20, n)], endStreamMarker)
											if endIdx >= 0 || dataEnd == n {
												// Valid stream found
												rawData = string(fileContent[dataStart:dataEnd])
												break
											}
										}
									}
								}
							}
						}
						
						// If not found by object number, try searching all stream markers
						// Try different stream marker variations
						streamMarkers := [][]byte{
							[]byte("\nstream\n"),
							[]byte("\r\nstream\r\n"),
							[]byte("stream\n"),
							[]byte("stream\r\n"),
							[]byte("\nstream\r\n"),
						}
						
						var allIndices []int
						var usedMarker []byte
						
						for _, marker := range streamMarkers {
							indices := findAllIndices(fileContent[:n], marker)
							if len(indices) > 0 {
								allIndices = indices
								usedMarker = marker
								break
							}
						}
						
						if len(allIndices) > 0 {
							// Try each stream marker position, starting from the end
							// Embedded files are usually near the end of the PDF
							for i := len(allIndices) - 1; i >= 0; i-- {
								streamIdx := allIndices[i]
								dataStart := streamIdx + len(usedMarker)
								dataEnd := dataStart + int(streamLength)
								
								// Check bounds
								if dataEnd > n {
									continue
								}
								
								candidateData := fileContent[dataStart:dataEnd]
								
								// Must match exact length (or close to it, accounting for potential whitespace)
								if len(candidateData) < int(streamLength)-10 || len(candidateData) > int(streamLength)+10 {
									continue
								}
								
								// Must not be all zeros
								if isAllZeros(candidateData) {
									continue
								}
								
								// For SecuritySettings.xml, it should be XML content or compressed data
								// Try this candidate - it matches length and has content
								rawData = string(candidateData)
								break
							}
							
							// If still not found with exact length, try the last stream
							if len(rawData) == 0 && len(allIndices) > 0 {
								streamIdx := allIndices[len(allIndices)-1]
								dataStart := streamIdx + len(usedMarker)
								if dataStart < n {
									dataEnd := dataStart + int(streamLength)
									if dataEnd > n {
										dataEnd = n
									}
									candidateData := fileContent[dataStart:dataEnd]
									if len(candidateData) > 0 && !isAllZeros(candidateData) {
										rawData = string(candidateData)
									}
								}
							}
						}
					}
				}
			}
			
			// If still no data, return error
			if len(rawData) == 0 {
				if !lengthVal.IsNull() {
					return nil, fmt.Errorf("stream object found (Length: %s) but could not read stream data from file", lengthVal.String())
				}
				return nil, errors.New("file stream has no accessible data")
			}
		}

		// Check for filters (compression)
		filter := fileStream.Key("Filter")
		needsDecompression := false
		if !filter.IsNull() {
			// Filter can be a name or an array
			// Check if it's an array by checking Len() > 0 and trying Index(0)
			if filter.Len() > 0 {
				// If it's an array, check first element
				filterName := filter.Index(0).Name()
				if filterName == "/FlateDecode" {
					needsDecompression = true
				}
			} else {
				// Try as a single name
				filterName := filter.Name()
				if filterName == "/FlateDecode" {
					needsDecompression = true
				}
			}
		}

		// Convert to bytes
		fileBytes := []byte(rawData)
		
		// Handle hex-encoded strings (PDF format: <hexdata>)
		// Only if it looks like hex (starts and ends with < > and contains only hex chars)
		if len(fileBytes) >= 2 && fileBytes[0] == '<' && fileBytes[len(fileBytes)-1] == '>' {
			// Check if it's valid hex (not XML)
			hexStr := string(fileBytes[1 : len(fileBytes)-1])
			// Simple check: if it contains only hex characters and spaces/newlines
			isHex := true
			for _, r := range hexStr {
				if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || r == ' ' || r == '\n' || r == '\r' || r == '\t') {
					isHex = false
					break
				}
			}
			if isHex {
				// Remove whitespace and decode hex
				hexStr = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(hexStr, " ", ""), "\n", ""), "\r", "")
				decoded, err := hex.DecodeString(hexStr)
				if err == nil {
					fileBytes = decoded
				}
			}
		}

		// Handle FlateDecode decompression if needed
		if needsDecompression && len(fileBytes) > 0 {
			// Decompress using zlib
			zr, err := zlib.NewReader(bytes.NewReader(fileBytes))
			if err != nil {
				return nil, fmt.Errorf("failed to create zlib reader: %w", err)
			}
			defer zr.Close()
			
			decompressed, err := io.ReadAll(zr)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress stream: %w", err)
			}
			fileBytes = decompressed
		} else if !needsDecompression && len(fileBytes) > 0 {
			// If no filter specified but data doesn't look like XML, try decompression as fallback
			// SecuritySettings.xml should start with <?xml
			if !bytes.HasPrefix(fileBytes, []byte("<?xml")) && !bytes.HasPrefix(fileBytes, []byte("<")) {
				// Try decompression as fallback - maybe filter wasn't detected
				zr, err := zlib.NewReader(bytes.NewReader(fileBytes))
				if err == nil {
					defer zr.Close()
					decompressed, err := io.ReadAll(zr)
					if err == nil && len(decompressed) > 0 {
						// If decompressed data looks like XML, use it
						if bytes.HasPrefix(decompressed, []byte("<?xml")) || bytes.HasPrefix(decompressed, []byte("<")) {
							fileBytes = decompressed
						}
					}
				}
			}
		}

		if len(fileBytes) == 0 {
			return nil, errors.New("extracted file data is empty")
		}

		return fileBytes, nil
	}

	return nil, errors.New("file " + filename + " not found in embedded files")
}

// findAllIndices finds all occurrences of a byte slice in another byte slice
func findAllIndices(data, pattern []byte) []int {
	var indices []int
	idx := 0
	for {
		pos := bytes.Index(data[idx:], pattern)
		if pos == -1 {
			break
		}
		indices = append(indices, idx+pos)
		idx += pos + 1
	}
	return indices
}

// isAllZeros checks if a byte slice contains only zero bytes
func isAllZeros(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
