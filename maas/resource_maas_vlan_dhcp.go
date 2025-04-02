package maas

import (
	"context"
	"fmt"
	"strconv"

	"slices"
	"log"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasVlanDHCP() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage DHCP on MAAS network VLANs.",
		CreateContext: resourceVlanDHCPCreate,
		ReadContext:   resourceVlanDHCPRead,
		UpdateContext: resourceVlanDHCPUpdate,
		DeleteContext: resourceVlanDHCPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"fabric": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Database ID of the fabric of the VLAN whose DHCP is managed.",
			},
			"ip_ranges": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of IP range ids to server DHCP to. IP ranges must be of type dynamic.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"primary_rack_controller": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"relay_vlan"},
				AtLeastOneOf:  []string{"primary_rack_controller", "relay_vlan"},
				Description:   "The system_id of the Rack controller to to use as primary for DHCP.",
			},
			"relay_vlan": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"primary_rack_controller", "secondary_rack_controller"},
				AtLeastOneOf:  []string{"primary_rack_controller", "relay_vlan"},
				Description:   "VID of the VLAN to to use as a relay for DHCP.",
			},
			"secondary_rack_controller": {
				Type:          schema.TypeString,
				Optional:      true,
				RequiredWith:  []string{"primary_rack_controller"},
				ConflictsWith: []string{"relay_vlan"},
				Description:   "The system_id of the Rack controller to to use as secondary for DHCP.",
			},
			"subnets": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of subnet ids to serve DHCP on their dynamic IP ranges.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"vlan": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "VID of the VLAN whose DHCP is managed.",
			},
		},
	}
}

func resourceVlanDHCPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Println("running resourceVlanDHCPCreate")
	client := meta.(*ClientConfig).Client

	err := confirmAllIPRangesDynamic(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = confirmAllSubnetsWithADynamicIPRange(client, d)
	if err != nil {
		return diag.FromErr(err)
	}
	fabricID := d.Get("fabric").(int)
	vlanID := d.Get("vlan").(int)
	params := getVlanDHCPParams(d)
	log.Printf("about to update vlan dhcp params with fabricID: %d, vlanID: %d, params: %+v", fabricID, vlanID, params)
	_, err = client.VLAN.Update(fabricID, vlanID, params)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.Itoa(vlanID))

	return resourceVlanDHCPRead(ctx, d, meta)
}

func resourceVlanDHCPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Println("running resourceVlanDHCPRead")
	client := meta.(*ClientConfig).Client

	vlanID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	fabricID := d.Get("fabric").(int)
	vlan, err := client.VLAN.Get(fabricID, vlanID)
	if err != nil {
		return diag.FromErr(err)
	}

	tfState := map[string]interface{}{
		"primary_rack_controller":   vlan.PrimaryRack,
		"secondary_rack_controller": vlan.SecondaryRack,
	}
	if vlan.RelayVLAN != nil {
		tfState["relay_vlan"] = vlan.RelayVLAN.ID
	} else {
		tfState["relay_vlan"] = 0
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVlanDHCPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Println("running resourceVlanDHCPUpdate")
	client := meta.(*ClientConfig).Client

	if d.HasChange("ip_ranges") {
		oldVal, newVal := d.GetChange("ip_ranges")
		return diag.Errorf("Changing 'ip_ranges' from %v to %v is not allowed. Please recreate the resource.", oldVal, newVal)
	}

	if d.HasChange("subnets") {
		oldVal, newVal := d.GetChange("subnets")
		return diag.Errorf("Changing 'subnets' from %v to %v is not allowed. Please recreate the resource.", oldVal, newVal)
	}

	vlanID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	fabricID := d.Get("fabric").(int)
	if _, err := client.VLAN.Update(fabricID, vlanID, getVlanDHCPParams(d)); err != nil {
		return diag.FromErr(err)
	}

	return resourceVlanDHCPRead(ctx, d, meta)
}

func resourceVlanDHCPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Println("running resourceVlanDHCPDelete")
	client := meta.(*ClientConfig).Client

	fabricID := d.Get("fabric").(int)
	vlanID:= d.Get("vlan").(int)
	_, err := client.VLAN.Update(fabricID, vlanID, &entity.VLANParams{
		PrimaryRack: "", SecondaryRack: "", RelayVLAN: 0,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getVlanDHCPParams(d *schema.ResourceData) *entity.VLANParams {
	log.Println("running getVlanDHCPParams")
	vlanParams := entity.VLANParams{
		DHCPOn: true,
	}
	
	if v, ok := d.GetOk("primary_rack_controller"); ok {
		vlanParams.PrimaryRack = v.(string)
	}
	if v, ok := d.GetOk("secondary_rack_controller"); ok {
		vlanParams.SecondaryRack = v.(string)
	}
	if v, ok := d.GetOk("relay_vlan"); ok {
		vlanParams.RelayVLAN = v.(int)
	}
	return &vlanParams
}

func confirmAllSubnetsWithADynamicIPRange(client *client.Client, d *schema.ResourceData) error {
	log.Println("running confirmAllSubnetsWithADynamicIPRange")
	for _, subnetID := range d.Get("subnets").(*schema.Set).List() {
		subnetIPRanges, err := client.Subnet.GetReservedIPRanges(subnetID.(int))
		if err != nil {
			return err
		}
		foundDynamic := false
		for _, ipRange := range subnetIPRanges {
			if slices.Contains(ipRange.Purpose, "dynamic") {
				foundDynamic = true
				break
			}
		}
		if !foundDynamic {
			return fmt.Errorf("subnet %s does not have any dynamic IP range", subnetID)
		}
	}

	return nil
}

func confirmAllIPRangesDynamic(client *client.Client, d *schema.ResourceData) error {
	log.Println("running confirmAllIPRangesDynamic")
	for _, ipRangeID := range d.Get("ip_ranges").(*schema.Set).List() {
		ipRange, err := client.IPRange.Get(ipRangeID.(int))
		if err != nil {
			return err
		}
		if ipRange.Type != "dynamic" {
			return fmt.Errorf("IP range %s is not dynamic", ipRangeID)
		}
	}

	return nil
}
