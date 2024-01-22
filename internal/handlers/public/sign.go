package handlers

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/pdf/sign"
	"github.com/shurco/gosign/pkg/utils/fsutil"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// SignPDF is ...
func SignPDF(c *fiber.Ctx) error {
	response := &models.Sign{}

	fileHeader, err := c.FormFile("document")
	if err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}

	if fileHeader.Header["Content-Type"][0] != "application/pdf" {
		response.Error = "File format not supported"
		return webutil.Response(c, fiber.StatusOK, "Verify", response)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}
	defer file.Close()

	fileNameUUID := uuid.New().String()
	fileNameSignedUUID := uuid.New().String()
	fileExt := fsutil.ExtName(fileHeader.Filename)
	response.FileName = fmt.Sprintf("%s.%s", fileNameUUID, fileExt)
	response.FileNameSigned = fmt.Sprintf("%s.%s", fileNameSignedUUID, fileExt)
	filePath := fmt.Sprintf("./lc_uploads/%s", response.FileName)

	if err := c.SaveFile(fileHeader, filePath); err != nil {
		return webutil.StatusInternalServerError(c)
	}

	// sign process
	certificate := "../cert/ca/user-cert.pem"
	privateKey := "../cert/ca/user-key.pem"

	chainCACrt := "../cert/ca/ca-cert.pem"
	chainCrt := "../cert/ca/ca-intermediate-cert.pem"

	certificateData, err := os.ReadFile(certificate)
	if err != nil {
		log.Fatal(err)
	}
	certificateDataBlock, _ := pem.Decode(certificateData)
	if certificateDataBlock == nil {
		log.Fatal(errors.New("failed to parse PEM block containing the certificate"))
	}

	cert, err := x509.ParseCertificate(certificateDataBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	keyData, err := os.ReadFile(privateKey)
	if err != nil {
		log.Fatal(err)
	}
	keyDataBlock, _ := pem.Decode(keyData)
	if keyDataBlock == nil {
		log.Fatal(errors.New("failed to parse PEM block containing the private key"))
	}

	pkey, err := x509.ParseECPrivateKey(keyDataBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	certificateChains := make([][]*x509.Certificate, 0)

	if chainCACrt != "" && chainCrt != "" {
		chainCAData, err := os.ReadFile(chainCACrt)
		if err != nil {
			log.Fatal(err)
		}
		certificateCAPool := x509.NewCertPool()
		if ok := certificateCAPool.AppendCertsFromPEM(chainCAData); !ok {
			log.Fatalf("failed to parse root certificate")
		}

		chainData, err := os.ReadFile(chainCrt)
		if err != nil {
			log.Fatal(err)
		}
		certificatePool := x509.NewCertPool()
		if ok := certificatePool.AppendCertsFromPEM(chainData); !ok {
			log.Fatalf("failed to parse root certificate")
		}

		certificateChains, err = cert.Verify(x509.VerifyOptions{
			Roots:         certificateCAPool,
			Intermediates: certificatePool,
			CurrentTime:   cert.NotBefore,
			KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	err = sign.SignFile("../../cmd/goSign/lc_uploads/"+response.FileName, "../../cmd/goSign/lc_signed/"+response.FileNameSigned, sign.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:   "User Name",
				Reason: "Signed by ...",
				Date:   time.Now().Local(),
			},
			CertType:   sign.CertificationSignature,
			DocMDPPerm: sign.AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
		Signer:            pkey,
		DigestAlgorithm:   crypto.SHA256,
		Certificate:       cert,
		CertificateChains: certificateChains,
		TSA: sign.TSA{
			URL: "http://timestamp.digicert.com",
		},
	})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Signed PDF written to " + response.FileNameSigned)
	}

	return webutil.Response(c, fiber.StatusOK, "Sign", response)
}
