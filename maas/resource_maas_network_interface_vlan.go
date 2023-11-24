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
		Description:   "Provides a resource to manage MAAS network Vlans.",
		CreateContext: resourceMaasNetworkInterfaceVlanCreate,
		ReadContext:   resourceMaasNetworkInterfaceVlanRead,
		UpdateContext: resourceMaasNetworkInterfaceVlanUpdate,
		DeleteContext: resourceMaasNetworkInterfaceVlanDelete,
		Importer: &schema.ResourceImporter{
			State: resourceMaasNetworkInterfaceVlanImport,
		},
		Schema: map[string]*schema.Schema{
			"accept_ra": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Accept router advertisements. (IPv6 only).",
			},
			"fabric": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (name or ID) of the fabric for the new VLAN.",
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
			"parent": {
				Type:        schema.TypeString,
				Required:    true,
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

func resourceMaasNetworkInterfaceVlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	parentID, err := findInterfaceParent(client, machine.SystemID, d.Get("parent").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	vlan, err := getVlan(client, fabric.ID, d.Get("vlan").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	params := getNetworkInterfaceVlanParams(d, parentID, vlan.ID)
	networkInterface, err := client.NetworkInterfaces.CreateVLAN(machine.SystemID, params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(networkInterface.ID))

	return resourceMaasNetworkInterfaceVlanRead(ctx, d, meta)

}

func resourceMaasNetworkInterfaceVlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	tfState := map[string]interface{}{
		"mtu":    networkInterface.EffectiveMTU,
		"parent": networkInterface.Parents[0],
		"tags":   networkInterface.Tags,
		"vlan":   strconv.Itoa(networkInterface.VLAN.VID),
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceMaasNetworkInterfaceVlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	vlan, err := getVlan(client, fabric.ID, d.Get("vlan").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	params := getNetworkInterfaceVlanUpdateParams(d, parentID, vlan.ID)
	_, err = client.NetworkInterface.Update(machine.SystemID, id, params)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceMaasNetworkInterfaceVlanRead(ctx, d, meta)
}
func resourceMaasNetworkInterfaceVlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func getNetworkInterfaceVlanParams(d *schema.ResourceData, parentID int, vlanID int) *entity.NetworkInterfaceVLANParams {
	return &entity.NetworkInterfaceVLANParams{
		AcceptRA: d.Get("accept_ra").(bool),
		MTU:      d.Get("mtu").(int),
		Parents:  []int{parentID},
		Tags:     strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		VLAN:     vlanID,
	}
}

func getNetworkInterfaceVlanUpdateParams(d *schema.ResourceData, parentID int, vlanID int) *entity.NetworkInterfaceUpdateParams {
	return &entity.NetworkInterfaceUpdateParams{
		AcceptRA: d.Get("accept_ra").(bool),
		MTU:      d.Get("mtu").(int),
		Parents:  []int{parentID},
		Tags:     strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		VLAN:     vlanID,
	}
}

func resourceMaasNetworkInterfaceVlanImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE:VLAN_ID", d.Id())
	}

	d.Set("machine", idParts[0])
	d.SetId(idParts[1])

	return []*schema.ResourceData{d}, nil
}
