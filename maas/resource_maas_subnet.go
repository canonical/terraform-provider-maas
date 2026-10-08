package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/go-set/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMAASSubnet() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS network subnets.\n\n**NOTE:** The MAAS provider currently supports both standalone resources and in-line resources for subnet IP ranges. You cannot use in-line `ip_ranges` in conjunction with standalone `maas_subnet_ip_range` resources. Doing so will cause conflicts and will overwrite subnet IP ranges.",
		CreateContext: resourceSubnetCreate,
		ReadContext:   resourceSubnetRead,
		UpdateContext: resourceSubnetUpdate,
		DeleteContext: resourceSubnetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				subnet, err := getSubnet(client, d.Id())
				if err != nil {
					return nil, err
				}

				d.SetId(strconv.Itoa(subnet.ID))

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"allow_dns": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Boolean value that indicates if the MAAS DNS resolution is enabled for this subnet. Defaults to `true`.",
			},
			"allow_proxy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Boolean value that indicates if `maas-proxy` allows requests from this subnet. Defaults to `true`.",
			},
			"cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The subnet CIDR.",
			},
			"dns_servers": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "List of IP addresses set as DNS servers for the new subnet. If this argument is omitted, it is computed from MAAS and left unmanaged. Set it explicitly to `[]` to remove all DNS servers from the subnet.",
				Elem: &schema.Schema{
					ValidateDiagFunc: isElementIPAddress,
					Type:             schema.TypeString,
				},
			},
			"fabric": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The fabric identifier (ID or name) for the new subnet. This argument is computed if it's not set.",
			},
			"gateway_ip": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
				Description:      "Gateway IP address for the new subnet. This argument is computed if it's not set.",
			},
			"ip_ranges": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of IP ranges configured on the new subnet. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).",
				Deprecated:  "This field is deprecated and will be removed in a future release. Use the `maas_subnet_ip_range` resource to manage subnet IP ranges instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comment": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A description of this range.",
						},
						"end_ip": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
							Description:      "The end IP for the new IP range (inclusive).",
						},
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the IP range.",
						},
						"start_ip": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
							Description:      "The start IP for the new IP range (inclusive).",
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"dynamic", "reserved"}, false)),
							Description:      "The IP range type. Valid options are: `dynamic`, `reserved`.",
						},
					},
				},
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The subnet name.",
			},
			"rdns_mode": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          2,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 2)),
				Description:      "How reverse DNS is handled for this subnet. Defaults to `2`. Valid options are:\n\t* `0` - Disabled, no reverse zone is created.\n\t* `1` - Enabled, generate reverse zone.\n\t* `2` - RFC2317, extends `1` to create the necessary parent zone with the appropriate CNAME resource records for the network, if the network is small enough to require the support described in RFC2317.",
			},
			"vlan": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"fabric"},
				Description:  "The VLAN identifier (ID or traffic segregation ID; if a value matches both, the traffic segregation ID wins) for the new subnet. If this is set, the `fabric` argument is required. This argument is computed if it's not set.",
			},
		},
	}
}

func resourceSubnetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	params, err := getSubnetParams(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	subnet, err := client.Subnets.Create(params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(subnet.ID))

	if err := createIPRanges(client, d, subnet.ID); err != nil {
		return diag.FromErr(err)
	}

	return resourceSubnetRead(ctx, d, meta)
}

func resourceSubnetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	subnet, err := client.Subnet.Get(id)
	if err != nil {
		return unsetIfNotFoundError(d, err)
	}

	// Handle potential "<nil>" value from net.IP
	gatewayIP := ipToString(subnet.GatewayIP)

	// Populate dns_servers from MAAS. When the argument is omitted from the
	// configuration it is computed, so whatever MAAS reports is authoritative.
	dnsServers := make([]string, len(subnet.DNSServers))
	for i, ip := range subnet.DNSServers {
		dnsServers[i] = ipToString(ip)
	}

	tfState := map[string]any{
		"allow_dns":   subnet.AllowDNS,
		"allow_proxy": subnet.AllowProxy,
		"cidr":        subnet.CIDR,
		"dns_servers": dnsServers,
		"fabric":      strconv.Itoa(subnet.VLAN.FabricID),
		"gateway_ip":  gatewayIP,
		"name":        subnet.Name,
		"rdns_mode":   subnet.RDNSMode,
		"vlan":        strconv.Itoa(subnet.VLAN.VID),
	}

	// Only manage ip_ranges if they're configured or already tracked in state
	if _, ok := d.GetOk("ip_ranges"); ok {
		allIPRanges, err := client.IPRanges.Get()
		if err != nil {
			return diag.FromErr(err)
		}

		// Filter IP ranges belonging to this subnet and build the set
		ipRangesSet := make([]map[string]any, 0)

		for _, ipr := range allIPRanges {
			if ipr.Subnet.ID == id {
				ipRangeMap := map[string]any{
					"id":       ipr.ID,
					"type":     ipr.Type,
					"start_ip": ipToString(ipr.StartIP),
					"end_ip":   ipToString(ipr.EndIP),
					"comment":  ipr.Comment,
				}
				ipRangesSet = append(ipRangesSet, ipRangeMap)
			}
		}

		tfState["ip_ranges"] = ipRangesSet
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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

	return resourceSubnetRead(ctx, d, meta)
}

func resourceSubnetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Delete IP ranges tracked in state first
	if ipRangesRaw, ok := d.GetOk("ip_ranges"); ok {
		ipRangesSet := ipRangesRaw.(*schema.Set)
		for _, i := range ipRangesSet.List() {
			ipr := i.(map[string]any)
			if rangeID, ok := ipr["id"]; ok {
				err = client.IPRange.Delete(rangeID.(int))
				if err != nil && !strings.Contains(err.Error(), "404 Not Found") {
					return diag.FromErr(err)
				}
			}
		}
	}

	if err := client.Subnet.Delete(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateIPRanges(client *client.Client, d *schema.ResourceData, subnetID int) error {
	if !d.HasChange("ip_ranges") {
		return nil
	}

	oldRanges, newRanges := d.GetChange("ip_ranges")
	oldSet := oldRanges.(*schema.Set)
	newSet := newRanges.(*schema.Set)

	// Build sets of old and new IDs for comparison
	oldIDs := set.New[int](0)
	newIDs := set.New[int](0)

	for _, i := range oldSet.List() {
		if id, ok := i.(map[string]any)["id"]; ok {
			oldIDs.Insert(id.(int))
		}
	}

	for _, i := range newSet.List() {
		if id, ok := i.(map[string]any)["id"]; ok {
			newIDs.Insert(id.(int))
		}
	}

	// First, delete any IP ranges that were removed or changed
	// (changed ranges will have new hash, so they won't be in newIDs)
	for _, id := range oldIDs.Slice() {
		if !newIDs.Contains(id) {
			if err := client.IPRange.Delete(id); err != nil {
				return fmt.Errorf("failed to delete IP range %d: %w", id, err)
			}
		}
	}

	// Then, update existing ranges or create new ones
	for _, i := range newSet.List() {
		newRange := i.(map[string]any)
		params := entity.IPRangeParams{
			Subnet:  strconv.Itoa(subnetID),
			Type:    newRange["type"].(string),
			StartIP: newRange["start_ip"].(string),
			EndIP:   newRange["end_ip"].(string),
			Comment: newRange["comment"].(string),
		}

		// If this range has an ID from old state, update it
		if id, ok := newRange["id"]; ok {
			rangeID := id.(int)
			if oldIDs.Contains(rangeID) {
				if _, err := client.IPRange.Update(rangeID, &params); err != nil {
					return fmt.Errorf("failed to update IP range %d: %w", rangeID, err)
				}

				continue
			}
		}

		// No ID or ID not in old set means this is a new range, create it
		if _, err := client.IPRanges.Create(&params); err != nil {
			return fmt.Errorf("failed to create IP range %s-%s: %w", params.StartIP, params.EndIP, err)
		}
	}

	return nil
}

func createIPRanges(client *client.Client, d *schema.ResourceData, subnetID int) error {
	p, ok := d.GetOk("ip_ranges")
	if !ok {
		return nil
	}

	for _, i := range p.(*schema.Set).List() {
		ipr := i.(map[string]any)

		params := entity.IPRangeParams{
			Subnet:  strconv.Itoa(subnetID),
			Type:    ipr["type"].(string),
			StartIP: ipr["start_ip"].(string),
			EndIP:   ipr["end_ip"].(string),
			Comment: ipr["comment"].(string),
		}
		if _, err := client.IPRanges.Create(&params); err != nil {
			return fmt.Errorf("failed to create IP range %s-%s: %w", params.StartIP, params.EndIP, err)
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
		Managed:    true,
	}

	// There are three cases for dns_servers, to distinguish them we need GetRawConfig
	// because d.Get cannot tell an explicit empty list apart from an unset one on an
	// Optional+Computed attribute:
	//
	//	omitted     -> computed; the field is not sent, so MAAS keeps whatever it has
	//	[]          -> clear every DNS server on the subnet
	//	[a, b, ...] -> set those DNS servers
	//
	// The omitted cases is required to keep the dns_servers field a computed field, to
	// avoid breaking changes to the provider
	if rawConfig := d.GetRawConfig(); !rawConfig.IsNull() {
		// isNull=False and isKnown=True means the attribute was set in the configuration, even if it was an empty list
		if dnsServers := rawConfig.GetAttr("dns_servers"); !dnsServers.IsNull() && dnsServers.IsKnown() {
			if dnsServers.LengthInt() == 0 {
				// User has specified [], tell MAAS to remove all dns servers by sending a list with empty string
				params.DNSServers = []string{""}
			} else {
				// User has specified a list of DNS servers
				params.DNSServers = convertToStringSlice(d.Get("dns_servers"))
			}
		}
	}

	if p, ok := d.GetOk("fabric"); ok {
		fabric, err := getFabric(client, p.(string))
		if err != nil {
			return nil, err
		}

		params.Fabric = strconv.Itoa(fabric.ID)

		if p, ok := d.GetOk("vlan"); ok {
			vlan, err := getVLAN(client, fabric.ID, p.(string), lookupVLANByVID)
			if err != nil {
				return nil, err
			}

			params.VLAN = strconv.Itoa(vlan.ID)
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
		if strconv.Itoa(s.ID) == identifier || s.CIDR == identifier {
			return &s, nil
		}
	}

	return nil, err
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
