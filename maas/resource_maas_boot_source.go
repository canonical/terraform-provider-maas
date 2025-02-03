package maas

import (
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
)

func findBootSource(client *client.Client) (*entity.BootSource, error) {
	bootsources, err := client.BootSources.Get()
	if err != nil {
		return nil, err
	}
	if len(bootsources) > 0 {
		return &bootsources[0], nil
	}
	return nil, nil
}

func getBootSource(client *client.Client) (*entity.BootSource, error) {
	bootsource, err := findBootSource(client)
	if err != nil {
		return nil, err
	}
	if bootsource == nil {
		return nil, fmt.Errorf("boot source was not found")
	}
	return bootsource, nil
}
