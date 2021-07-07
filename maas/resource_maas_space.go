package maas

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasSpace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSpaceCreate,
		ReadContext:   resourceSpaceRead,
		UpdateContext: resourceSpaceUpdate,
		DeleteContext: resourceSpaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
				space, err := getSpace(client, d.Id())
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
					"id":   fmt.Sprintf("%v", space.ID),
					"name": space.Name,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSpaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	space, err := client.Spaces.Create(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", space.ID))

	return resourceSpaceUpdate(ctx, d, m)
}

func resourceSpaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := client.Space.Get(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSpaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := client.Space.Update(id, d.Get("name").(string)); err != nil {
		return diag.FromErr(err)
	}

	return resourceSpaceRead(ctx, d, m)
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.Space.Delete(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func findSpace(client *client.Client, identifier string) (*entity.Space, error) {
	spaces, err := client.Spaces.Get()
	if err != nil {
		return nil, err
	}
	for _, s := range spaces {
		if fmt.Sprintf("%v", s.ID) == identifier || s.Name == identifier {
			return &s, nil
		}
	}
	return nil, nil
}

func getSpace(client *client.Client, identifier string) (*entity.Space, error) {
	space, err := findSpace(client, identifier)
	if err != nil {
		return nil, err
	}
	if space == nil {
		return nil, fmt.Errorf("space (%s) was not found", identifier)
	}
	return space, nil
}
