package revocation

import (
	"crypto/x509"
	"encoding/asn1"
)

// InfoArchival is the pkcs7 container containing the revocation information for
// all embedded certificates.
//
// Currently the internal structure is exposed but I don't like to expose the
// asn1.RawValue objects. We can probably make them private and expose the
// information with functions.
type InfoArchival struct {
	CRL   CRL   `asn1:"tag:0,optional,explicit"`
	OCSP  OCSP  `asn1:"tag:1,optional,explicit"`
	Other Other `asn1:"tag:2,optional,explicit"`
}

// AddCRL is used to embed an CRL to revocation.InfoArchival object. You directly
// pass the bytes of a downloaded CRL to this function.
func (r *InfoArchival) AddCRL(b []byte) error {
	r.CRL = append(r.CRL, asn1.RawValue{FullBytes: b})
	return nil
}

// AddOCSP is used to embed the raw bytes of an OCSP response.
func (r *InfoArchival) AddOCSP(b []byte) error {
	r.OCSP = append(r.OCSP, asn1.RawValue{FullBytes: b})
	return nil
}

// IsRevoked checks if there is a status included for the certificate and returns
// true if the certificate is marked as revoked.
//
// TODO: implement full CRL/OCSP status check; currently returns false (not revoked).
// TODO: report if no CRL/OCSP response is embedded for the certificate.
// TODO: expose revocation time, reason, etc.
func (r *InfoArchival) IsRevoked(_ *x509.Certificate) bool {
	return false
}

// CRL contains the raw bytes of a pkix.CertificateList and can be parsed with
// x509.PParseCRL.
type CRL []asn1.RawValue

// OCSP contains the raw bytes of an OCSP response and can be parsed with
// x/crypto/ocsp.ParseResponse
type OCSP []asn1.RawValue

// ANS.1 Object OtherRevInfo
type Other struct {
	Type  asn1.ObjectIdentifier
	Value []byte
}
