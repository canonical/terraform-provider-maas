package maas

import (
	"github.com/juju/gomaasapi"

	"github.com/ionutbalutoiu/gomaasclient/gmaw"
)

type Config struct {
	APIKey     string
	APIURL     string
	ApiVersion string
}

func (c *Config) Client() (*gomaasapi.MAASObject, error) {
	client, err := gmaw.GetClient(c.APIURL, c.APIKey, c.ApiVersion)
	if err != nil {
		return nil, err
	}
	return client, nil
}
