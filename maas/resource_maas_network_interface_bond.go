package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maas/gomaasclient/client"
	"github.com/maas/gomaasclient/entity"
)

func resourceMaasNetworkInterfaceBond() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage a bond network interface from an existing MAAS machine.",
		CreateContext: resourceNetworkInterfaceBondCreate,
		ReadContext:   resourceNetworkInterfaceBondRead,
		UpdateContext: resourceNetworkInterfaceBondUpdate,
		DeleteContext: resourceNetworkInterfaceBondDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE:NETWORK_INTERFACE", d.Id())
				}
				client := m.(*client.Client)
				machine, err := getMachine(client, idParts[0])
				if err != nil {
					return nil, err
				}
				n, err := getNetworkInterfaceBond(client, machine.SystemID, idParts[1])
				if err != nil {
					return nil, err
				}
				np1, err := findNetworkInterfacePhysical(client, machine.SystemID, n.Parents[0])
				if err != nil {
					return nil, err
				}
				np2, err := findNetworkInterfacePhysical(client, machine.SystemID, n.Parents[1])
				if err != nil {
					return nil, err
				}

				tfState := map[string]interface{}{
					"id":                    fmt.Sprintf("%v", n.ID),
					"machine":               machine.SystemID,
					"parents":               []int{np1.ID, np2.ID},
					"vlan":                  fmt.Sprintf("%v", n.VLAN.ID),
					"name":                  n.Name,
					"mac_address":           n.MACAddress,
					"mtu":                   n.EffectiveMTU,
					"bond_mode":             n.BondMode,
					"bond_xmit_hash_policy": n.BondXMitHashPolicy,
					"bond_miimon":           n.BondMIIMon,
					"bond_downdelay":        n.BondDownDelay,
					"bond_updelay":          n.BondUpDelay,
					"bond_lacp_rate":        n.Params.(map[string]interface{})["bond_lacp_rate"],
					"bond_num_grat_arp":     n.Params.(map[string]interface{})["bond_num_grat_arp"],
					"tags":                  n.Tags,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (system ID, hostname, or FQDN) of the machine with the bond network interface.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Computed:    false,
				Description: "The name of the the bond network interface.",
			},
			"vlan": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "MAAS VLAN ID the interface is connected to.",
			},
			"parents": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Parent interface ids that make this bond.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The MTU of the bond network interface. This argument is computed if it's not set.",
			},
			"bond_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    false,
				Description: "The operating mode of the bond. Supported bonding modes: balance-rr, active-backup, balance-xor, broadcast, 802.3ad, balance-tlb, balance-alb",
				Default:     "active-backup",
			},
			"bond_miimon": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    false,
				Description: "The link monitoring frequency in miliseconds. (Default: 100).",
				Default:     100,
			},
			"bond_updelay": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Specifies the time in miliseconds to wait before enabling a slave after a link recovery has been detected.",
			},
			"bond_downdelay": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Specifies the time in miliseconds to wait before disabling a slave after a link failure has been detected.",
			},
			"bond_lacp_rate": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Option specifying the rate at which to ask the link partner to transmit LACPDU packets in 802.3ad mode. Available options are fast or slow. (Default: slow).",
			},
			"bond_num_grat_arp": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The number of peer notifications (IPv4 ARP or IPv6 Neighbour Advertisements) to be issued after a failover. (Default: 1).",
			},
			"bond_xmit_hash_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    false,
				Description: "The transmit hash policy to use for slave selection in balance-xor, 802.3ad, and tlb modes. Possible values are: layer2, layer2+3, layer3+4, encap2+3, encap3+4. (Default layer2)",
				Default:     "layer2",
			},
			"mac_address": {
				Type:        schema.TypeString,
				Optional:    false,
				Computed:    false,
				Required:    true,
				Description: "MAC address of the interface",
			},
			"tags": {
				Type:        schema.TypeList,
				Required:    false,
				Optional:    true,
				Description: "Tags for the interface.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceNetworkInterfaceBondCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	networkInterface, err := findNetworkInterfaceBond(client, machine.SystemID, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if networkInterface == nil {
		networkInterface, err = client.NetworkInterfaces.CreateBond(machine.SystemID, getNetworkInterfaceBondParams(d))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(fmt.Sprintf("%v", networkInterface.ID))

	return resourceNetworkInterfaceBondUpdate(ctx, d, m)
}

func resourceNetworkInterfaceBondRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

	tfState := map[string]interface{}{
		"name":                  networkInterface.Name,
		"mtu":                   networkInterface.EffectiveMTU,
		"parents":               d.Get("parents"),
		"bond_mode":             networkInterface.Params.(map[string]interface{})["bond_mode"],
		"bond_xmit_hash_policy": networkInterface.Params.(map[string]interface{})["bond_xmit_hash_policy"],
		"bond_miimon":           networkInterface.Params.(map[string]interface{})["bond_miimon"],
		"bond_updelay":          networkInterface.BondUpDelay,
		"bond_downdelay":        networkInterface.BondDownDelay,
		"bond_lacp_rate":        networkInterface.Params.(map[string]interface{})["bond_lacp_rate"],
		"bond_num_grat_arp":     networkInterface.Params.(map[string]interface{})["bond_num_grat_arp"],
		"mac_address":           networkInterface.MACAddress,
		"vlan":                  fmt.Sprintf("%v", networkInterface.VLAN.ID),
		"tags":                  networkInterface.Tags,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfaceBondUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = client.NetworkInterface.Update(machine.SystemID, id, getNetworkInterfaceBondParams(d)); err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkInterfaceBondRead(ctx, d, m)
}

func resourceNetworkInterfaceBondDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

func getNetworkInterfaceBondParams(d *schema.ResourceData) *entity.NetworkInterfaceBondParams {
	parents := make([]int, 0, 2)
	for _, v := range d.Get("parents").([]interface{}) {
		parents = append(parents, v.(int))
	}
	tags := make([]string, 0, len(d.Get("tags").([]interface{})))
	for _, v := range d.Get("tags").([]interface{}) {
		tags = append(tags, v.(string))
	}
	return &entity.NetworkInterfaceBondParams{
		NetworkInterfacePhysicalParams: entity.NetworkInterfacePhysicalParams{
			Name:       d.Get("name").(string),
			VLAN:       d.Get("vlan").(string),
			MACAddress: d.Get("mac_address").(string),
			MTU:        d.Get("mtu").(int),
			Tags:       strings.Join(tags, ","),
		},
		Parents:            parents,
		BondMode:           d.Get("bond_mode").(string),
		BondMiimon:         d.Get("bond_miimon").(int),
		BondDownDelay:      d.Get("bond_downdelay").(int),
		BondUpDelay:        d.Get("bond_updelay").(int),
		BondLACPRate:       d.Get("bond_lacp_rate").(string),
		BondXMitHashPolicy: d.Get("bond_xmit_hash_policy").(string),
		BondNumberGratARP:  d.Get("bond_num_grat_arp").(int),
	}
}

func findNetworkInterfaceBond(client *client.Client, machineSystemID string, identifier string) (*entity.NetworkInterface, error) {
	networkInterfaces, err := client.NetworkInterfaces.Get(machineSystemID)
	if err != nil {
		return nil, err
	}
	for _, n := range networkInterfaces {
		if n.Type != "bond" {
			continue
		}
		if n.Name == identifier || fmt.Sprintf("%v", n.ID) == identifier {
			return &n, nil
		}
	}
	return nil, nil
}

func getNetworkInterfaceBond(client *client.Client, machineSystemID string, identifier string) (*entity.NetworkInterface, error) {
	n, err := findNetworkInterfaceBond(client, machineSystemID, identifier)
	if err != nil {
		return nil, err
	}
	if n != nil {
		return n, nil
	}
	return nil, fmt.Errorf("bond network interface (%s) was not found on machine (%s)", identifier, machineSystemID)
}
