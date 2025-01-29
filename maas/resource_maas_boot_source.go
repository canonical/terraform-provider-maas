package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maas/gomaasclient/client"
	"github.com/maas/gomaasclient/entity"
)

func resourceMaasBootSource() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a resource to manage MAAS network boot sources.",
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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The boot source name.",
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
