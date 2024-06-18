package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"google.golang.org/grpc/credentials"
)

func loadClientTLSCredentials() (credentials.TransportCredentials, error) {
	clientCert, err := tls.LoadX509KeyPair("client.crt", "client.key")
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
		Certificates: []tls.Certificate{clientCert},
		// pool of acceptable server CAs
		RootCAs: caCertPool,
		// set minimum version to latest
		MinVersion: tls.VersionTLS13,
		// TLS 1.3 ciphersuites are not configurable
	}

	return credentials.NewTLS(config), nil
}
