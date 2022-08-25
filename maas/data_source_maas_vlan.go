package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
)

func dataSourceMaasVlan() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an existing MAAS VLAN.",
		ReadContext: dataSourceVlanRead,

		Schema: map[string]*schema.Schema{
			"fabric": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The fabric identifier (ID or name) for the VLAN.",
			},
			"vlan": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The VLAN identifier (ID or traffic segregation ID).",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The MTU used on the VLAN.",
			},
			"dhcp_on": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean value indicating if DHCP is enabled on the VLAN.",
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
		},
	}
}

func dataSourceVlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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
