package maas

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasSubnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSubnetCreate,
		ReadContext:   resourceSubnetRead,
		UpdateContext: resourceSubnetUpdate,
		DeleteContext: resourceSubnetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
				subnet, err := getSubnet(client, d.Id())
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
					"id":          fmt.Sprintf("%v", subnet.ID),
					"cidr":        subnet.CIDR,
					"name":        subnet.Name,
					"fabric":      fmt.Sprintf("%v", subnet.VLAN.FabricID),
					"vlan":        fmt.Sprintf("%v", subnet.VLAN.VID),
					"rdns_mode":   subnet.RDNSMode,
					"allow_dns":   subnet.AllowDNS,
					"allow_proxy": subnet.AllowProxy,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"fabric": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vlan": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"fabric"},
			},
			"ip_ranges": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"dynamic", "reserved"}, false)),
						},
						"start_ip": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
						},
						"end_ip": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
						},
						"comment": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"rdns_mode": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          2,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 2)),
			},
			"allow_dns": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"allow_proxy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"gateway_ip": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
			},
			"dns_servers": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					ValidateDiagFunc: isElementIPAddress,
					Type:             schema.TypeString,
				},
			},
		},
	}
}

func resourceSubnetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	params, err := getSubnetParams(client, d)
	if err != nil {
		return diag.FromErr(err)
	}
	subnet, err := client.Subnets.Create(params)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", subnet.ID))

	return resourceSubnetUpdate(ctx, d, m)
}

func resourceSubnetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	subnet, err := client.Subnet.Get(id)
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
		"gateway_ip":  gatewayIp,
		"dns_servers": dnsServers,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSubnetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	params, err := getSubnetParams(client, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := client.Subnet.Update(id, params); err != nil {
		return diag.FromErr(err)
	}
	if err := updateIPRanges(client, d, id); err != nil {
		return diag.FromErr(err)
	}

	return resourceSubnetRead(ctx, d, m)
}

func resourceSubnetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.Subnet.Delete(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateIPRanges(client *client.Client, d *schema.ResourceData, subnetID int) error {
	p, ok := d.GetOk("ip_ranges")
	if !ok {
		return nil
	}
	// Removing existing IP ranges on this subnet
	ipRanges, err := client.IPRanges.Get()
	if err != nil {
		return err
	}
	for _, ipr := range ipRanges {
		if ipr.Subnet.ID != subnetID {
			continue
		}
		if err := client.IPRange.Delete(ipr.ID); err != nil {
			return err
		}
	}
	// Create the new IP ranges on this subnet
	for _, i := range p.(*schema.Set).List() {
		ipr := i.(map[string]interface{})
		params := entity.IPRangeParams{
			Subnet:  fmt.Sprintf("%v", subnetID),
			Type:    ipr["type"].(string),
			StartIP: ipr["start_ip"].(string),
			EndIP:   ipr["end_ip"].(string),
			Comment: ipr["comment"].(string),
		}
		if _, err := client.IPRanges.Create(&params); err != nil {
			return err
		}
	}
	return nil
}

func getSubnetParams(client *client.Client, d *schema.ResourceData) (*entity.SubnetParams, error) {
	params := entity.SubnetParams{
		CIDR:       d.Get("cidr").(string),
		Name:       d.Get("name").(string),
		RDNSMode:   d.Get("rdns_mode").(int),
		AllowDNS:   d.Get("allow_dns").(bool),
		AllowProxy: d.Get("allow_proxy").(bool),
		GatewayIP:  d.Get("gateway_ip").(string),
		DNSServers: convertToStringSlice(d.Get("dns_servers")),
		Managed:    true,
	}
	if p, ok := d.GetOk("fabric"); ok {
		fabric, err := getFabric(client, p.(string))
		if err != nil {
			return nil, err
		}
		params.Fabric = fmt.Sprintf("%v", fabric.ID)
		if p, ok := d.GetOk("vlan"); ok {
			vlan, err := getVlan(client, fabric.ID, p.(string))
			if err != nil {
				return nil, err
			}
			params.VLAN = fmt.Sprintf("%v", vlan.ID)
			params.VID = vlan.VID
		}
	}
	return &params, nil
}

func findSubnet(client *client.Client, identifier string) (*entity.Subnet, error) {
	subnets, err := client.Subnets.Get()
	if err != nil {
		return nil, err
	}
	for _, s := range subnets {
		if fmt.Sprintf("%v", s.ID) == identifier || s.CIDR == identifier {
			return &s, nil
		}
	}
	return nil, nil
}

func getSubnet(client *client.Client, identifier string) (*entity.Subnet, error) {
	subnet, err := findSubnet(client, identifier)
	if err != nil {
		return nil, err
	}
	if subnet == nil {
		return nil, fmt.Errorf("subnet (%s) was not found", identifier)
	}
	return subnet, nil
}
