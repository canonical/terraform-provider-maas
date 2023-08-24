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

func resourceMaasNetworkInterfaceVlan() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage a vlan network interface from an existing MAAS machine.",
		CreateContext: resourceNetworkInterfaceVlanCreate,
		ReadContext:   resourceNetworkInterfaceVlanRead,
		UpdateContext: resourceNetworkInterfaceVlanUpdate,
		DeleteContext: resourceNetworkInterfaceVlanDelete,
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
				n, err := getNetworkInterfaceVlan(client, machine.SystemID, idParts[1])
				if err != nil {
					return nil, err
				}
				ifParts := strings.Split(idParts[1], ".")
				if strings.Contains(ifParts[0], "bond") {
					np, err := getNetworkInterfaceBond(client, machine.SystemID, ifParts[0])
					if err != nil {
						return nil, err
					}

					tfState := map[string]interface{}{
						"id":      fmt.Sprintf("%v", n.ID),
						"machine": machine.SystemID,
						"parent":  np.ID,
						"vlan":    fmt.Sprintf("%v", n.VLAN.ID),
						"tags":    n.Tags,
						"name":    n.Name,
					}
					if err := setTerraformState(d, tfState); err != nil {
						return nil, err
					}
					return []*schema.ResourceData{d}, nil
				} else {
					np, err := getNetworkInterfacePhysical(client, machine.SystemID, ifParts[0])
					if err != nil {
						return nil, err
					}

					tfState := map[string]interface{}{
						"id":      fmt.Sprintf("%v", n.ID),
						"machine": machine.SystemID,
						"parent":  np.ID,
						"vlan":    fmt.Sprintf("%v", n.VLAN.ID),
						"tags":    n.Tags,
						"name":    n.Name,
					}
					if err := setTerraformState(d, tfState); err != nil {
						return nil, err
					}
					return []*schema.ResourceData{d}, nil

				}

			},
		},

		Schema: map[string]*schema.Schema{
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (system ID, hostname, or FQDN) of the machine with the vlan network interface.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "The name of the the vlan network interface.",
			},
			"vlan": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "MAAS VLAN ID the interface is connected to.",
			},
			"parent": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Parent interface ID for this vlan interface.",
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tag names to be assigned to the vlan network interface. This argument is computed if it's not set.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The MTU of the vlan network interface. This argument is computed if it's not set.",
			},
		},
	}
}

func resourceNetworkInterfaceVlanCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	np, err := getNetworkInterfacePhysical(client, d.Get("machine").(string), fmt.Sprintf("%v", d.Get("parent").(int)))
	if err != nil {
		np, err = getNetworkInterfaceBond(client, d.Get("machine").(string), fmt.Sprintf("%v", d.Get("parent").(int)))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	networkInterface, err := findNetworkInterfaceVlan(client, machine.SystemID, np.Name+"."+d.Get("vlan").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if networkInterface == nil {
		networkInterface, err = client.NetworkInterfaces.CreateVLAN(machine.SystemID, getNetworkInterfaceVlanParams(d))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(fmt.Sprintf("%v", networkInterface.ID))

	return resourceNetworkInterfaceVlanUpdate(ctx, d, m)
}

func resourceNetworkInterfaceVlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		"name":   networkInterface.Name,
		"tags":   networkInterface.Tags,
		"mtu":    networkInterface.EffectiveMTU,
		"parent": d.Get("parent"),
		"vlan":   fmt.Sprintf("%v", networkInterface.VLAN.ID),
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfaceVlanUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = client.NetworkInterface.Update(machine.SystemID, id, getNetworkInterfaceVlanParams(d)); err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkInterfaceVlanRead(ctx, d, m)
}

func resourceNetworkInterfaceVlanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func getNetworkInterfaceVlanParams(d *schema.ResourceData) *entity.NetworkInterfaceVLANParams {
	return &entity.NetworkInterfaceVLANParams{
		Parent: d.Get("parent").(int),
		VLAN:   d.Get("vlan").(string),
		MTU:    d.Get("mtu").(int),
		Tags:   []string{strings.Join(convertToStringSlice(d.Get("tags").([]interface{})), ",")},
	}
}

func findNetworkInterfaceVlan(client *client.Client, machineSystemID string, identifier string) (*entity.NetworkInterface, error) {
	networkInterfaces, err := client.NetworkInterfaces.Get(machineSystemID)
	if err != nil {
		return nil, err
	}
	for _, n := range networkInterfaces {
		if n.Type != "vlan" {
			continue
		}
		if fmt.Sprintf("%v", n.ID) == identifier || n.Name == identifier {
			return &n, nil
		}
	}
	return nil, nil
}

func getNetworkInterfaceVlan(client *client.Client, machineSystemID string, identifier string) (*entity.NetworkInterface, error) {
	n, err := findNetworkInterfaceVlan(client, machineSystemID, identifier)
	if err != nil {
		return nil, err
	}
	if n != nil {
		return n, nil
	}
	return nil, fmt.Errorf("vlan network interface (%s) was not found on machine (%s)", identifier, machineSystemID)
}
