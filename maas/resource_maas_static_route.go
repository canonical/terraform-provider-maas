package maas

import (
	"context"
	"fmt"
	"strconv"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASStaticRoute() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS static routes.",
		CreateContext: resourceStaticRouteCreate,
		ReadContext:   resourceStaticRouteRead,
		UpdateContext: resourceStaticRouteUpdate,
		DeleteContext: resourceStaticRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"destination": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Destination subnet name for the route.",
			},
			"gateway_ip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP address of the gateway on the source subnet.",
			},
			"metric": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Weight of the route on a deployed machine. Defaults to 0.",
			},
			"source": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Source subnet name for the route.",
			},
		},
	}
}

func resourceStaticRouteCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	cfg := meta.(*ClientConfig)
	client := cfg.Client

	params, err := getStaticRouteParams(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	staticRoute, err := client.StaticRoutes.Create(params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", staticRoute.ID))

	return resourceStaticRouteRead(ctx, d, meta)
}

func resourceStaticRouteRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	cfg := meta.(*ClientConfig)
	client := cfg.Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	staticRoute, err := client.StaticRoute.Get(id)
	if err != nil {
		return diag.FromErr(err)
	}

	tfState := map[string]any{
		"destination": staticRoute.Destination.Name,
		"source":      staticRoute.Source.Name,
		"gateway_ip":  staticRoute.GatewayIP,
		"metric":      staticRoute.Metric,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceStaticRouteUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	cfg := meta.(*ClientConfig)
	client := cfg.Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	params, err := getStaticRouteParams(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := client.StaticRoute.Update(id, params); err != nil {
		return diag.FromErr(err)
	}

	return resourceStaticRouteRead(ctx, d, meta)
}

func resourceStaticRouteDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	cfg := meta.(*ClientConfig)
	client := cfg.Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.StaticRoute.Delete(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getStaticRouteParams(client *client.Client, d *schema.ResourceData) (*entity.StaticRouteParams, error) {
	metric := 0
	if v, ok := d.GetOk("metric"); ok {
		metric = v.(int)
	}

	params := entity.StaticRouteParams{
		Source:      d.Get("source").(string),
		Destination: d.Get("destination").(string),
		GatewayIP:   d.Get("gateway_ip").(string),
		Metric:      metric,
	}

	return &params, nil
}
