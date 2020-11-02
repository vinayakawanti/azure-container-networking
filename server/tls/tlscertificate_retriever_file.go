// Copyright 2020 Microsoft. All rights reserved.

package tls

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/pkcs12"
	"io/ioutil"
)

const (
	CertLabel = "CERTIFICATE"
	PrivateKeyLabel = "PRIVATE KEY"
)

type filetlsCertificateRetriever struct {
	pemBlock []*pem.Block
	settings TlsSettings
}

// GetCertificate Returns the certificate associated with the pfx
func (fcert *filetlsCertificateRetriever) GetCertificate() (*x509.Certificate, error) {
	for _, block := range fcert.pemBlock {
		if block.Type == CertLabel {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse certificate with error %+v", err)
			}
			if cert.IsCA != true {
				return cert, nil
			}
		}
	}
	return nil, fmt.Errorf("No Certificate block found")
}

// GetPrivateKey Returns the private key associated with the pfx
func (fcert *filetlsCertificateRetriever) GetPrivateKey() (crypto.PrivateKey, error) {

	for _, block := range fcert.pemBlock {
		if block.Type == PrivateKeyLabel {
			pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("Could not parse private key %+v", err)
			}
			return pk, nil
		}
	}

	return nil, nil
}

// readPemFile reads a pfx certificate converts it to PEM
func (fcert *filetlsCertificateRetriever) readPemFile() error {
	content, err := ioutil.ReadFile(fcert.settings.TlsCertificateFilePath)
	if err != nil {
		return fmt.Errorf("Error reading file: %+v ", err)
	}
	pemBlock, err := pkcs12.ToPEM(content, "")

	if err != nil {
		return fmt.Errorf("Could not convert pfx to PEM format")
	}

	fcert.pemBlock = pemBlock

	return nil
}

// NewFileTlsCertificateRetriever creates a TlsCertificateRetriever
// NewFileTlsCertificateRetriever depends on the pfx being available
// linux users generally store certificates at /etc/ssl/certs/
func NewFileTlsCertificateRetriever(settings TlsSettings) (TlsCertificateRetriever, error) {
	fileCertStoreRetriever := &filetlsCertificateRetriever{
		settings: settings,
	}

	err := fileCertStoreRetriever.readPemFile()
	if err != nil {
		return nil, fmt.Errorf("Failed to read pfx file with error %+v", err)
	}

	return fileCertStoreRetriever, nil
}
