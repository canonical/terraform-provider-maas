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
		ReadContext: dataSourceVlanRead,

		Schema: map[string]*schema.Schema{
			"fabric_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"vid": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"mtu": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"space": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	fabricID := d.Get("fabric_id").(int)
	vlans, err := client.VLANs.Get(fabricID)
	if err != nil {
		return diag.FromErr(err)
	}

	vid := d.Get("vid").(int)
	for _, vlan := range vlans {
		if vlan.FabricID != fabricID {
			continue
		}
		if vlan.VID != vid {
			continue
		}
		if err := d.Set("mtu", vlan.MTU); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("name", vlan.Name); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("space", vlan.Space); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(fmt.Sprintf("%v", vlan.ID))
		return nil
	}

	return diag.FromErr(fmt.Errorf("could not find matching VLAN"))
}
