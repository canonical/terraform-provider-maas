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
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				cfg := meta.(*ClientConfig)
				client := cfg.Client
				staticRoute, err := getStaticRoute(client, d.Id())
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
					"id":          fmt.Sprintf("%v", staticRoute.ID),
					"source":      staticRoute.Source.Name,
					"destination": staticRoute.Destination.Name,
					"gateway_ip":  staticRoute.GatewayIP,
					"metric":      staticRoute.Metric,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
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

func resourceStaticRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return resourceStaticRouteUpdate(ctx, d, meta)
}

func resourceStaticRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	source := staticRoute.Source.Name
	destination := staticRoute.Destination.Name
	gatewayIP := staticRoute.GatewayIP
	metric := staticRoute.Metric

	tfState := map[string]interface{}{
		"destination": destination,
		"source":      source,
		"gateway_ip":  gatewayIP,
		"metric":      metric,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceStaticRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceStaticRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func findStaticRoute(client *client.Client, identifier string) (*entity.StaticRoute, error) {
	staticRoutes, err := client.StaticRoutes.Get()
	if err != nil {
		return nil, err
	}

	for _, s := range staticRoutes {
		if fmt.Sprintf("%v", s.ID) == identifier {
			return &s, nil
		}
	}

	return nil, nil
}

func getStaticRoute(client *client.Client, identifier string) (*entity.StaticRoute, error) {
	staticRoute, err := findStaticRoute(client, identifier)
	if err != nil {
		return nil, err
	}

	if staticRoute == nil {
		return nil, fmt.Errorf("staticRoute (%s) was not found", identifier)
	}

	return staticRoute, nil
}
