package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maas/gomaasclient/client"
)

func dataSourceMaasNetworkInterfacePhysical() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkInterfacePhysicalRead,

		Schema: map[string]*schema.Schema{
			"machine": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vlan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mtu": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworkInterfacePhysicalRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)
	n, err := getNetworkInterfacePhysical(client, d.Get("machine").(string), d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	tfState := map[string]interface{}{
		"id":          fmt.Sprintf("%v", n.ID),
		"machine":     d.Get("machine").(string),
		"mac_address": n.MACAddress,
		"vlan":        fmt.Sprintf("%v", n.VLAN.ID),
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
