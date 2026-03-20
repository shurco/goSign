package verify

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shurco/gosign/pkg/pdf/sign"
)

// TestFile_afterSign exercises verify.File on a freshly signed PDF (same cert chain as pkg/pdf/sign tests).
func TestFile_afterSign(t *testing.T) {
	const certPEM = `-----BEGIN CERTIFICATE-----
MIIDBzCCAnCgAwIBAgIJAIJ/XyRx/DG0MA0GCSqGSIb3DQEBCwUAMIGZMQswCQYD
VQQGEwJOTDEVMBMGA1UECAwMWnVpZC1Ib2xsYW5kMRIwEAYDVQQHDAlSb3R0ZXJk
YW0xEjAQBgNVBAoMCVVuaWNvZGVyczELMAkGA1UECwwCSVQxGjAYBgNVBAMMEUpl
cm9lbiBCb2JiZWxkaWprMSIwIAYJKoZIhvcNAQkBFhNqZXJvZW5AdW5pY29kZXJz
Lm5sMCAXDTE3MDkxNzExMjkzNloYDzMwMTcwMTE4MTEyOTM2WjCBmTELMAkGA1UE
BhMCTkwxFTATBgNVBAgMDFp1aWQtSG9sbGFuZDESMBAGA1UEBwwJUm90dGVyZGFt
MRIwEAYDVQQKDAlVbmljb2RlcnMxCzAJBgNVBAsMAklUMRowGAYDVQQDDBFKZXJv
ZW4gQm9iYmVsZGlqazEiMCAGCSqGSIb3DQEJARYTamVyb2VuQHVuaWNvZGVycy5u
bDCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAmrvrZiUZZ/nSmFKMsQXg5slY
TQjj7nuenczt7KGPVuGA8nNOqiGktf+yep5h2r87jPvVjVXjJVjOTKx9HMhaFECH
KHKV72iQhlw4fXa8iB1EDeGuwP+pTpRWlzurQ/YMxvemNJVcGMfTE42X5Bgqh6Dv
kddRTAeeqQDBD6+5VPsCAwEAAaNTMFEwHQYDVR0OBBYEFETizi2bTLRMIknQXWDR
nQ59xI99MB8GA1UdIwQYMBaAFETizi2bTLRMIknQXWDRnQ59xI99MA8GA1UdEwEB
/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADgYEAkOHdI9f4I1rd7DjOXnT6IJl/4mIQ
kkaeZkjcsgdZAeW154vjDEr8sIdq+W15huWJKZkqwhn1sJLqSOlEhaYbJJNHVKc9
ZH5r6ujfc336AtjrjCL3OYHQQj05isKm9ii5IL/i+rlZ5xro/dJ91jnjqNVQPvso
oA4h5BVsLZPIYto=
-----END CERTIFICATE-----`
	const keyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQCau+tmJRln+dKYUoyxBeDmyVhNCOPue56dzO3soY9W4YDyc06q
IaS1/7J6nmHavzuM+9WNVeMlWM5MrH0cyFoUQIcocpXvaJCGXDh9dryIHUQN4a7A
/6lOlFaXO6tD9gzG96Y0lVwYx9MTjZfkGCqHoO+R11FMB56pAMEPr7lU+wIDAQAB
AoGADPlKsILV0YEB5mGtiD488DzbmYHwUpOs5gBDxr55HUjFHg8K/nrZq6Tn2x4i
iEvWe2i2LCaSaBQ9H/KqftpRqxWld2/uLbdml7kbPh0+57/jsuZZs3jlN76HPMTr
uYcfG2UiU/wVTcWjQLURDotdI6HLH2Y9MeJhybctywDKWaECQQDNejmEUybbg0qW
2KT5u9OykUpRSlV3yoGlEuL2VXl1w5dUMa3rw0yE4f7ouWCthWoiCn7dcPIaZeFf
5CoshsKrAkEAwMenQppKsLk62m8F4365mPxV/Lo+ODg4JR7uuy3kFcGvRyGML/FS
TB5NI+DoTmGEOZVmZeLEoeeSnO0B52Q28QJAXFJcYW4S+XImI1y301VnKsZJA/lI
KYidc5Pm0hNZfWYiKjwgDtwzF0mLhPk1zQEyzJS2p7xFq0K3XqRfpp3t/QJACW77
sVephgJabev25s4BuQnID2jxuICPxsk/t2skeSgUMq/ik0oE0/K7paDQ3V0KQmMc
MqopIx8Y3pL+f9s4kQJADWxxuF+Rb7FliXL761oa2rZHo4eciey2rPhJIU/9jpCc
xLqE5nXC5oIUTbuSK+b/poFFrtjKUFgxf0a/W2Ktsw==
-----END RSA PRIVATE KEY-----`

	cb, _ := pem.Decode([]byte(certPEM))
	kb, _ := pem.Decode([]byte(keyPEM))
	if cb == nil || kb == nil {
		t.Fatal("pem decode")
	}
	cert, err := x509.ParseCertificate(cb.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	pkey, err := x509.ParsePKCS1PrivateKey(kb.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	in := filepath.Join(testRepoRoot(t), "fixtures", "testfiles", "testfile20.pdf")
	out := filepath.Join(t.TempDir(), "signed.pdf")
	if err := sign.SignFile(in, out, sign.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        "Verify Integration",
				Location:    "CI",
				Reason:      "Test",
				ContactInfo: "None",
				Date:        time.Now().Local(),
			},
			CertType:   sign.CertificationSignature,
			DocMDPPerm: sign.AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
		DigestAlgorithm: crypto.SHA512,
		Signer:          pkey,
		Certificate:     cert,
	}); err != nil {
		t.Fatalf("SignFile: %v", err)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })

	resp, err := File(f)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if resp == nil || len(resp.Signers) == 0 {
		t.Fatalf("expected signers in response")
	}
}
