package maas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMAASDevices() *schema.Resource {
	return &schema.Resource{
		Description: "Lists MAAS devices visible to the user.",
		ReadContext: dataSourceDevicesRead,

		Schema: map[string]*schema.Schema{
			"devices": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of devices visible to the user.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The device hostname.",
						},
						"system_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The system ID of the device.",
						},
					},
				},
			},
		},
	}
}

func dataSourceDevicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	devices, err := client.Devices.Get()
	if err != nil {
		return diag.FromErr(err)
	}

	items := []map[string]interface{}{}
	for _, device := range devices {
		items = append(items, map[string]interface{}{
			"system_id": device.SystemID,
			"hostname":  device.Hostname,
		})
	}

	if err := d.Set("devices", items); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("maas_devices")

	return nil
}
