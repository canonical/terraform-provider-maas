package maas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func resourceMaasMachineResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"architecture": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "amd64/generic",
				Description: "The architecture type of the machine. Defaults to `amd64/generic`.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The domain of the machine. This is computed if it's not set.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The machine hostname. This is computed if it's not set.",
			},
			"min_hwe_kernel": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The minimum kernel version allowed to run on this machine. Only used when deploying Ubuntu. This is computed if it's not set.",
			},
			"pool": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource pool of the machine. This is computed if it's not set.",
			},
			"power_parameters": {
				Type:      schema.TypeMap,
				Required:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Serialized JSON string containing the parameters specific to the `power_type`. See [Power types](https://maas.io/docs/api#power-types) section for a list of the available power parameters for each power type.",
			},
			"power_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A power management type (e.g. `ipmi`).",
			},
			"pxe_mac_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The MAC address of the machine's PXE boot NIC.",
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The zone of the machine. This is computed if it's not set.",
			},
		},
	}
}

func resourceMaasMachineStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	// Convert power_parameters from map[string]string to a serialized JSON string.
	oldPowerParametersRaw := rawState["power_parameters"].(map[string]interface{})
	flattenedOldPowerParameters, err := structure.FlattenJsonToString(oldPowerParametersRaw)
	if err != nil {
		return nil, err
	}

	rawState["power_parameters"] = flattenedOldPowerParameters

	return rawState, nil
}
