package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasResourcePool() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an existing MAAS resource pool.",
		ReadContext: dataSourceResourcePoolRead,

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
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

func dataSourceResourcePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	resourcePool, err := getResourcePool(client, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", resourcePool.ID))

	d.Set("description", resourcePool.Description)
	d.Set("name", resourcePool.Name)

	return nil
}
