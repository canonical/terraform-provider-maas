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
				Default:  "untagged",
			},
			"mtu": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1500,
			},
			"accept_ra": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"autoconf": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceNetworkInterfacePhysicalCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machineId := d.Get("machine_id").(string)
	networkInterface, err := findNetworkInterfacePhysical(client, d)
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

func findNetworkInterfacePhysical(client *client.Client, d *schema.ResourceData) (*entity.NetworkInterface, error) {
	machineId := d.Get("machine_id").(string)
	networkInterfaces, err := client.NetworkInterfaces.Get(machineId)
	if err != nil {
		return nil, err
	}

	macAddress := d.Get("mac_address").(string)
	for _, n := range networkInterfaces {
		if n.MACAddress == macAddress {
			return &n, nil
		}
	}

	return nil, nil
}
