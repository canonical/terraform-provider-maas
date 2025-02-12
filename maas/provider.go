package maas

import (
	"context"
	"fmt"
	"os"

	"github.com/canonical/gomaasclient/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     os.Getenv("MAAS_API_KEY"),
				Description: "The MAAS API key",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     os.Getenv("MAAS_API_URL"),
				Description: "The MAAS API URL (eg: http://127.0.0.1:5240/MAAS)",
			},
			"api_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "2.0",
				Description: "The MAAS API version (default 2.0)",
			},
			"installation_method": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MAAS_INSTALLATION_METHOD", "snap"),
				Description: "The MAAS installation method. Valid options: `snap`, and `deb`.",
			},
			"tls_ca_cert_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Certificate CA bundle path to use to verify the MAAS certificate.",
				Default:     os.Getenv("MAAS_API_CACERT"),
			},
			"tls_insecure_skip_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     "false",
				Description: "Skip TLS certificate verification.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"maas_boot_source":                resourceMAASBootSource(),
			"maas_device":                     resourceMaasDevice(),
			"maas_instance":                   resourceMaasInstance(),
			"maas_vm_host":                    resourceMaasVMHost(),
			"maas_vm_host_machine":            resourceMaasVMHostMachine(),
			"maas_machine":                    resourceMaasMachine(),
			"maas_network_interface_bridge":   resourceMaasNetworkInterfaceBridge(),
			"maas_network_interface_bond":     resourceMaasNetworkInterfaceBond(),
			"maas_network_interface_physical": resourceMaasNetworkInterfacePhysical(),
			"maas_network_interface_vlan":     resourceMaasNetworkInterfaceVlan(),
			"maas_network_interface_link":     resourceMaasNetworkInterfaceLink(),
			"maas_fabric":                     resourceMaasFabric(),
			"maas_vlan":                       resourceMaasVlan(),
			"maas_subnet":                     resourceMaasSubnet(),
			"maas_subnet_ip_range":            resourceMaasSubnetIPRange(),
			"maas_dns_domain":                 resourceMaasDnsDomain(),
			"maas_dns_record":                 resourceMaasDnsRecord(),
			"maas_space":                      resourceMaasSpace(),
			"maas_block_device":               resourceMaasBlockDevice(),
			"maas_tag":                        resourceMaasTag(),
			"maas_user":                       resourceMaasUser(),
			"maas_resource_pool":              resourceMaasResourcePool(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"maas_boot_source":                dataSourceMaasBootSource(),
			"maas_fabric":                     dataSourceMaasFabric(),
			"maas_vlan":                       dataSourceMaasVlan(),
			"maas_subnet":                     dataSourceMaasSubnet(),
			"maas_machine":                    dataSourceMaasMachine(),
			"maas_network_interface_physical": dataSourceMaasNetworkInterfacePhysical(),
			"maas_device":                     dataSourceMaasDevice(),
			"maas_resource_pool":              dataSourceMaasResourcePool(),
			"maas_rack_controller":            dataSourceMaasRackController(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

type ClientConfig struct {
	Client             *client.Client
	InstallationMethod string
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	if apiKey == "" {
		return nil, diag.FromErr(fmt.Errorf("MAAS API key cannot be empty"))
	}
	apiURL := d.Get("api_url").(string)
	if apiURL == "" {
		return nil, diag.FromErr(fmt.Errorf("MAAS API URL cannot be empty"))
	}
	config := Config{
		APIKey:                apiKey,
		APIURL:                apiURL,
		ApiVersion:            d.Get("api_version").(string),
		TLSCACertPath:         d.Get("tls_ca_cert_path").(string),
		TLSInsecureSkipVerify: d.Get("tls_insecure_skip_verify").(bool),
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

	return &ClientConfig{Client: c, InstallationMethod: d.Get("installation_method").(string)}, diags
}
