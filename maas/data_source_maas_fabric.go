package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasFabric() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an existing MAAS network fabric.",
		ReadContext: dataSourceFabricRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The fabric name.",
			},
		},
	}
}

func dataSourceFabricRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	fabric, err := getFabric(client, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", fabric.ID))

	return nil
}
