package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasVlan() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an existing MAAS VLAN.",
		ReadContext: dataSourceVlanRead,

		Schema: map[string]*schema.Schema{
			"dhcp_on": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean value indicating if DHCP is enabled on the VLAN.",
			},
			"fabric": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The fabric identifier (ID or name) for the VLAN.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The MTU used on the VLAN.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VLAN name.",
			},
			"space": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VLAN space.",
			},
			"vlan": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The VLAN identifier (ID or traffic segregation ID).",
			},
		},
	}
}

func dataSourceVlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	vlan, err := getVlan(client, fabric.ID, d.Get("vlan").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	tfState := map[string]interface{}{
		"id":      fmt.Sprintf("%v", vlan.ID),
		"mtu":     vlan.MTU,
		"dhcp_on": vlan.DHCPOn,
		"name":    vlan.Name,
		"space":   vlan.Space,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
