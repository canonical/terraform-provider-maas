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
	ApiVersion            string
	TLSCACertPath         string
	TLSInsecureSkipVerify bool
}

func (c *Config) Client() (*client.Client, error) {
	if !c.useTLS() {
		return client.GetClient(c.APIURL, c.APIKey, c.ApiVersion)
	}

	tlsConfig := &tls.Config{}
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

	return client.GetTLSClient(c.APIURL, c.APIKey, c.ApiVersion, tlsConfig)
}

func (c *Config) useTLS() bool {
	return c.TLSCACertPath != "" || c.TLSInsecureSkipVerify
}
