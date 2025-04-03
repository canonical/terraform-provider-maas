package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASNetworkInterfaceVLAN() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS network VLANs.",
		CreateContext: resourceNetworkInterfaceVLANCreate,
		ReadContext:   resourceNetworkInterfaceVLANRead,
		UpdateContext: resourceNetworkInterfaceVLANUpdate,
		DeleteContext: resourceNetworkInterfaceVLANDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNetworkInterfaceVLANImport,
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
				Description: "The identifier (name or ID) of the fabric for the new VLAN interface.",
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (system ID, hostname, or FQDN) of the machine with the VLAN interface.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The MTU of the VLAN interface.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the VLAN interface.",
			},
			"parent": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent interface name for this VLAN interface.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A set of tag names to be assigned to the VLAN interface.",
			},
			"vlan": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Database ID of the VLAN the VLAN interface is connected to.",
			},
		},
	}
}

func resourceNetworkInterfaceVLANCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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

	vlan, err := getVLAN(client, fabric.ID, strconv.Itoa(d.Get("vlan").(int)))
	if err != nil {
		return diag.FromErr(err)
	}

	params := getNetworkInterfaceVLANParams(d, parentID, vlan.ID)

	networkInterface, err := client.NetworkInterfaces.CreateVLAN(machine.SystemID, params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(networkInterface.ID))

	return resourceNetworkInterfaceVLANRead(ctx, d, meta)
}

func resourceNetworkInterfaceVLANRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	p := networkInterface.Params.(map[string]any)
	if _, ok := p["accept-ra"]; ok {
		d.Set("accept_ra", p["accept-ra"].(bool))
	} else {
		d.Set("accept_ra", false)
	}

	tfState := map[string]any{
		"mtu":    networkInterface.EffectiveMTU,
		"name":   networkInterface.Name,
		"parent": networkInterface.Parents[0],
		"tags":   networkInterface.Tags,
		"vlan":   networkInterface.VLAN.ID,
	}
	if _, ok := d.GetOk("fabric"); !ok {
		tfState["fabric"] = strconv.Itoa(networkInterface.VLAN.FabricID)
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfaceVLANUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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

	vlan, err := getVLAN(client, fabric.ID, strconv.Itoa(d.Get("vlan").(int)))
	if err != nil {
		return diag.FromErr(err)
	}

	params := getNetworkInterfaceVLANUpdateParams(d, parentID, vlan.ID)

	_, err = client.NetworkInterface.Update(machine.SystemID, id, params)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkInterfaceVLANRead(ctx, d, meta)
}
func resourceNetworkInterfaceVLANDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func getNetworkInterfaceVLANParams(d *schema.ResourceData, parentID int, vlanID int) *entity.NetworkInterfaceVLANParams {
	return &entity.NetworkInterfaceVLANParams{
		AcceptRA: d.Get("accept_ra").(bool),
		MTU:      d.Get("mtu").(int),
		Parents:  []int{parentID},
		Tags:     strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		VLAN:     vlanID,
	}
}

func getNetworkInterfaceVLANUpdateParams(d *schema.ResourceData, parentID int, vlanID int) *entity.NetworkInterfaceUpdateParams {
	return &entity.NetworkInterfaceUpdateParams{
		AcceptRA: d.Get("accept_ra").(bool),
		MTU:      d.Get("mtu").(int),
		Parents:  []int{parentID},
		Tags:     strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		VLAN:     vlanID,
	}
}

func resourceNetworkInterfaceVLANImport(d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE:VLAN_INTERFACE_ID", d.Id())
	}

	d.Set("machine", idParts[0])
	d.SetId(idParts[1])

	return []*schema.ResourceData{d}, nil
}
