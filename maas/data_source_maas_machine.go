package maas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func dataSourceMaasMachine() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMachineRead,

		Schema: map[string]*schema.Schema{
			"architecture": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The architecture type of the machine.",
			},
			"domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The domain of the machine.",
			},
			"hostname": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"hostname", "pxe_mac_address"},
				Description:  "The machine hostname.",
			},
			"min_hwe_kernel": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The minimum kernel version allowed to run on this machine.",
			},
			"pool": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource pool of the machine.",
			},
			"power_parameters": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Serialized JSON string containing the parameters specific to the `power_type`. See [Power types](https://maas.io/docs/api#power-types) section for a list of the available power parameters for each power type.",
			},
			"power_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The power management type (e.g. `ipmi`) of the machine.",
			},
			"pxe_mac_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"hostname", "pxe_mac_address"},
				Description:  "The MAC address of the machine's PXE boot NIC.",
			},
			"zone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The zone of the machine.",
			},
		},
	}
}

func dataSourceMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	var identifier string

	if v, ok := d.GetOk("hostname"); ok {
		identifier = v.(string)
	} else if v, ok := d.GetOk("pxe_mac_address"); ok {
		identifier = v.(string)
	}

	machine, err := getMachine(client, identifier)
	if err != nil {
		return diag.FromErr(err)
	}
	powerParams, err := client.Machine.GetPowerParameters(machine.SystemID)
	if err != nil {
		return diag.FromErr(err)
	}
	powerParamsJson, err := structure.FlattenJsonToString(powerParams)
	if err != nil {
		return diag.FromErr(err)
	}
	tfState := map[string]interface{}{
		"id":               machine.SystemID,
		"architecture":     machine.Architecture,
		"min_hwe_kernel":   machine.MinHWEKernel,
		"hostname":         machine.Hostname,
		"domain":           machine.Domain.Name,
		"zone":             machine.Zone.Name,
		"pool":             machine.Pool.Name,
		"power_type":       machine.PowerType,
		"power_parameters": powerParamsJson,
		"pxe_mac_address":  machine.BootInterface.MACAddress,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
