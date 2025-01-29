package maas

import (
	"github.com/maas/gomaasclient/client"
)

func dataSourceMaasBootSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasBootSourceRead,

		Schema: map[string]*schema.Schema{
			"Created": {
				Type:		 schema.TypeString,
				Computed:	 true,
				Description: "The creation time of the boot source.",
			},
			"Updated": {
				Type:		 schema.TypeString,
				Computed:	 true,
				Description: "The time of most recent update of the boot source.",
			},
			"URL": {
				Type:		 schema.TypeString,
				Computed:	 true,
				Description: "The URL of the boot source.",
			},
			"KeyringFilename": {
				Type:		 schema.TypeString,
				Computed:	 true,
				Description: "The filename on the keyring for the boot source.",
			},
			"KeyringData": {
				Type:		 schema.TypeString,
				Computed:	 true,
				Description: "The data on the keyring for the boot source.",
			},
			"ResourceURI": {
				Type:		 schema.TypeString,
				Computed:	 true,
				Description: "The resource URI for the book source.",
			},
		}
	}
}

func dataSourceMaasBootSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	bootsource, err := getBootSource(client, d.GetOk("URL").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", fabric.ID))

	return nil
}
