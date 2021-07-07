package maas

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasVlan() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVlanCreate,
		ReadContext:   resourceVlanRead,
		UpdateContext: resourceVlanUpdate,
		DeleteContext: resourceVlanDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected FABRIC:VLAN", d.Id())
				}
				client := m.(*client.Client)
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
			"fabric": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vid": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"mtu": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"dhcp_on": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"space": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceVlanCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	fabric, err := getFabric(client, d.Get("fabric").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	vlan, err := client.VLANs.Create(fabric.ID, getVlanParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", vlan.ID))

	return resourceVlanUpdate(ctx, d, m)
}

func resourceVlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

func resourceVlanUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

	return resourceVlanRead(ctx, d, m)
}

func resourceVlanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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
