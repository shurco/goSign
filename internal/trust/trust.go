package trust

import (
	"bytes"
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
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"

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
	if err := api.ExtractAttachments(bytes.NewReader(data), "./", []string{securityFile}, nil); err != nil {
		return errors.New("error extracting attachments")
	}

	fileContents, err := os.ReadFile("./" + securityFile)
	if err != nil {
		return errors.New("error opening file")
	}

	var securitySettings SecuritySettings
	err = xml.Unmarshal(fileContents, &securitySettings)
	if err != nil {
		return errors.New("error decoding XML")
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
