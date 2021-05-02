package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
)

func dataSourceMaasFabric() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFabricRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceFabricRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	fabrics, err := client.Fabrics.Get()
	if err != nil {
		return diag.FromErr(err)
	}

	fabricName := d.Get("name").(string)
	for _, fabric := range fabrics {
		if fabric.Name != fabricName {
			continue
		}
		d.SetId(fmt.Sprintf("%v", fabric.ID))
		return nil
	}

	return diag.FromErr(fmt.Errorf("could not find matching fabric"))
}
