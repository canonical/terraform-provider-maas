package maas

import (
	"github.com/juju/gomaasapi"
)

type Config struct {
	APIKey     string
	APIURL     string
	ApiVersion string
}

func (c *Config) Client() (interface{}, error) {
	client, err := gomaasapi.NewController(
		gomaasapi.ControllerArgs{
			BaseURL: gomaasapi.AddAPIVersionToURL(c.APIURL, c.ApiVersion),
			APIKey:  c.APIKey,
		})
	if err != nil {
		return nil, err
	}
	return client, nil
}
