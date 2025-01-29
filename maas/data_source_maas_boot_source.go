package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maas/gomaasclient/client"
)

func dataSourceMaasBootSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasBootSourceRead,

		Schema: map[string]*schema.Schema{
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the boot source.",
			},
			"keyring_data": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The data on the keyring for the boot source.",
			},
			"keyring_filename": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The filename on the keyring for the boot source.",
			},
			"resource_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource URI for the book source.",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time of most recent update of the boot source.",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the boot source.",
			},
		},
	}
}

func dataSourceMaasBootSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	bootsource, err := getBootSource(client, d.Get("url").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", bootsource.ID))

	return nil
}
