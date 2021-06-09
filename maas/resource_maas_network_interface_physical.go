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

var (
	defaultVLAN     = "untagged"
	defaultMTU      = 1500
	defaultAcceptRA = false
	defaultAutoconf = false
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
				machine, err := findMachine(client, idParts[0])
				if err != nil {
					return nil, err
				}
				networkInterface, err := findNetworkInterfacePhysical(client, machine.SystemID, idParts[1])
				if err != nil {
					return nil, err
				}
				if networkInterface == nil {
					return nil, fmt.Errorf("physical network interface (%s) was not found on machine (%s)", idParts[1], machine.Hostname)
				}
				if err := d.Set("machine_id", machine.SystemID); err != nil {
					return nil, err
				}
				if err := d.Set("mac_address", networkInterface.MACAddress); err != nil {
					return nil, err
				}
				if err := d.Set("vlan", defaultVLAN); err != nil {
					return nil, err
				}
				if err := d.Set("mtu", defaultMTU); err != nil {
					return nil, err
				}
				if err := d.Set("accept_ra", defaultAcceptRA); err != nil {
					return nil, err
				}
				if err := d.Set("autoconf", defaultAutoconf); err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("%v", networkInterface.ID))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"machine_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vlan": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  defaultVLAN,
			},
			"mtu": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  defaultMTU,
			},
			"accept_ra": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  defaultAcceptRA,
			},
			"autoconf": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  defaultAutoconf,
			},
		},
	}
}

func resourceNetworkInterfacePhysicalCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machineId := d.Get("machine_id").(string)
	networkInterface, err := findNetworkInterfacePhysical(client, d.Get("machine_id").(string), d.Get("mac_address").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if networkInterface == nil {
		networkInterface, err = client.NetworkInterfaces.CreatePhysical(machineId, getNetworkInterfacePhysicalParams(d))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(fmt.Sprintf("%v", networkInterface.ID))

	return resourceNetworkInterfacePhysicalUpdate(ctx, d, m)
}

func resourceNetworkInterfacePhysicalRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machineId := d.Get("machine_id").(string)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	networkInterface, err := client.NetworkInterface.Get(machineId, id)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", networkInterface.Name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfacePhysicalUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machineId := d.Get("machine_id").(string)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.NetworkInterface.Update(machineId, id, getNetworkInterfacePhysicalParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkInterfacePhysicalRead(ctx, d, m)
}

func resourceNetworkInterfacePhysicalDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machineId := d.Get("machine_id").(string)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.NetworkInterface.Delete(machineId, id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getNetworkInterfacePhysicalParams(d *schema.ResourceData) *entity.NetworkInterfacePhysicalParams {
	params := entity.NetworkInterfacePhysicalParams{
		MACAddress: d.Get("mac_address").(string),
		VLAN:       d.Get("vlan").(string),
		MTU:        d.Get("mtu").(int),
		AcceptRA:   d.Get("accept_ra").(bool),
		Autoconf:   d.Get("autoconf").(bool),
	}

	if p, ok := d.GetOk("name"); ok {
		params.Name = p.(string)
	}
	if p, ok := d.GetOk("tags"); ok {
		params.Tags = strings.Join(convertToStringSlice(p.(*schema.Set).List()), ",")
	}

	return &params
}

func findNetworkInterfacePhysical(client *client.Client, machineId string, identifier string) (*entity.NetworkInterface, error) {
	networkInterfaces, err := client.NetworkInterfaces.Get(machineId)
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
