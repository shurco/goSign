package models

type Verify struct {
	Verify   bool          `json:"verify"`
	Error    string        `json:"error,omitempty"`
	Document *DocumentInfo `json:"document"`
	Signers  []*Signer     `json:"signers"`
}

type DocumentInfo struct {
	Creator string `json:"creator"`
	Hash    string `json:"hash"`
}

type Signer struct {
	Name               string       `json:"name"`
	Reason             string       `json:"reason"`
	ValidSignature     bool         `json:"valid_signature"`
	TrustedIssuer      bool         `json:"trusted_issuer"`
	CertSubject        *CertSubject `json:"cert_subject"`
	SigFormat          string       `json:"sig_format"`
	RevokedCertificate bool         `json:"revoked_certificate"`
	TimeStamp          *TimeStamp   `json:"time_stamp"`
}

type TimeStamp struct {
	Time int64 `json:"time"`
}

type CertSubject struct {
	Organisation []string `json:"organisation"`
	CommonName   string   `json:"common_name"`
}
