package maas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func resourceMaasMachineResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"power_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"power_parameters": {
				Type:      schema.TypeMap,
				Required:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pxe_mac_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"architecture": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "amd64/generic",
			},
			"min_hwe_kernel": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
