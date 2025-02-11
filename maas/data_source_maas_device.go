package maas

import (
	"context"
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasDevice() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an existing MAAS device.",
		ReadContext: dataSourceDeviceRead,

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the device.",
			},
			"domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The domain of the device.",
			},
			"fqdn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The device FQDN.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The device hostname.",
			},
			"ip_addresses": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of IP addressed assigned to the device.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"network_interfaces": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of network interfaces attached to the device.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The id of the network interface.",
						},
						"mac_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "MAC address of the network interface.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the network interface.",
						},
					},
				},
			},
			"owner": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The owner of the device.",
			},
			"zone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The zone of the device.",
			},
		},
	}
}

func dataSourceDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	device, err := getDevice(client, d.Get("hostname").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(device.SystemID)

	d.Set("description", device.Description)
	d.Set("domain", device.Domain.Name)
	d.Set("fqdn", device.FQDN)
	d.Set("hostname", device.Hostname)
	d.Set("owner", device.Owner)
	d.Set("zone", device.Zone.Name)

	ipAddresses := make([]string, len(device.IPAddresses))
	for i, ip := range device.IPAddresses {
		ipAddresses[i] = ip.String()
	}
	if err := d.Set("ip_addresses", ipAddresses); err != nil {
		return diag.FromErr(err)
	}

	networkInterfaces := make([]map[string]interface{}, len(device.InterfaceSet))
	for i, networkInterface := range device.InterfaceSet {
		networkInterfaces[i] = map[string]interface{}{
			"id":          networkInterface.ID,
			"mac_address": networkInterface.MACAddress,
			"name":        networkInterface.Name,
		}
	}
	if err := d.Set("network_interfaces", networkInterfaces); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getDevice(client *client.Client, identifier string) (*entity.Device, error) {
	device, err := findDevice(client, identifier)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, fmt.Errorf("device (%s) was not found", identifier)
	}
	return device, nil
}

func findDevice(client *client.Client, identifier string) (*entity.Device, error) {
	devices, err := client.Devices.Get()
	if err != nil {
		return nil, err
	}
	for _, d := range devices {
		if d.SystemID == identifier || d.Hostname == identifier {
			return &d, nil
		}
	}
	return nil, nil
}
