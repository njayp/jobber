package authn

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"google.golang.org/grpc/credentials"
)

func LoadTLSCredentials() (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile("ca.crt")
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append CA certificate")
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		// pool of acceptable client CAs
		ClientCAs: caCertPool,
		// require verify client CA
		ClientAuth: tls.RequireAndVerifyClientCert,
		// set minimum version to latest
		MinVersion: tls.VersionTLS13,
		// TLS 1.3 ciphersuites are not configurable
	}

	return credentials.NewTLS(config), nil
}
