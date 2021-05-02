package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The MAAS API key",
			},
			"api_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The MAAS API URL (eg: http://127.0.0.1:5240/MAAS)",
			},
			"api_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "2.0",
				Description: "The MAAS API version (default 2.0)",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"maas_instance":                   resourceMaasInstance(),
			"maas_pod":                        resourceMaasPod(),
			"maas_pod_machine":                resourceMaasPodMachine(),
			"maas_machine":                    resourceMaasMachine(),
			"maas_network_interface_physical": resourceMaasNetworkInterfacePhysical(),
			"maas_tag":                        resourceMaasTag(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"maas_fabric": dataSourceMaasFabric(),
			"maas_vlan":   dataSourceMaasVlan(),
			"maas_subnet": dataSourceMaasSubnet(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := Config{
		APIKey:     d.Get("api_key").(string),
		APIURL:     d.Get("api_url").(string),
		ApiVersion: d.Get("api_version").(string),
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := config.Client()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create MAAS client",
			Detail:   fmt.Sprintf("Unable to create authenticated MAAS client: %s", err),
		})
		return nil, diags
	}

	return c, diags
}
