package main

import (
	"fmt"
	"log"
)

var (
	caKeyPath      = "./ca/ca-key.pem"
	caPath         = "./ca/ca-cert.pem"
	caCRLPath      = "./ca/ca-crl.pem"
	caInterKeyPath = "./ca/ca-intermediate-key.pem"
	caInterPath    = "./ca/ca-intermediate-cert.pem"
	uKeyPath       = "./ca/user-key.pem"
	uPath          = "./ca/user-cert.pem"
)

func main() {
	// CA cert
	pKey, err := generatePrivateKey(caKeyPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("CA private key file generated", caKeyPath)

	caCert, err := generateCA(pKey.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("CA certificate file generated")

	if err := generateCRL(pKey.PrivateKey, caCert.Cert); err != nil {
		log.Fatal(err)
	}
	fmt.Println("CRL file generated", caCRLPath)

	// CA Intermed cert
	ipKey, err := generatePrivateKey(caInterKeyPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Private key file generated", caInterKeyPath)

	if err := generateIntermediateCert(ipKey.PrivateKey); err != nil {
		log.Fatal(err)
	}
	fmt.Println("CA Intermed certificate file generated")

	// user cert
	upKey, err := generatePrivateKey(uKeyPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("user private key file generated", uKeyPath)

	if err := generateCert(upKey.PrivateKey); err != nil {
		log.Fatal(err)
	}
	fmt.Println("user certificate file generated", uPath)
}
