package maas

import (
	"context"
	"fmt"
	"strconv"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMAASReservedIP() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS reserved IPs (static DHCP leases). Requires MAAS 3.6 or later.",
		CreateContext: resourceReservedIPCreate,
		ReadContext:   resourceReservedIPRead,
		UpdateContext: resourceReservedIPUpdate,
		DeleteContext: resourceReservedIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
			return checkSemverConstraint(meta.(*ClientConfig).MAASVersion, ">=3.6.0")
		},
		Schema: map[string]*schema.Schema{
			"comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the reserved IP.",
			},
			"ip": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "The IP address to reserve.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
			},
			"mac_address": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "The MAC address of the host to bind the reserved IP to.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsMACAddress),
			},
			"subnet": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The ID of the subnet the reserved IP belongs to. If not set, MAAS auto-detects it from the IP address.",
			},
		},
	}
}

func resourceReservedIPCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	params := &entity.ReservedIPCreateParams{
		IP:         d.Get("ip").(string),
		MACAddress: d.Get("mac_address").(string),
		Comment:    d.Get("comment").(string),
	}

	if subnet, ok := d.GetOk("subnet"); ok {
		params.Subnet = subnet.(int)
	}

	reservedIP, err := client.ReservedIPs.Create(params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", reservedIP.ID))

	return resourceReservedIPRead(ctx, d, meta)
}

func resourceReservedIPRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	reservedIP, err := client.ReservedIP.Get(id)
	if err != nil {
		return unsetIfNotFoundError(d, err)
	}

	tfState := map[string]any{
		"ip":          reservedIP.IP,
		"mac_address": reservedIP.MACAddress,
		"comment":     reservedIP.Comment,
		"subnet":      reservedIP.Subnet.ID,
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceReservedIPUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := client.ReservedIP.Update(id, &entity.ReservedIPUpdateParams{
		Comment: d.Get("comment").(string),
	}); err != nil {
		return diag.FromErr(err)
	}

	return resourceReservedIPRead(ctx, d, meta)
}

func resourceReservedIPDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.ReservedIP.Delete(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
