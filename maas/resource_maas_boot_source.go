package maas

import (
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
)

func findBootSource(client *client.Client, identifier string) (*entity.BootSource, error) {
	bootsources, err := client.BootSources.Get()
	if err != nil {
		return nil, err
	}
	for _, f := range bootsources {
		if fmt.Sprintf("%v", f.ID) == identifier || f.URL == identifier {
			return &f, nil
		}
	}
	return nil, nil
}

func getBootSource(client *client.Client, identifier string) (*entity.BootSource, error) {
	bootsource, err := findBootSource(client, identifier)
	if err != nil {
		return nil, err
	}
	if bootsource == nil {
		return nil, fmt.Errorf("boot source (%s) was not found", identifier)
	}
	return bootsource, nil
}
