package maas

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASVLANDHCP() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage DHCP on MAAS network VLANs.",
		CreateContext: resourceVLANDHCPCreate,
		ReadContext:   resourceVLANDHCPRead,
		UpdateContext: resourceVLANDHCPUpdate,
		DeleteContext: resourceVLANDHCPDelete,

		Schema: map[string]*schema.Schema{
			"fabric": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Database ID of the fabric of the VLAN whose DHCP is managed. This parameter `fabric` and `vlan` are used to identify the VLAN.",
			},
			"ip_ranges": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of IP range ids to serve DHCP to. IP ranges must be of type dynamic.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"primary_rack_controller": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"relay_vlan"},
				AtLeastOneOf:  []string{"primary_rack_controller", "relay_vlan"},
				Description:   "The system_id of the Rack controller to use as primary for DHCP.",
			},
			"relay_vlan": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"primary_rack_controller", "secondary_rack_controller"},
				AtLeastOneOf:  []string{"primary_rack_controller", "relay_vlan"},
				Description:   "Database ID of the VLAN to use as a relay for DHCP.",
			},
			"secondary_rack_controller": {
				Type:          schema.TypeString,
				Optional:      true,
				RequiredWith:  []string{"primary_rack_controller"},
				ConflictsWith: []string{"relay_vlan"},
				Description:   "The system_id of the Rack controller to use as secondary for DHCP.",
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
				ForceNew:    true,
				Description: "VID of the VLAN whose DHCP is managed. This parameter `vlan` and `fabric` are used to identify the VLAN.",
			},
		},
	}
}

func resourceVLANDHCPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client
	// Validation
	err := confirmAllIPRangesDynamic(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = confirmAllSubnetsWithADynamicIPRange(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = confirmIPRangeSubnetsInVLAN(client, d)
	if err != nil {
		return diag.FromErr(err)
	}
	// Turn on DHCP
	fabricID := d.Get("fabric").(int)
	vlanID := d.Get("vlan").(int)
	params := getVLANDHCPParams(d)

	_, err = client.VLAN.Update(fabricID, vlanID, params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d/%d", fabricID, vlanID))

	return resourceVLANDHCPRead(ctx, d, meta)
}

func resourceVLANDHCPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabricID, vlanID, err := SplitStateIDIntoInts(d.Id(), "/")
	if err != nil {
		return diag.FromErr(err)
	}

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

func resourceVLANDHCPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	if d.HasChange("ip_ranges") {
		oldVal, newVal := d.GetChange("ip_ranges")
		return diag.Errorf("Changing 'ip_ranges' from %v to %v is not allowed. Please recreate the resource.", oldVal, newVal)
	}

	if d.HasChange("subnets") {
		oldVal, newVal := d.GetChange("subnets")
		return diag.Errorf("Changing 'subnets' from %v to %v is not allowed. Please recreate the resource.", oldVal, newVal)
	}

	fabricID, vlanID, err := SplitStateIDIntoInts(d.Id(), "/")
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := client.VLAN.Update(fabricID, vlanID, getVLANDHCPParams(d)); err != nil {
		return diag.FromErr(err)
	}

	return resourceVLANDHCPRead(ctx, d, meta)
}

func resourceVLANDHCPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabricID := d.Get("fabric").(int)
	vlanID := d.Get("vlan").(int)

	// gomaasclient requires a pointer to an empty string in order to nil the values below
	nilValue := ""

	_, err := client.VLAN.Update(fabricID, vlanID, &entity.VLANParams{
		PrimaryRack: &nilValue, SecondaryRack: &nilValue, RelayVLAN: &nilValue, DHCPOn: false,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getVLANDHCPParams(d *schema.ResourceData) *entity.VLANParams {
	vlanParams := entity.VLANParams{}
	if v, ok := d.GetOk("primary_rack_controller"); ok {
		vlanParams.DHCPOn = true
		primaryRack := v.(string)
		vlanParams.PrimaryRack = &primaryRack
	}

	if v, ok := d.GetOk("secondary_rack_controller"); ok {
		secondaryRack := v.(string)
		vlanParams.SecondaryRack = &secondaryRack
	}

	if v, ok := d.GetOk("relay_vlan"); ok {
		relayVLAN := strconv.Itoa(v.(int))
		vlanParams.RelayVLAN = &relayVLAN
	}

	return &vlanParams
}

func confirmAllSubnetsWithADynamicIPRange(client *client.Client, d *schema.ResourceData) error {
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
			return fmt.Errorf("subnet %s does not have any dynamic IP ranges", subnetID)
		}
	}

	return nil
}

func confirmAllIPRangesDynamic(client *client.Client, d *schema.ResourceData) error {
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

func confirmIPRangeSubnetsInVLAN(client *client.Client, d *schema.ResourceData) error {
	expectedVLANVID := d.Get("vlan").(int)
	expectedFabricID := d.Get("fabric").(int)

	for _, ipRangeID := range d.Get("ip_ranges").(*schema.Set).List() {
		ipRange, err := client.IPRange.Get(ipRangeID.(int))
		if err != nil {
			return err
		}

		if ipRange.Subnet.VLAN.FabricID != expectedFabricID || ipRange.Subnet.VLAN.VID != expectedVLANVID {
			return fmt.Errorf("IP range id=%d in fabric id=%d, vlan vid=%d and subnet id=%d is not in the same VLAN as the VLAN DHCP resource, with fabric id=%d and vlan vid=%d", ipRangeID, ipRange.Subnet.VLAN.FabricID, ipRange.Subnet.VLAN.VID, ipRange.Subnet.ID, expectedFabricID, expectedVLANVID)
		}
	}

	return nil
}
