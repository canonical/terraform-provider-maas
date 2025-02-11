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

func resourceMaasNetworkInterfacePhysical() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage a physical network interface from an existing MAAS machine.",
		CreateContext: resourceNetworkInterfacePhysicalCreate,
		ReadContext:   resourceNetworkInterfacePhysicalRead,
		UpdateContext: resourceNetworkInterfacePhysicalUpdate,
		DeleteContext: resourceNetworkInterfacePhysicalDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE/NETWORK_INTERFACE", d.Id())
				}
				client := meta.(*ClientConfig).Client

				machine, err := getMachine(client, idParts[0])
				if err != nil {
					return nil, err
				}
				n, err := getNetworkInterfacePhysical(client, machine.SystemID, idParts[1])
				if err != nil {
					return nil, err
				}
				d.Set("machine", idParts[0])
				d.SetId(strconv.Itoa(n.ID))
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
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Database ID of the VLAN the physical network interface is connected to.",
			},
		},
	}
}

func resourceNetworkInterfacePhysicalCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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
	} else {
		networkInterface, err = client.NetworkInterface.Update(machine.SystemID, networkInterface.ID, getNetworkInterfaceUpdateParams(d))
	}
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.Itoa(networkInterface.ID))

	tfState := map[string]interface{}{
		"mac_address": networkInterface.MACAddress,
		"mtu":         networkInterface.EffectiveMTU,
		"name":        networkInterface.Name,
		"tags":        networkInterface.Tags,
		"vlan":        networkInterface.VLAN.ID,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfacePhysicalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	tfState := map[string]interface{}{
		"mac_address": networkInterface.MACAddress,
		"mtu":         networkInterface.EffectiveMTU,
		"name":        networkInterface.Name,
		"tags":        networkInterface.Tags,
		"vlan":        networkInterface.VLAN.ID,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfacePhysicalUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	networkInterface, err := client.NetworkInterface.Update(machine.SystemID, id, getNetworkInterfaceUpdateParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	tfState := map[string]interface{}{
		"mac_address": networkInterface.MACAddress,
		"mtu":         networkInterface.EffectiveMTU,
		"name":        networkInterface.Name,
		"tags":        networkInterface.Tags,
		"vlan":        networkInterface.VLAN.ID,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfacePhysicalDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func getNetworkInterfacePhysicalParams(d *schema.ResourceData) *entity.NetworkInterfacePhysicalParams {
	return &entity.NetworkInterfacePhysicalParams{
		MACAddress: d.Get("mac_address").(string),
		MTU:        d.Get("mtu").(int),
		Name:       d.Get("name").(string),
		Tags:       strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		VLAN:       d.Get("vlan").(int),
	}
}

func getNetworkInterfaceUpdateParams(d *schema.ResourceData) *entity.NetworkInterfaceUpdateParams {
	return &entity.NetworkInterfaceUpdateParams{
		MACAddress: d.Get("mac_address").(string),
		MTU:        d.Get("mtu").(int),
		Name:       d.Get("name").(string),
		Tags:       strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		VLAN:       d.Get("vlan").(int),
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
		if n.MACAddress == identifier || n.Name == identifier || strconv.Itoa(n.ID) == identifier {
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
