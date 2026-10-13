// Package mtls configures mutual TLS for Switch<->Backend.
// Fail-closed: if enabled and certs are missing, the process refuses to start.
package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type Config struct {
	CertFile   string
	KeyFile    string
	CAFile     string
	ClientAuth bool
}

func Load(c Config) (*tls.Config, error) {
	if c.CertFile == "" || c.KeyFile == "" || c.CAFile == "" {
		return nil, fmt.Errorf("mtls: cert, key, and CA paths are all required")
	}
	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("mtls: load keypair: %w", err)
	}
	caPEM, err := os.ReadFile(c.CAFile)
	if err != nil {
		return nil, fmt.Errorf("mtls: read CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("mtls: no certs parsed from CA file")
	}
	t := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		MinVersion:   tls.VersionTLS13,
	}
	if c.ClientAuth {
		t.ClientCAs = pool
		t.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return t, nil
}

func Enabled() bool {
	return os.Getenv("HORIZON_MTLS_ENABLED") == "true"
}

func FromEnv() Config {
	return Config{
		CertFile:   os.Getenv("HORIZON_MTLS_CERT"),
		KeyFile:    os.Getenv("HORIZON_MTLS_KEY"),
		CAFile:     os.Getenv("HORIZON_MTLS_CA"),
		ClientAuth: os.Getenv("HORIZON_MTLS_REQUIRE_CLIENT") == "true",
	}
}
