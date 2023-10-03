package maas

import (
	"github.com/maas/gomaasclient/client"
)

type Config struct {
	APIKey             string
	APIURL             string
	ApiVersion         string
	InsecureSkipVerify bool
}

func (c *Config) Client() (*client.Client, error) {
	return client.GetClient(c.APIURL, c.APIKey, c.ApiVersion, c.InsecureSkipVerify)
}
