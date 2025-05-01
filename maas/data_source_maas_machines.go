package maas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMAASMachines() *schema.Resource {
	return &schema.Resource{
		Description: "Lists MAAS machines visible to the user.",
		ReadContext: dataSourceMachinesRead,

		Schema: map[string]*schema.Schema{
			"machines": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of machines visible to the user.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The machine hostname.",
						},
						"system_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The system ID of the machine.",
						},
					},
				},
			},
		},
	}
}

func dataSourceMachinesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machines, err := client.Machines.Get(nil)
	if err != nil {
		return diag.FromErr(err)
	}

	items := []map[string]interface{}{}
	for _, machine := range machines {
		items = append(items, map[string]interface{}{
			"system_id": machine.SystemID,
			"hostname":  machine.Hostname,
		})
	}

	if err := d.Set("machines", items); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("maas_machines")

	return nil
}
