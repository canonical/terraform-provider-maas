package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
)

func dataSourceMaasSubnet() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSubnetRead,

		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:     schema.TypeString,
				Required: true,
			},
			"fabric": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vid": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rdns_mode": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"allow_dns": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allow_proxy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"gateway_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceSubnetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	subnet, err := getSubnet(client, d.Get("cidr").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	gatewayIp := subnet.GatewayIP.String()
	if gatewayIp == "<nil>" {
		gatewayIp = ""
	}
	dnsServers := make([]string, len(subnet.DNSServers))
	for i, ip := range subnet.DNSServers {
		dnsServers[i] = ip.String()
	}
	tfState := map[string]interface{}{
		"id":          fmt.Sprintf("%v", subnet.ID),
		"fabric":      subnet.VLAN.Fabric,
		"vid":         subnet.VLAN.VID,
		"name":        subnet.Name,
		"rdns_mode":   subnet.RDNSMode,
		"allow_dns":   subnet.AllowDNS,
		"allow_proxy": subnet.AllowProxy,
		"gateway_ip":  gatewayIp,
		"dns_servers": dnsServers,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
