package models

import "time"

// Verify is ...
type Verify struct {
	Verify   bool          `json:"verify"`
	Error    string        `json:"error,omitempty"`
	Document *DocumentInfo `json:"document"`
	Signers  []*Signer     `json:"signers"`
}

// DocumentInfo is ...
type DocumentInfo struct {
	Creator string `json:"creator"`
	Hash    string `json:"hash"`
}

// Signer is ...
type Signer struct {
	Name               string        `json:"name"`
	Reason             string        `json:"reason"`
	ValidSignature     bool          `json:"valid_signature"`
	TrustedIssuer      TrustedIssuer `json:"trusted_issuer"`
	CertSubject        *CertSubject  `json:"cert_subject"`
	SigFormat          string        `json:"sig_format"`
	RevokedCertificate bool          `json:"revoked_certificate"`
	TimeStamp          *TimeStamp    `json:"time_stamp"`
}

// TrustedIssuer is ...
type TrustedIssuer struct {
	Valid bool   `json:"valid"`
	List  string `json:"list"`
	Name  string `json:"name"`
}

// TimeStamp is ...
type TimeStamp struct {
	Time int64 `json:"time"`
}

// CertSubject is ...
type CertSubject struct {
	Organization string `json:"organization"`
	CommonName   string `json:"common_name"`
}

// TrustList is ...
type TrustCerts struct {
	Certs []TrustCert `json:"trust_cert"`
}

// TrustCert is ...
type TrustCert struct {
	List      string     `json:"list"`
	Name      string     `json:"name"`
	AKI       string     `json:"aki"`
	SKI       string     `json:"ski"`
	CreatedAt *time.Time `json:"created,omitempty"`
}
