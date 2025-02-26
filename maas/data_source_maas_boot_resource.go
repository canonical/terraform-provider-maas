package maas

import (
	"context"
	"fmt"

	"github.com/canonical/gomaasclient/entity"
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
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The id of the network interface.",
						},
						"mac_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "MAC address of the network interface.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the network interface.",
						},
					},
				},
			},
			"boot_source": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The boot source database ID this resource is associated with.",
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

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.FromErr(err)
	}
	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}

	var output []entity.BootResource
	for _, res := range resources {
		if res.Name == fmt.Sprintf("%s/%s", d.Get("os"), d.Get("release")) {
			output = append(output, res)
		}
	}

	tfState := map[string]interface{}{
		"boot_resources": output,
		"boot_source":    bootsource.ID,
		"os":             d.Get("os"),
		"release":        d.Get("release"),
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
