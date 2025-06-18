package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMAASConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about a particular MAAS configuration setting. See MAAS server in the [MAAS API documentation](https://maas.io/docs/api) for a full list of keys and their values.",
		ReadContext: dataSourceMAASConfigurationRead,

		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Key corresponding to the configuration setting.",
			},
			"value": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Value for the configuration setting.",
			},
		},
	}
}

func dataSourceMAASConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	key := d.Get("key").(string)

	value, err := client.MAASServer.Get(key)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting key %v from MAAS: %v", key, err))
	}

	d.SetId(key)

	tfState := map[string]any{
		"key":   key,
		"value": NormalizeConfigValue(value),
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(fmt.Errorf("error setting Terraform state for key %v: %v", key, err))
	}

	return nil
}
