// Copyright 2020 Microsoft. All rights reserved.

package tls

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/billgraziano/dpapi"
	"io/ioutil"
	"strings"
)

const (
	CertLabel       = "CERTIFICATE"
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
				return nil, fmt.Errorf("Failed to parse certificate at location %s with error %+v", fcert.settings.TLSCertificatePath, err)
			}
			if !cert.IsCA {
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
			pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("Could not parse private key %+v", err)
			}
			return pk, nil
		}
	}
	return nil, fmt.Errorf("No private key found in certificate bundle located at %s", fcert.settings.TLSCertificatePath)
}

// readPemFile reads a pfx certificate converts it to PEM
func (fcert *filetlsCertificateRetriever) readPemFile() error {
	content, err := ioutil.ReadFile(fcert.settings.TLSCertificatePath)
	if err != nil {
		return fmt.Errorf("Error reading file from path %s with error: %+v ", fcert.settings.TLSCertificatePath, err)
	}

	decrypted, err := dpapi.Decrypt(string(content))
	decrypted = formatDecryptedPemString(decrypted)
	if err != nil {
		return fmt.Errorf("Error reading file from path %s with error: %+v ", fcert.settings.TLSCertificatePath, err)
	}

	pemBlocks := make([]*pem.Block, 0)
	var pemBlock *pem.Block
	nextPemBlock := []byte(decrypted)

	for {
		pemBlock, nextPemBlock = pem.Decode(nextPemBlock)

		if pemBlock == nil {
			break
		}
		pemBlocks = append(pemBlocks, pemBlock)
	}

	if len(pemBlocks) < 2 {
		return fmt.Errorf("Invalid PEM format located at %s", fcert.settings.TLSCertificatePath)
	}

	fcert.pemBlock = pemBlocks
	return nil
}

// formatDecryptedPemString ensures pem format
// removes spaces that should be line breaks
// ensures headers are properly formatted
// removes null terminated strings that dpapi.decrypt introduces
func formatDecryptedPemString(s string) string {
	s = strings.ReplaceAll(s, " ", "\r\n")
	s = strings.ReplaceAll(s, "\000", "")
	s = strings.ReplaceAll(s, "-----BEGIN\r\nPRIVATE\r\nKEY-----", "-----BEGIN PRIVATE KEY-----")
	s = strings.ReplaceAll(s, "-----END\r\nPRIVATE\r\nKEY-----", "-----END PRIVATE KEY-----")
	s = strings.ReplaceAll(s, "-----BEGIN\r\nCERTIFICATE-----", "-----BEGIN CERTIFICATE-----")
	s = strings.ReplaceAll(s, "-----END\r\nCERTIFICATE-----", "-----END CERTIFICATE-----")
	return s
}

// NewFileTlsCertificateRetriever creates a TlsCertificateRetriever
// NewFileTlsCertificateRetriever depends on the pfx being available
// linux users generally store certificates at /etc/ssl/certs/
func NewFileTlsCertificateRetriever(settings TlsSettings) (TlsCertificateRetriever, error) {
	fileCertStoreRetriever := &filetlsCertificateRetriever{
		settings: settings,
	}
	if err := fileCertStoreRetriever.readPemFile(); err != nil {
		return nil, fmt.Errorf("Failed to read pfx file with error %+v", err)
	}
	return fileCertStoreRetriever, nil
}