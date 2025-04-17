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
				Description: "The MAAS API key. If not provided, it will be read from the MAAS_API_KEY environment variable.",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     os.Getenv("MAAS_API_URL"),
				Description: "The MAAS API URL (eg: http://127.0.0.1:5240/MAAS). If not provided, it will be read from the MAAS_API_URL environment variable.",
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
				Description: "Certificate CA bundle path to use to verify the MAAS certificate. If not provided, it will be read from the MAAS_API_CACERT environment variable.",
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
			"maas_boot_source_selection":      resourceMAASBootSourceSelection(),
			"maas_boot_source":                resourceMAASBootSource(),
			"maas_device":                     resourceMAASDevice(),
			"maas_instance":                   resourceMAASInstance(),
			"maas_vm_host":                    resourceMAASVMHost(),
			"maas_vm_host_machine":            resourceMAASVMHostMachine(),
			"maas_machine":                    resourceMAASMachine(),
			"maas_network_interface_bridge":   resourceMAASNetworkInterfaceBridge(),
			"maas_network_interface_bond":     resourceMAASNetworkInterfaceBond(),
			"maas_network_interface_physical": resourceMAASNetworkInterfacePhysical(),
			"maas_network_interface_vlan":     resourceMAASNetworkInterfaceVLAN(),
			"maas_network_interface_link":     resourceMAASNetworkInterfaceLink(),
			"maas_fabric":                     resourceMAASFabric(),
			"maas_vlan":                       resourceMAASVLAN(),
			"maas_vlan_dhcp":                  resourceMAASVLANDHCP(),
			"maas_subnet":                     resourceMAASSubnet(),
			"maas_subnet_ip_range":            resourceMAASSubnetIPRange(),
			"maas_dns_domain":                 resourceMAASDNSDomain(),
			"maas_dns_record":                 resourceMAASDNSRecord(),
			"maas_space":                      resourceMAASSpace(),
			"maas_block_device":               resourceMAASBlockDevice(),
			"maas_block_device_tag":           resourceMAASBlockDeviceTag(),
			"maas_tag":                        resourceMAASTag(),
			"maas_network_interface_tag":      resourceMAASNetworkInterfaceTag(),
			"maas_user":                       resourceMAASUser(),
			"maas_resource_pool":              resourceMAASResourcePool(),
			"maas_volume_group":               resourceMAASVolumeGroup(),
			"maas_zone":                       resourceMAASZone(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"maas_boot_source":                dataSourceMAASBootSource(),
			"maas_boot_source_selection":      dataSourceMAASBootSourceSelection(),
			"maas_fabric":                     dataSourceMAASFabric(),
			"maas_vlan":                       dataSourceMAASVLAN(),
			"maas_subnet":                     dataSourceMAASSubnet(),
			"maas_machine":                    dataSourceMAASMachine(),
			"maas_machines":                   dataSourceMAASMachines(),
			"maas_network_interface_physical": dataSourceMAASNetworkInterfacePhysical(),
			"maas_device":                     dataSourceMAASDevice(),
			"maas_devices":                    dataSourceMAASDevices(),
			"maas_resource_pool":              dataSourceMAASResourcePool(),
			"maas_rack_controller":            dataSourceMAASRackController(),
			"maas_zone":                       dataSourceMAASZone(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

type ClientConfig struct {
	Client             *client.Client
	InstallationMethod string
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
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
		APIVersion:            d.Get("api_version").(string),
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
