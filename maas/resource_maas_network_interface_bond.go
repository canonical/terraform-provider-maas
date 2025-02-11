package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasNetworkInterfaceBond() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS network Bonds.",
		CreateContext: resourceNetworkInterfaceBondCreate,
		ReadContext:   resourceNetworkInterfaceBondRead,
		UpdateContext: resourceNetworkInterfaceBondUpdate,
		DeleteContext: resourceNetworkInterfaceBondDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNetworkInterfaceBondImport,
		},
		Schema: map[string]*schema.Schema{
			"accept_ra": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Accept router advertisements. (IPv6 only).",
			},
			"bond_downdelay": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Specifies the time, in milliseconds, to wait before disabling a slave after a link failure has been detected.",
			},
			"bond_lacp_rate": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Option specifying the rate at which to ask the link partner to transmit LACPDU packets in 802.3ad mode. Available options are ``fast`` or ``slow``. (Default: ``slow``).",
			},
			"bond_miimon": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The link monitoring frequency in milliseconds. (Default: 100).",
			},
			"bond_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The operating mode of the bond. (Default: active-backup). Supported bonding modes: - ``balance-rr``: Transmit packets in sequential order from the first available slave through the last. This mode provides load balancing and fault tolerance. - ``active-backup``: Only one slave in the bond is active. A different slave becomes active if, and only if, the active slave fails. The bond's MAC address is externally visible on only one port (network adapter) to avoid confusing the switch. - ``balance-xor``: Transmit based on the selected transmit hash policy. The default policy is a simple [(source MAC address XOR'd with destination MAC address XOR packet type ID) modulo slave count]. - ``broadcast``: Transmits everything on all slave interfaces. This mode provides fault tolerance. - ``802.3ad``: IEEE 802.3ad dynamic link aggregation. Creates aggregation groups that share the same speed and duplex settings. Uses all slaves in the active aggregator according to the 802.3ad specification. - ``balance-tlb``: Adaptive transmit load balancing: channel bonding that does not require any special switch support. - ``balance-alb``: Adaptive load balancing: includes balance-tlb plus receive load balancing (rlb) for IPV4 traffic, and does not require any special switch support. The receive load balancing is achieved by ARP negotiation.",
			},
			"bond_num_grat_arp": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The number of peer notifications (IPv4 ARP or IPv6 Neighbour Advertisements) to be issued after a failover. (Default: 1).",
			},
			"bond_updelay": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Specifies the time, in milliseconds, to wait before enabling a slave after a link recovery has been detected.",
			},
			"bond_xmit_hash_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The transmit hash policy to use for slave selection in balance-xor, 802.3ad, and tlb modes. Possible values are: ``layer2``, ``layer2+3``, ``layer3+4``, ``encap2+3``, ``encap3+4``. (Default: ``layer2``).",
			},
			"mac_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The bond interface MAC address.",
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (system ID, hostname, or FQDN) of the machine with the bond interface.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The MTU of the bond interface.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The bond interface name.",
			},
			"parents": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Parent interface names for this bond interface.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A set of tag names to be assigned to the bond interface.",
			},
			"vlan": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Database ID of the VLAN the bond interface is connected to.",
			},
		},
	}
}

func resourceNetworkInterfaceBondCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	p, err := findBondParentsID(client, machine.SystemID, d.Get("parents").(*schema.Set).List())
	if err != nil {
		return diag.FromErr(err)
	}

	params := getNetworkInterfaceBondParams(d, p)
	networkInterface, err := client.NetworkInterfaces.CreateBond(machine.SystemID, params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(networkInterface.ID))

	return resourceNetworkInterfaceBondRead(ctx, d, meta)
}

func resourceNetworkInterfaceBondRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	networkInterface, err := client.NetworkInterface.Get(machine.SystemID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	p := networkInterface.Params.(map[string]interface{})
	if _, ok := p["accept-ra"]; ok {
		d.Set("accept_ra", p["accept-ra"].(bool))
	} else {
		d.Set("accept_ra", false)
	}
	if _, ok := p["bond_downdelay"]; ok {
		d.Set("bond_downdelay", int64(p["bond_downdelay"].(float64)))
	}
	if _, ok := p["bond_lacp_rate"]; ok {
		d.Set("bond_lacp_rate", p["bond_lacp_rate"].(string))
	}
	if _, ok := p["bond_miimon"]; ok {
		d.Set("bond_miimon", int64(p["bond_miimon"].(float64)))
	}
	if _, ok := p["bond_mode"]; ok {
		d.Set("bond_mode", p["bond_mode"].(string))
	}
	if _, ok := p["bond_num_grat_arp"]; ok {
		d.Set("bond_num_grat_arp", int64(p["bond_num_grat_arp"].(float64)))
	}
	if _, ok := p["bond_updelay"]; ok {
		d.Set("bond_updelay", int64(p["bond_updelay"].(float64)))
	}
	if _, ok := p["bond_xmit_hash_policy"]; ok {
		d.Set("bond_xmit_hash_policy", p["bond_xmit_hash_policy"].(string))
	}

	tfState := map[string]interface{}{
		"mac_address": networkInterface.MACAddress,
		"mtu":         networkInterface.EffectiveMTU,
		"name":        networkInterface.Name,
		"parents":     networkInterface.Parents,
		"tags":        networkInterface.Tags,
		"vlan":        networkInterface.VLAN.ID,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfaceBondUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	p, err := findBondParentsID(client, machine.SystemID, d.Get("parents").(*schema.Set).List())
	if err != nil {
		return diag.FromErr(err)
	}

	params := getNetworkInterfaceBondUpdateParams(d, p)
	_, err = client.NetworkInterface.Update(machine.SystemID, id, params)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkInterfaceBondRead(ctx, d, meta)
}

func resourceNetworkInterfaceBondDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.NetworkInterface.Delete(machine.SystemID, id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getNetworkInterfaceBondParams(d *schema.ResourceData, parentIDs []int) *entity.NetworkInterfaceBondParams {
	return &entity.NetworkInterfaceBondParams{
		AcceptRA:           d.Get("accept_ra").(bool),
		BondDownDelay:      d.Get("bond_downdelay").(int),
		BondLACPRate:       d.Get("bond_lacp_rate").(string),
		BondMiimon:         d.Get("bond_miimon").(int),
		BondMode:           d.Get("bond_mode").(string),
		BondNumberGratARP:  d.Get("bond_num_grat_arp").(int),
		BondUpDelay:        d.Get("bond_updelay").(int),
		BondXMitHashPolicy: d.Get("bond_xmit_hash_policy").(string),
		MACAddress:         d.Get("mac_address").(string),
		MTU:                d.Get("mtu").(int),
		Name:               d.Get("name").(string),
		Parents:            parentIDs,
		Tags:               strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		VLAN:               d.Get("vlan").(int),
	}
}

func getNetworkInterfaceBondUpdateParams(d *schema.ResourceData, parentIDs []int) *entity.NetworkInterfaceUpdateParams {

	return &entity.NetworkInterfaceUpdateParams{
		AcceptRA:           d.Get("accept_ra").(bool),
		BondDownDelay:      d.Get("bond_downdelay").(int),
		BondLACPRate:       d.Get("bond_lacp_rate").(string),
		BondMiimon:         d.Get("bond_miimon").(int),
		BondMode:           d.Get("bond_mode").(string),
		BondNumberGratARP:  d.Get("bond_num_grat_arp").(int),
		BondUpDelay:        d.Get("bond_updelay").(int),
		BondXMitHashPolicy: d.Get("bond_xmit_hash_policy").(string),
		MACAddress:         d.Get("mac_address").(string),
		MTU:                d.Get("mtu").(int),
		Name:               d.Get("name").(string),
		Parents:            parentIDs,
		Tags:               strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		VLAN:               d.Get("vlan").(int),
	}
}

func findBondParentsID(client *client.Client, machineSystemID string, parents []interface{}) ([]int, error) {
	var result []int
	for _, p := range parents {
		if p, ok := p.(string); ok {
			networkInterface, err := getNetworkInterface(client, machineSystemID, p)
			if err != nil {
				return nil, err
			}
			if networkInterface.Type != "physical" {
				continue
			}
			result = append(result, networkInterface.ID)
		}
	}

	return result, nil
}

func resourceNetworkInterfaceBondImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE:BOND_INTERFACE_ID", d.Id())
	}

	d.Set("machine", idParts[0])
	d.SetId(idParts[1])

	return []*schema.ResourceData{d}, nil
}
