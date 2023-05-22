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

func resourceMaasNetworkInterfaceBridge() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage a bridge network interface from an existing MAAS machine.",
		CreateContext: resourceNetworkInterfaceBridgeCreate,
		ReadContext:   resourceNetworkInterfaceBridgeRead,
		UpdateContext: resourceNetworkInterfaceBridgeUpdate,
		DeleteContext: resourceNetworkInterfaceBridgeDelete,
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
				n, err := getNetworkInterfaceBridge(client, machine.SystemID, idParts[1])
				if err != nil {
					return nil, err
				}
				np, err := findNetworkInterfacePhysical(client, machine.SystemID, n.Parents[0])
				if err != nil {
					return nil, err
				}
				if np == nil {
                                        np, err = findNetworkInterfaceVlan(client, machine.SystemID, n.Parents[0])
                                        if np == nil {
					    return nil, fmt.Errorf("Parent interface (%s) not found", n.Parents[0])
                                        }
                                        if err != nil {
                                            return nil, err
                                }

				}
				tfState := map[string]interface{}{
					"id":          fmt.Sprintf("%v", n.ID),
					"machine":     machine.SystemID,
					"parent":      np.ID,
					"vlan":        fmt.Sprintf("%v", n.VLAN.ID),
					"name":        n.Name,
					"mac_address":        n.MACAddress,
					"mtu"	     : n.EffectiveMTU,
					"bridge_fd"  : n.BridgeFD,
					"bridge_stp"  : n.BridgeSTP,
					"bridge_type" : n.Params.(map[string]interface{})["bridge_type"],
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
				Description: "The identifier (system ID, hostname, or FQDN) of the machine with the bridge network interface.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Computed:    false,
				Description: "The name of the the bridge network interface.",
			},
			"vlan": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "MAAS VLAN ID the interface is connected to.",
			},
			"parent": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Parent interface ID for this bridge interface.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The MTU of the bridge network interface. This argument is computed if it's not set.",
			},
			"bridge_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    false,
				Description: "The type of bridge to create. Possible values are: standard, ovs.",
				Default:     "standard",
			},
			"bridge_fd": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Set bridge forward delay to time seconds. (Default: 15).",
			},
			"bridge_stp": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Turn spanning tree protocol on or off. (Default: False).",
			},
			"mac_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "MAC address of the interface",
			},
		},
	}
}

func resourceNetworkInterfaceBridgeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	networkInterface, err := findNetworkInterfaceBridge(client, machine.SystemID, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if networkInterface == nil {
		networkInterface, err = client.NetworkInterfaces.CreateBridge(machine.SystemID, getNetworkInterfaceBridgeParams(d))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(fmt.Sprintf("%v", networkInterface.ID))

	return resourceNetworkInterfaceBridgeUpdate(ctx, d, m)
}

func resourceNetworkInterfaceBridgeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		"name": networkInterface.Name,
		"mtu":  networkInterface.EffectiveMTU,
		"parent":  d.Get("parent"),
		"bridge_type": d.Get("bridge_type"),
		"bridge_stp": networkInterface.BridgeSTP,
		"bridge_fd": networkInterface.BridgeFD,
		"mac_address": networkInterface.MACAddress,
		"vlan": fmt.Sprintf("%v", networkInterface.VLAN.ID),
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfaceBridgeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = client.NetworkInterface.Update(machine.SystemID, id, getNetworkInterfaceBridgeParams(d)); err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkInterfaceBridgeRead(ctx, d, m)
}

func resourceNetworkInterfaceBridgeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func getNetworkInterfaceBridgeParams(d *schema.ResourceData) *entity.NetworkInterfaceBridgeParams {
	return &entity.NetworkInterfaceBridgeParams{
		NetworkInterfacePhysicalParams: entity.NetworkInterfacePhysicalParams{
			Name:           d.Get("name").(string),
			VLAN:           d.Get("vlan").(string),
			MACAddress:	d.Get("mac_address").(string),
			MTU:		d.Get("mtu").(int),
		},
		Parent:		  d.Get("parent").(int),
		Bridgetype:       d.Get("bridge_type").(string),
		BridgeSTP:       d.Get("bridge_stp").(bool),
		BridgeFD:       d.Get("bridge_fd").(int),
	}
}

func findNetworkInterfaceBridge(client *client.Client, machineSystemID string, identifier string) (*entity.NetworkInterface, error) {
	networkInterfaces, err := client.NetworkInterfaces.Get(machineSystemID)
	if err != nil {
		return nil, err
	}
	for _, n := range networkInterfaces {
		if n.Type != "bridge" {
			continue
		}
		if n.Name == identifier || fmt.Sprintf("%v", n.ID) == identifier {
			return &n, nil
		}
	}
	return nil, nil
}

func getNetworkInterfaceBridge(client *client.Client, machineSystemID string, identifier string) (*entity.NetworkInterface, error) {
	n, err := findNetworkInterfaceBridge(client, machineSystemID, identifier)
	if err != nil {
		return nil, err
	}
	if n != nil {
		return n, nil
	}
	return nil, fmt.Errorf("bridge network interface (%s) was not found on machine (%s)", identifier, machineSystemID)
}
