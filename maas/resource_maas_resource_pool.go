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

func resourceMaasResourcePool() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS resource pools.",
		CreateContext: resourceResourcePoolCreate,
		ReadContext:   resourceResourcePoolRead,
		UpdateContext: resourceResourcePoolUpdate,
		DeleteContext: resourceResourcePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				resourcePool, err := getResourcePool(client, d.Id())
				if err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("%v", resourcePool.ID))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the resource pool.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource pool.",
			},
		},
	}
}

func resourceResourcePoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	resourcePoolParams := entity.ResourcePoolParams{
		Description: d.Get("description").(string),
		Name:        d.Get("name").(string),
	}

	resourcePool, err := client.ResourcePools.Create(&resourcePoolParams)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", resourcePool.ID))

	return resourceResourcePoolRead(ctx, d, meta)
}

func resourceResourcePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resourcePoolParams := entity.ResourcePoolParams{
		Description: d.Get("description").(string),
		Name:        d.Get("name").(string),
	}

	resourcePool, err := client.ResourcePool.Update(id, &resourcePoolParams)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", resourcePool.ID))

	return resourceResourcePoolRead(ctx, d, meta)
}

func resourceResourcePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(client.ResourcePool.Delete(id))
}

func resourceResourcePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	resourcePool, err := getResourcePool(client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", resourcePool.ID))

	d.Set("description", resourcePool.Description)
	d.Set("name", resourcePool.Name)

	return nil
}

func getResourcePool(client *client.Client, identifier string) (*entity.ResourcePool, error) {
	resourcePool, err := findResourcePool(client, identifier)
	if err != nil {
		return nil, err
	}
	if resourcePool == nil {
		return nil, fmt.Errorf("resource pool (%s) was not found", identifier)
	}
	return resourcePool, nil
}

func findResourcePool(client *client.Client, identifier string) (*entity.ResourcePool, error) {
	resourcePools, err := client.ResourcePools.Get()
	if err != nil {
		return nil, err
	}
	for _, d := range resourcePools {
		if fmt.Sprintf("%v", d.ID) == identifier || d.Name == identifier {
			return &d, nil
		}
	}
	return nil, nil
}
