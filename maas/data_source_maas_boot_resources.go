package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasBootResources() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasBootResourcesRead,
		Description: "Provides a data source to manage MAAS bootresources.",

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
						"last_deployed": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time of last deploy for this resource",
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

func dataSourceMaasBootResourcesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	err := awaitImportComplete(client)
	if err != nil {
		return diag.Errorf("Could not await image importing: %v", err)
	}

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.FromErr(err)
	}
	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", bootsource.ID))

	var foundresources []map[string]interface{}
	for _, res := range resources {
		if res.Name == fmt.Sprintf("%s/%s", d.Get("os"), d.Get("release")) {
			this_resource := map[string]interface{}{
				"name":          res.Name,
				"architecture":  res.Architecture,
				"last_deployed": res.LastDeployed,
				"subarches":     res.Subarches,
			}
			foundresources = append(foundresources, this_resource)
		}
	}

	tfState := map[string]interface{}{
		"boot_resources": foundresources,
		"os":             d.Get("os"),
		"release":        d.Get("release"),
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
