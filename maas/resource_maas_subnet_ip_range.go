package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasSubnetIPRange() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSubnetIPRangeCreate,
		ReadContext:   resourceSubnetIPRangeRead,
		UpdateContext: resourceSubnetIPRangeUpdate,
		DeleteContext: resourceSubnetIPRangeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
				idParts := strings.Split(d.Id(), ":")
				var ipRange *entity.IPRange
				var err error
				if len(idParts) == 2 {
					if idParts[0] == "" || idParts[1] == "" {
						return nil, fmt.Errorf("unexpected format of ID (%q), expected START_IP:END_IP", d.Id())
					}
					ipRange, err = getSubnetIPRange(client, idParts[0], idParts[1])
					if err != nil {
						return nil, err
					}
				} else {
					id, err := strconv.Atoi(d.Id())
					if err != nil {
						return nil, err
					}
					ipRange, err = client.IPRange.Get(id)
					if err != nil {
						return nil, err
					}
				}
				tfState := map[string]interface{}{
					"id":       fmt.Sprintf("%v", ipRange.ID),
					"subnet":   fmt.Sprintf("%v", ipRange.Subnet.ID),
					"type":     ipRange.Type,
					"start_ip": ipRange.StartIP.String(),
					"end_ip":   ipRange.EndIP.String(),
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"subnet": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"dynamic", "reserved"}, false)),
			},
			"start_ip": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
			},
			"end_ip": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceSubnetIPRangeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	subnet, err := findSubnet(client, d.Get("subnet").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	ipRange, err := client.IPRanges.Create(getSubnetIPRangeParams(d, subnet.ID))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", ipRange.ID))

	return resourceSubnetIPRangeUpdate(ctx, d, m)
}

func resourceSubnetIPRangeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	ipRange, err := client.IPRange.Get(id)
	if err != nil {
		return diag.FromErr(err)
	}
	tfState := map[string]interface{}{
		"comment": ipRange.Comment,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSubnetIPRangeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	subnet, err := findSubnet(client, d.Get("subnet").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := client.IPRange.Update(id, getSubnetIPRangeParams(d, subnet.ID)); err != nil {
		return diag.FromErr(err)
	}

	return resourceSubnetIPRangeRead(ctx, d, m)
}

func resourceSubnetIPRangeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.IPRange.Delete(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getSubnetIPRangeParams(d *schema.ResourceData, subnetID int) *entity.IPRangeParams {
	return &entity.IPRangeParams{
		Subnet:  fmt.Sprintf("%v", subnetID),
		Type:    d.Get("type").(string),
		StartIP: d.Get("start_ip").(string),
		EndIP:   d.Get("end_ip").(string),
		Comment: d.Get("comment").(string),
	}
}

func getSubnetIPRange(client *client.Client, startIP string, endIP string) (*entity.IPRange, error) {
	ipRanges, err := client.IPRanges.Get()
	if err != nil {
		return nil, err
	}
	for _, ipr := range ipRanges {
		if ipr.StartIP.String() == startIP && ipr.EndIP.String() == endIP {
			return &ipr, nil
		}
	}
	return nil, fmt.Errorf("IP range (%s->%s) was not found", startIP, endIP)
}
