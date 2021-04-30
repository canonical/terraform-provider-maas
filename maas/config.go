package maas

import (
	"github.com/ionutbalutoiu/gomaasclient/client"
)

type Config struct {
	APIKey     string
	APIURL     string
	ApiVersion string
}

func (c *Config) Client() (*client.Client, error) {
	return client.GetClient(c.APIURL, c.APIKey, c.ApiVersion)
}
