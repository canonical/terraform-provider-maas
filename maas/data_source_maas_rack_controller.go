package maas

import (
	"context"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasRackController() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an existing MAAS rack controller.",
		ReadContext: resourceRackControllerRead,

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the rack controller.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The hostname of the rack controller.",
			},
			"services": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The services running on the rack controller.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the service.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the service.",
						},
					},
				},
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The MAAS version of the rack controller.",
			},
		},
	}
}

func resourceRackControllerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	hostname := d.Get("hostname").(string)
	rackControllers, err := client.RackControllers.Get(
		&entity.RackControllersGetParams{
			Hostname: []string{hostname},
		})
	if err != nil {
		return diag.FromErr(err)
	}
	if len(rackControllers) == 0 {
		return diag.Errorf("rack controller (%s) was not found", hostname)
	}
	d.SetId(rackControllers[0].SystemID)

	d.Set("description", rackControllers[0].Description)
	d.Set("version", rackControllers[0].Description)

	services := make([]map[string]interface{}, len(rackControllers[0].ServiceSet))
	for i, service := range rackControllers[0].ServiceSet {
		services[i] = map[string]interface{}{
			"name":   service.Name,
			"status": service.Status,
		}
	}
	if err := d.Set("services", services); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
