package maas

import (
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
)

func getBootSource(client *client.Client) (*entity.BootSource, error) {
	bootsources, err := client.BootSources.Get()
	if err != nil {
		return nil, err
	}
	if len(bootsources) == 0 {
		return nil, fmt.Errorf("boot source was not found")
	}
	if len(bootsources) > 1 {
		return nil, fmt.Errorf("expected a single boot source")
	}
	return &bootsources[0], nil
}
