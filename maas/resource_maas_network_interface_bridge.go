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
		Description:   "Provides a resource to manage MAAS network Bridges.",
		CreateContext: resourceMaasNetworkInterfaceBridgeCreate,
		ReadContext:   resourceMaasNetworkInterfaceBridgeRead,
		UpdateContext: resourceMaasNetworkInterfaceBridgeUpdate,
		DeleteContext: resourceMaasNetworkInterfaceBridgeDelete,
		Importer: &schema.ResourceImporter{
			State: resourceMaasNetworkInterfaceBridgeImport,
		},
		Schema: map[string]*schema.Schema{
			"accept_ra": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Accept router advertisements. (IPv6 only).",
			},
			"bridge_fd": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Set bridge forward delay to time seconds. (Default: 15).",
			},
			"bridge_stp": {
				Type:     schema.TypeBool,
				Optional: true,
				// Computed:    true,
				Description: "Turn spanning tree protocol on or off. (Default: False).",
			},
			"bridge_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The type of bridge to create. Possible values are: ``standard``, ``ovs``.",
			},
			"mac_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "MAC address of the interface.",
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "List of MAAS machines' identifiers (system ID, hostname, or FQDN) that will be tagged with the new tag.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum transmission unit.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Name of the interface.",
			},
			"parent": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Parent interface name for this bridge interface.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Tags for the interface.",
			},
			"vlan": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "VLAN the interface is connected to.",
			},
		},
	}
}

func resourceMaasNetworkInterfaceBridgeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	parentID, err := findInterfaceParent(client, machine.SystemID, d.Get("parent").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	params := getNetworkInterfaceBridgeParams(d, parentID)
	networkInterface, err := client.NetworkInterfaces.CreateBridge(machine.SystemID, params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(networkInterface.ID))

	return resourceMaasNetworkInterfaceBridgeRead(ctx, d, meta)
}

func resourceMaasNetworkInterfaceBridgeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	p := networkInterface.Params.(map[string]interface{})
	if _, ok := p["accept-ra"]; ok {
		d.Set("accept_ra", p["accept-ra"].(bool))
	}
	if _, ok := p["bridge_fd"]; ok {
		d.Set("bridge_fd", int64(p["bridge_fd"].(float64)))
	}
	if _, ok := p["bridge_stp"]; ok {
		d.Set("bridge_stp", p["bridge_stp"].(bool))
	}
	if _, ok := p["bridge_type"]; ok {
		d.Set("bridge_type", p["bridge_type"].(string))
	}

	tfState := map[string]interface{}{
		"mac_address": networkInterface.MACAddress,
		"mtu":         networkInterface.EffectiveMTU,
		"name":        networkInterface.Name,
		"tags":        networkInterface.Tags,
		"vlan":        fmt.Sprintf("%v", networkInterface.VLAN.ID),
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMaasNetworkInterfaceBridgeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	parentID, err := findInterfaceParent(client, machine.SystemID, d.Get("parent").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	params := getNetworkInterfaceBridgeUpdateParams(d, parentID)
	_, err = client.NetworkInterface.Update(machine.SystemID, id, params)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceMaasNetworkInterfaceBridgeRead(ctx, d, meta)
}

func resourceMaasNetworkInterfaceBridgeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func getNetworkInterfaceBridgeParams(d *schema.ResourceData, parentID int) *entity.NetworkInterfaceBridgeParams {
	return &entity.NetworkInterfaceBridgeParams{
		AcceptRA:   d.Get("accept_ra").(bool),
		BridgeType: d.Get("bridge_type").(string),
		BridgeSTP:  d.Get("bridge_stp").(bool),
		BridgeFD:   d.Get("bridge_fd").(int),
		MACAddress: d.Get("mac_address").(string),
		MTU:        d.Get("mtu").(int),
		Name:       d.Get("name").(string),
		Parents:    []int{parentID},
		Tags:       strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		VLAN:       d.Get("vlan").(int),
	}
}

func getNetworkInterfaceBridgeUpdateParams(d *schema.ResourceData, parentID int) *entity.NetworkInterfaceUpdateParams {
	return &entity.NetworkInterfaceUpdateParams{
		AcceptRA:   d.Get("accept_ra").(bool),
		BridgeType: d.Get("bridge_type").(string),
		BridgeSTP:  d.Get("bridge_stp").(bool),
		BridgeFD:   d.Get("bridge_fd").(int),
		MACAddress: d.Get("mac_address").(string),
		MTU:        d.Get("mtu").(int),
		Name:       d.Get("name").(string),
		Parents:    []int{parentID},
		Tags:       strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		VLAN:       d.Get("vlan").(int),
	}
}

func findInterfaceParent(client *client.Client, machineSystemID string, parent string) (int, error) {
	networkInterface, err := getNetworkInterface(client, machineSystemID, parent)
	if err != nil {
		return 0, err
	}

	return networkInterface.ID, nil
}

func resourceMaasNetworkInterfaceBridgeImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE:BRIDGE_ID", d.Id())
	}

	d.Set("machine", idParts[0])
	d.SetId(idParts[1])

	return []*schema.ResourceData{d}, nil
}
