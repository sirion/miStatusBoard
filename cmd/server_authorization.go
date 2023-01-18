package main

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func (s *Server) authorized(r *http.Request) *Error {
	switch s.configuration.Authorization.Type {
	case "none":
		return nil

	case "client-cert":
		return s.authorizeClientCert(r)

	case "client-cert-info":
		return s.authorizeClientCertInfo(r)

	default:
		return &Error{
			Code:    500,
			Message: "Invalid authorization configuration",
		}
	}
}

func (s *Server) authorizeClientCert(r *http.Request) *Error {
	// TODO: Parse root certs when reading configuration, not on every request
	if s.configuration.Authorization.Cert == "" {
		return &Error{
			Code:    500,
			Message: "Root certificate not configured",
		}
	}

	cert, serverErr := ParseCertificateBase64(r.Header.Get(s.configuration.Authorization.Header))
	if serverErr != nil {
		return serverErr
	}

	serverErr = VerifyCertificate(cert, s.configuration.Authorization.Cert)
	if serverErr != nil {
		return serverErr
	}

	// Verify against allowlist if there are entries
	if len(s.configuration.Authorization.Users) > 0 {
		if !s.configuration.Authorization.authorizedUsers[cert.Subject.CommonName] {
			return &Error{
				Code:    403,
				Message: "User not authorized",
			}
		}
	}

	return nil
}

func (s *Server) authorizeClientCertInfo(r *http.Request) *Error {
	// verify := r.Header.Get("X-SSL-Client-Verify")
	// if verify != "SUCCESS" {
	// 	return &Error{
	// 		Code:    401,
	// 		Message: "Client not authenticated",
	// 	}
	// }

	sdn := strings.ToLower(r.Header.Get(s.configuration.Authorization.Header))
	if sdn == "" {
		return &Error{
			Code:    401,
			Message: "Client not authenticated",
		}
	}

	var user string
	sndParts := strings.Split(sdn, ",")
	for _, part := range sndParts {
		entry := strings.Split(strings.TrimSpace(part), "=")
		if len(entry) == 2 && entry[0] == "cn" {
			user = entry[1]
			break
		}
	}

	if !s.configuration.Authorization.authorizedUsers[user] {
		return &Error{
			Code:    403,
			Message: "User not authorized",
		}
	}

	return nil
}

/// Functions

func ParseCertificateBase64(certStringBase64 string) (*x509.Certificate, *Error) {
	certData, err := base64.StdEncoding.DecodeString(certStringBase64)
	if err != nil {
		return nil, &Error{
			Code:    400,
			Message: "Client certificate not available",
		}
	}

	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, &Error{
			Code:    400,
			Message: "Client certificate not valid",
		}
	}

	return cert, nil
}

func VerifyCertificate(cert *x509.Certificate, caCertPath string) *Error {
	// Verify against configured root certificates
	certData, err := os.ReadFile(caCertPath)
	if err != nil {
		return &Error{
			Code:    500,
			Message: fmt.Sprintf("Root certificate cannot be read: %s", err.Error()),
		}
	}

	rootCert, err := x509.ParseCertificate(certData)
	if err != nil {
		return &Error{
			Code:    500,
			Message: "Root certificate cannot be used",
		}
	}

	roots := x509.NewCertPool()
	roots.AddCert(rootCert)

	opts := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	_, err = cert.Verify(opts)
	if err != nil {
		return &Error{
			Code:    401,
			Message: "Not authorized",
		}
	}

	return nil
}
