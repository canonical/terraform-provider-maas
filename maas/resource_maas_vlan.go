package maas

import (
	"context"
	"fmt"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasVlan() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS network VLANs.",
		CreateContext: resourceVlanCreate,
		ReadContext:   resourceVlanRead,
		UpdateContext: resourceVlanUpdate,
		DeleteContext: resourceVlanDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected FABRIC:VLAN", d.Id())
				}
				client := meta.(*ClientConfig).Client

				fabric, err := getFabric(client, idParts[0])
				if err != nil {
					return nil, err
				}
				vlan, err := getVlan(client, fabric.ID, idParts[1])
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
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

func resourceVlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	vlan, err := client.VLANs.Create(fabric.ID, getVlanParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", vlan.ID))

	return resourceVlanUpdate(ctx, d, meta)
}

func resourceVlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	vlan, err := getVlan(client, fabric.ID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	tfState := map[string]interface{}{
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

func resourceVlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	vlan, err := getVlan(client, fabric.ID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := client.VLAN.Update(fabric.ID, vlan.VID, getVlanParams(d)); err != nil {
		return diag.FromErr(err)
	}

	return resourceVlanRead(ctx, d, meta)
}

func resourceVlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	vlan, err := getVlan(client, fabric.ID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.VLAN.Delete(fabric.ID, vlan.VID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getVlanParams(d *schema.ResourceData) *entity.VLANParams {
	return &entity.VLANParams{
		VID:    d.Get("vid").(int),
		MTU:    d.Get("mtu").(int),
		DHCPOn: d.Get("dhcp_on").(bool),
		Name:   d.Get("name").(string),
		Space:  d.Get("space").(string),
	}
}

func findVlan(client *client.Client, fabricID int, identifier string) (*entity.VLAN, error) {
	vlans, err := client.VLANs.Get(fabricID)
	if err != nil {
		return nil, err
	}
	for _, v := range vlans {
		if fmt.Sprintf("%v", v.VID) == identifier || fmt.Sprintf("%v", v.ID) == identifier {
			return &v, nil
		}
	}
	return nil, nil
}

func getVlan(client *client.Client, fabricID int, identifier string) (*entity.VLAN, error) {
	vlan, err := findVlan(client, fabricID, identifier)
	if err != nil {
		return nil, err
	}
	if vlan == nil {
		return nil, fmt.Errorf("vlan (%s) was not found", identifier)
	}
	return vlan, nil
}
