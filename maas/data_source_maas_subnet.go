package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasSubnet() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an existing MAAS network subnet.",
		ReadContext: dataSourceSubnetRead,

		Schema: map[string]*schema.Schema{
			"allow_dns": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean value that indicates if the MAAS DNS resolution is enabled for this subnet.",
			},
			"allow_proxy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean value that indicates if `maas-proxy` allows requests from this subnet.",
			},
			"cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The subnet CIDR.",
			},
			"dns_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of IP addresses set as DNS servers for the subnet.",
			},
			"fabric": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The subnet fabric.",
			},
			"gateway_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway IP address for the subnet.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The subnet name.",
			},
			"rdns_mode": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "How reverse DNS is handled for this subnet. It can have one of the following values:\n\t* `0` - Disabled, no reverse zone is created.\n\t* `1` - Enabled, generate reverse zone.\n\t* `2` - RFC2317, extends `1` to create the necessary parent zone with the appropriate CNAME resource records for the network, if the network is small enough to require the support described in RFC2317.",
			},
			"vid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The subnet VLAN traffic segregation ID.",
			},
		},
	}
}

func dataSourceSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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
