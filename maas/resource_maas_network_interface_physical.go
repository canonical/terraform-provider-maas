package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasNetworkInterfacePhysical() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkInterfacePhysicalCreate,
		ReadContext:   resourceNetworkInterfacePhysicalRead,
		UpdateContext: resourceNetworkInterfacePhysicalUpdate,
		DeleteContext: resourceNetworkInterfacePhysicalDelete,
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
			"machine": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vlan": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"mtu": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceNetworkInterfacePhysicalCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

	return resourceNetworkInterfacePhysicalUpdate(ctx, d, m)
}

func resourceNetworkInterfacePhysicalRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		"tags": networkInterface.Tags,
		"mtu":  networkInterface.EffectiveMTU,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfacePhysicalUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

	return resourceNetworkInterfacePhysicalRead(ctx, d, m)
}

func resourceNetworkInterfacePhysicalDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
