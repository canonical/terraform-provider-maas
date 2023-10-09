package maas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maas/gomaasclient/client"
)

func dataSourceMaasMachine() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMachineRead,

		Schema: map[string]*schema.Schema{
			"power_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"power_parameters": {
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pxe_mac_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"architecture": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"min_hwe_kernel": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)
	machine, err := getMachine(client, d.Get("hostname").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	powerParams, err := client.Machine.GetPowerParameters(machine.SystemID)
	if err != nil {
		return diag.FromErr(err)
	}
	tfState := map[string]interface{}{
		"id":               machine.SystemID,
		"power_type":       machine.PowerType,
		"power_parameters": powerParams,
		"pxe_mac_address":  machine.BootInterface.MACAddress,
		"architecture":     machine.Architecture,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
