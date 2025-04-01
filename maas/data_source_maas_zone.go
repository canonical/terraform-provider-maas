package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMAASZone() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an existing MAAS zone.",
		ReadContext: dataSourceZoneRead,

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A brief description of the zone.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The zone's name.",
			},
		},
	}
}

func dataSourceZoneRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	zone, err := getZone(client, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", zone.ID))

	tfstate := map[string]any{
		"name":        zone.Name,
		"description": zone.Description,
	}
	if err := setTerraformState(d, tfstate); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
