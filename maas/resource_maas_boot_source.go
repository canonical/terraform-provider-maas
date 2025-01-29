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
				boot_source, err := getBootSource(client, d.Id())
				if err != nil {
					return nil, err
				}
				if err := d.Set("URL", boot_source.URL); err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("%v", boot_source.ID))
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
	boot_sources, err := client.BootSources.Get()
	if err != nil {
		return nil, err
	}
	for _, f := range boot_sources {
		if fmt.Sprintf("%v", f.ID) == identifier || f.URL == identifier {
			return &f, nil
		}
	}
	return nil, nil
}

func getBootSource(client *client.Client, identifier string) (*entity.BootSource, error) {
	boot_source, err := findBootSource(client, identifier)
	if err != nil {
		return nil, err
	}
	if boot_source == nil {
		return nil, fmt.Errorf("boot source (%s) was not found", identifier)
	}
	return boot_source, nil
}
