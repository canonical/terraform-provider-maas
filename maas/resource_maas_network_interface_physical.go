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

func resourceMaasNetworkInterfacePhysical() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage a physical network interface from an existing MAAS machine.",
		CreateContext: resourceNetworkInterfacePhysicalCreate,
		ReadContext:   resourceNetworkInterfacePhysicalRead,
		UpdateContext: resourceNetworkInterfacePhysicalUpdate,
		DeleteContext: resourceNetworkInterfacePhysicalDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE:NETWORK_INTERFACE", d.Id())
				}
				client := meta.(*client.Client)
				machine, err := getMachine(client, idParts[0])
				if err != nil {
					return nil, err
				}
				n, err := getNetworkInterfacePhysical(client, machine.SystemID, idParts[1])
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
					"id":          fmt.Sprintf("%v", n.ID),
					"machine":     machine.SystemID,
					"mac_address": n.MACAddress,
					"vlan":        fmt.Sprintf("%v", n.VLAN.ID),
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"mac_address": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The physical network interface MAC address.",
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (system ID, hostname, or FQDN) of the machine with the physical network interface.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The MTU of the physical network interface. This argument is computed if it's not set.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The physical network interface name. This argument is computed if it's not set.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A set of tag names to be assigned to the physical network interface. This argument is computed if it's not set.",
			},
			"vlan": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "VLAN the physical network interface is connected to. Defaults to `untagged`.",
			},
		},
	}
}

func resourceNetworkInterfacePhysicalCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	networkInterface, err := findNetworkInterfacePhysical(client, machine.SystemID, d.Get("mac_address").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if networkInterface == nil {
		networkInterface, err = client.NetworkInterfaces.CreatePhysical(machine.SystemID, getNetworkInterfacePhysicalParams(d))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(fmt.Sprintf("%v", networkInterface.ID))

	return resourceNetworkInterfacePhysicalUpdate(ctx, d, meta)
}

func resourceNetworkInterfacePhysicalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

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
		"tags": networkInterface.Tags,
		"mtu":  networkInterface.EffectiveMTU,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfacePhysicalUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = client.NetworkInterface.Update(machine.SystemID, id, getNetworkInterfacePhysicalParams(d)); err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkInterfacePhysicalRead(ctx, d, meta)
}

func resourceNetworkInterfacePhysicalDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

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

func getNetworkInterfacePhysicalParams(d *schema.ResourceData) *entity.NetworkInterfacePhysicalParams {
	return &entity.NetworkInterfacePhysicalParams{
		MACAddress: d.Get("mac_address").(string),
		VLAN:       d.Get("vlan").(string),
		Name:       d.Get("name").(string),
		MTU:        d.Get("mtu").(int),
		Tags:       strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
	}
}

func findNetworkInterfacePhysical(client *client.Client, machineSystemID string, identifier string) (*entity.NetworkInterface, error) {
	networkInterfaces, err := client.NetworkInterfaces.Get(machineSystemID)
	if err != nil {
		return nil, err
	}
	for _, n := range networkInterfaces {
		if n.Type != "physical" {
			continue
		}
		if n.MACAddress == identifier || n.Name == identifier || fmt.Sprintf("%v", n.ID) == identifier {
			return &n, nil
		}
	}
	return nil, nil
}

func getNetworkInterfacePhysical(client *client.Client, machineSystemID string, identifier string) (*entity.NetworkInterface, error) {
	n, err := findNetworkInterfacePhysical(client, machineSystemID, identifier)
	if err != nil {
		return nil, err
	}
	if n != nil {
		return n, nil
	}
	return nil, fmt.Errorf("physical network interface (%s) was not found on machine (%s)", identifier, machineSystemID)
}
