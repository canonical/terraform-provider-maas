package maas

import (
	"context"
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/canonical/gomaasclient/entity"
)

func dataSourceMAASBootResources() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMAASBootResourcesRead,
		Description: "Provides a data source to fetch MAAS boot resources.",

		Schema: map[string]*schema.Schema{
			"boot_resources": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The set of boot resources for this os/release",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"architecture": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The architecture of this resource.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of this resource.",
						},
						"subarches": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The subarches for this resource.",
						},
					},
				},
			},
			"os": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The operating system for this resource.",
			},
			"release": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The specific release of the operating system for this resource.",
			},
		},
	}
}

func dataSourceMAASBootResourcesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("boot_resources")

	var foundResources []map[string]any

	for _, res := range resources {
		if res.Name == fmt.Sprintf("%s/%s", d.Get("os"), d.Get("release")) {
			thisResource := map[string]any{
				"name":         res.Name,
				"architecture": res.Architecture,
				"subarches":    res.Subarches,
			}
			foundResources = append(foundResources, thisResource)
		}
	}

	tfState := map[string]any{
		"boot_resources": foundResources,
		"os":             d.Get("os"),
		"release":        d.Get("release"),
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getBootResources(client *client.Client, syncType string) ([]entity.BootResource, error) {
	// syncType: one of synched, uploaded
	readParams := entity.BootResourcesReadParams{
		Type: syncType,
	}

	bootResources, err := client.BootResources.Get(&readParams)
	if err != nil {
		return nil, err
	}

	return bootResources, nil
}
