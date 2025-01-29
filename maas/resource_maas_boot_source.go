package maas

import (
	"context"
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasBootSource() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a resource to manage MAAS boot sources.",
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				client := meta.(*client.Client)
				bootsource, err := getBootSource(client, d.Id())
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
					"id":               bootsource.ID,
					"created":          bootsource.Created,
					"keyring_data":     bootsource.KeyringData,
					"keyring_filename": bootsource.KeyringFilename,
					"resource_uri":     bootsource.ResourceURI,
					"updated":          bootsource.Updated,
					"url":              bootsource.URL,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the boot source.",
			},
			"keyring_data": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The data on the keyring for the boot source.",
			},
			"keyring_filename": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The filename on the keyring for the boot source.",
			},
			"resource_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource URI for the book source.",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time of most recent update of the boot source.",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL of the boot source.",
			},
		},
	}
}

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
