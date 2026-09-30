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

func resourceMAASVLAN() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS network VLANs.",
		CreateContext: resourceVLANCreate,
		ReadContext:   resourceVLANRead,
		UpdateContext: resourceVLANUpdate,
		DeleteContext: resourceVLANDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected FABRIC:VLAN", d.Id())
				}

				client := meta.(*ClientConfig).Client

				fabric, err := getFabric(client, idParts[0])
				if err != nil {
					return nil, err
				}

				vlan, err := getVLAN(client, fabric.ID, idParts[1], lookupVLANByVID)
				if err != nil {
					return nil, err
				}

				tfState := map[string]any{
					"id":     fmt.Sprintf("%v", vlan.ID),
					"fabric": fmt.Sprintf("%v", fabric.ID),
					"vid":    vlan.VID,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"dhcp_on": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Boolean value. Whether or not DHCP should be managed on the new VLAN. This argument is computed if it's not set.",
			},
			"fabric": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (name or ID) of the fabric for the new VLAN.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The MTU to use on the new VLAN. This argument is computed if it's not set.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the new VLAN. This argument is computed if it's not set.",
			},
			"space": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The space of the new VLAN. Passing in an empty string (or the string `undefined`) will cause the VLAN to be placed in the `undefined` space. This argument is computed if it's not set.",
			},
			"vid": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The traffic segregation ID for the new VLAN.",
			},
		},
	}
}

func resourceVLANCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	vlan, err := client.VLANs.Create(fabric.ID, getVLANParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", vlan.ID))

	return resourceVLANUpdate(ctx, d, meta)
}

func resourceVLANRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	vlan, err := getVLAN(client, fabric.ID, d.Id(), lookupVLANByID)
	if err != nil {
		return diag.FromErr(err)
	}

	tfState := map[string]any{
		"mtu":     vlan.MTU,
		"dhcp_on": vlan.DHCPOn,
		"name":    vlan.Name,
		"space":   vlan.Space,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVLANUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	vlan, err := getVLAN(client, fabric.ID, d.Id(), lookupVLANByID)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := client.VLAN.Update(fabric.ID, vlan.VID, getVLANParams(d)); err != nil {
		return diag.FromErr(err)
	}

	return resourceVLANRead(ctx, d, meta)
}

func resourceVLANDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	vlan, err := getVLAN(client, fabric.ID, d.Id(), lookupVLANByID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.VLAN.Delete(fabric.ID, vlan.VID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getVLANParams(d *schema.ResourceData) *entity.VLANParams {
	return &entity.VLANParams{
		VID:    d.Get("vid").(int),
		MTU:    d.Get("mtu").(int),
		DHCPOn: d.Get("dhcp_on").(bool),
		Name:   d.Get("name").(string),
		Space:  d.Get("space").(string),
	}
}

// vlanLookup selects which VLAN attribute an identifier is matched against
// first. MAAS identifies VLANs within a fabric by both their database ID and
// their VID, and a fabric always contains a VID 0, so an identifier can
// legitimately match two different VLANs. Resolving deterministically keeps
// that collision from silently picking the wrong VLAN.
type vlanLookup int

const (
	// lookupVLANByID prefers the database ID, falling back to the VID.
	lookupVLANByID vlanLookup = iota
	// lookupVLANByVID prefers the VID, falling back to the database ID.
	lookupVLANByVID
)

func getVLAN(client *client.Client, fabricID int, identifier string, lookup vlanLookup) (*entity.VLAN, error) {
	vlans, err := client.VLANs.Get(fabricID)
	if err != nil {
		return nil, err
	}

	attrs := []func(entity.VLAN) string{
		func(v entity.VLAN) string { return strconv.Itoa(v.ID) },
		func(v entity.VLAN) string { return strconv.Itoa(v.VID) },
	}
	if lookup == lookupVLANByVID {
		attrs[0], attrs[1] = attrs[1], attrs[0]
	}

	for _, attr := range attrs {
		for i := range vlans {
			if attr(vlans[i]) == identifier {
				return &vlans[i], nil
			}
		}
	}

	return nil, fmt.Errorf("vlan (%s) was not found", identifier)
}
