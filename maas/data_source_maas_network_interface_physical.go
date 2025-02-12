package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasNetworkInterfacePhysical() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkInterfacePhysicalRead,

		Schema: map[string]*schema.Schema{
			"mac_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The physical network interface MAC address.",
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier (system ID, hostname, or FQDN) of the machine with the physical network interface.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The MTU of the physical network interface.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The physical network interface name.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A set of tag names assigned to the physical network interface.",
			},
			"vlan": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Database ID of the VLAN the physical network interface is connected to.",
			},
		},
	}
}

func dataSourceNetworkInterfacePhysicalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	n, err := getNetworkInterfacePhysical(client, d.Get("machine").(string), d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	tfState := map[string]interface{}{
		"id":          fmt.Sprintf("%v", n.ID),
		"mac_address": n.MACAddress,
		"machine":     d.Get("machine").(string),
		"mtu":         n.EffectiveMTU,
		"name":        n.Name,
		"tags":        n.Tags,
		"vlan":        n.VLAN.ID,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
