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

func resourceMaasFabric() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS network fabrics.",
		CreateContext: resourceFabricCreate,
		ReadContext:   resourceFabricRead,
		UpdateContext: resourceFabricUpdate,
		DeleteContext: resourceFabricDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				fabric, err := getFabric(client, d.Id())
				if err != nil {
					return nil, err
				}
				if err := d.Set("name", fabric.Name); err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("%v", fabric.ID))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The fabric name.",
			},
		},
	}
}

func resourceFabricCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := client.Fabrics.Create(getFabricParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", fabric.ID))

	return resourceFabricUpdate(ctx, d, meta)
}

func resourceFabricRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := client.Fabric.Get(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFabricUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := client.Fabric.Update(id, getFabricParams(d)); err != nil {
		return diag.FromErr(err)
	}

	return resourceFabricRead(ctx, d, meta)
}

func resourceFabricDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.Fabric.Delete(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getFabricParams(d *schema.ResourceData) *entity.FabricParams {
	return &entity.FabricParams{
		Name: d.Get("name").(string),
	}
}

func findFabric(client *client.Client, identifier string) (*entity.Fabric, error) {
	fabrics, err := client.Fabrics.Get()
	if err != nil {
		return nil, err
	}
	for _, f := range fabrics {
		if fmt.Sprintf("%v", f.ID) == identifier || f.Name == identifier {
			return &f, nil
		}
	}
	return nil, nil
}

func getFabric(client *client.Client, identifier string) (*entity.Fabric, error) {
	fabric, err := findFabric(client, identifier)
	if err != nil {
		return nil, err
	}
	if fabric == nil {
		return nil, fmt.Errorf("fabric (%s) was not found", identifier)
	}
	return fabric, nil
}
