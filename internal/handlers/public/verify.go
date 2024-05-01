package handlers

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/pdf/verify"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// VerifyPDF is ...
func VerifyPDF(c *fiber.Ctx) error {
	response := &models.Verify{}

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

	tempFile, err := os.CreateTemp("./lc_tmp", "tempfile-*.tmp")
	if err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	if _, err := io.Copy(tempFile, file); err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}

	verifyInfo, err := verify.File(tempFile)
	if err != nil {
		response.Error = err.Error()
		return webutil.StatusOK(c, "Verify", response)
	}

	if verifyInfo.Signers == nil {
		return webutil.Response(c, fiber.StatusOK, "Verify", response)
	}

	for _, value := range verifyInfo.Signers {
		signer := &models.Signer{
			Name:           value.Name,
			Reason:         value.Reason,
			ValidSignature: value.ValidSignature,
			TrustedIssuer: models.TrustedIssuer{
				Valid: value.TrustedIssuer,
			},
			CertSubject: &models.CertSubject{
				Organization: value.Certificates[0].Certificate.Subject.Organization[0],
				CommonName:   value.Certificates[0].Certificate.Subject.CommonName,
			},
			RevokedCertificate: value.RevokedCertificate,
			SigFormat:          value.SigFormat,
		}
		if value.TimeStamp != nil {
			signer.TimeStamp = &models.TimeStamp{
				Time: value.TimeStamp.Time.Unix(),
			}
		}

		// check in trusted base
		if !value.TrustedIssuer {
			for _, cert := range value.Certificates {
				aki := cert.Certificate.AuthorityKeyId
				trustCert, err := queries.DB.CheckAKI(context.Background(), strings.ToUpper(hex.EncodeToString(aki[:])))
				if err != nil && err != pgx.ErrNoRows {
					return webutil.StatusInternalServerError(c)
				}
				fmt.Print(trustCert)
				if trustCert != nil && trustCert.List != "" {
					signer.TrustedIssuer = models.TrustedIssuer{
						Valid: true,
						List:  trustCert.List,
						Name:  trustCert.Name,
					}
				}
			}
		}

		response.Signers = append(response.Signers, signer)
	}

	response.Verify = true
	//response.Document = &Document{
	//	Creator: verifyInfo.DocumentInfo.Creator,
	//	Hash:    verifyInfo.DocumentInfo.Hash,
	//}

	//if verifyInfo.DocumentInfo.Creator == "goSign (https://github.com/shurco/goSign)" {
	//	fmt.Print("check in database")
	//}

	return webutil.Response(c, fiber.StatusOK, "Verify", response)
}
