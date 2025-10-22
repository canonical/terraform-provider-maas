package maas

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/canonical/gomaasclient/client"
)

type Config struct {
	APIKey                string
	APIURL                string
	APIVersion            string
	TLSCACertPath         string
	TLSInsecureSkipVerify bool
}

func (c *Config) Client() (*client.Client, error) {
	if !c.useTLS() {
		return client.GetClient(c.APIURL, c.APIKey, c.APIVersion)
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if c.TLSInsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	if c.TLSCACertPath != "" {
		caCert, err := os.ReadFile(c.TLSCACertPath)
		if err != nil {
			return nil, err
		}

		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = pool
	}

	return client.GetTLSClient(c.APIURL, c.APIKey, c.APIVersion, tlsConfig)
}

func (c *Config) useTLS() bool {
	return c.TLSCACertPath != "" || c.TLSInsecureSkipVerify
}
