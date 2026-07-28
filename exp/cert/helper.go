package main

import (
	"crypto/ecdsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/shurco/gosign/pkg/security/cert"
)

func generatePrivateKey(path string) (*cert.PrivateKey, error) {
	p, err := cert.GetPrivateKey()
	if err != nil {
		return &cert.PrivateKey{}, err
	}
	return p, store(p.String(), path)
}

func store(c, path string) error {
	if isExist(path) {
		return fmt.Errorf("file %s already exists", path)
	}
	return os.WriteFile(path, []byte(c), 0o640)
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func getKeyIdentifier(publicKey *ecdsa.PublicKey) ([]byte, error) {
	b, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	ki := sha1.Sum(b)
	return ki[:], nil
}

func generateCA(pkey *ecdsa.PrivateKey) (*cert.Result, error) {
	ski, err := getKeyIdentifier(&pkey.PublicKey)
	if err != nil {
		return nil, err
	}

	template := cert.Certificate{
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"goSign"},
			CommonName:   "goSign Root CA",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(20, 0, 0),
		IsCA:         true,
		KeyUsage:     x509.KeyUsageCRLSign | x509.KeyUsageCertSign,
		SubjectKeyId: ski,
	}

	caCert, err := template.GetCertificate(pkey)
	if err != nil {
		return nil, err
	}

	return caCert, store(caCert.String(), caPath)
}

func generateCRL(pkey *ecdsa.PrivateKey, caCert *x509.Certificate) error {
	nextUpdate := time.Now().AddDate(20, 0, 0)
	crl, _, err := cert.CreateCRL(pkey, caCert, nil, nextUpdate)
	if err != nil {
		return err
	}

	return store(crl.String(), caCRLPath)
}

func generateIntermediateCert(pkey *ecdsa.PrivateKey) error {
	parentKey, err := getCAPrivateKey(caKeyPath)
	if err != nil {
		return err
	}

	parent, err := getCACert(caPath)
	if err != nil {
		return err
	}

	template := cert.Certificate{
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"goSign"},
			CommonName:   "goSign Sub-CA",
		},
		NotBefore:        time.Now(),
		NotAfter:         time.Now().AddDate(10, 0, 0),
		IsCA:             true,
		Parent:           parent,
		ParentPrivateKey: parentKey,
		KeyUsage:         x509.KeyUsageCRLSign | x509.KeyUsageCertSign,
	}

	cert, err := template.GetCertificate(pkey)
	if err != nil {
		return err
	}

	err = store(cert.String(), caInterPath)
	if err == nil {
		fmt.Println("Certificate file generated", caInterPath)
	}

	return err
}

func getCAPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pkey, err := cert.ParsePrivateKey(f)
	if err != nil {
		return nil, err
	}

	return pkey, nil
}

func getCACert(path string) (*x509.Certificate, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return cert.ParseCertificate(f)
}

func generateCert(pkey *ecdsa.PrivateKey) (err error) {
	var parent *x509.Certificate
	var parentKey *ecdsa.PrivateKey
	var aki, ski []byte

	if parentKey, err = getCAPrivateKey(caInterKeyPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			parentKey, err = getCAPrivateKey(caKeyPath)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if parent, err = getCACert(caInterPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			parent, err = getCACert(caPath)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if aki, err = getKeyIdentifier(&parentKey.PublicKey); err != nil {
		return err
	}

	if ski, err = getKeyIdentifier(&pkey.PublicKey); err != nil {
		return err
	}

	template := cert.Certificate{
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"goSign"},
			CommonName:   "goSign",
		},
		NotBefore:        time.Now(),
		NotAfter:         time.Now().AddDate(0, 0, 7),
		IsCA:             false,
		Parent:           parent,
		ParentPrivateKey: parentKey,
		KeyUsage:         x509.KeyUsageDigitalSignature,
		AuthorityKeyId:   aki,
		SubjectKeyId:     ski,
	}

	cert, err := template.GetCertificate(pkey)
	if err != nil {
		return err
	}

	if err = store(cert.String(), uPath); err == nil {
		fmt.Println("Certificate file generated", uPath)
	}

	return err
}
