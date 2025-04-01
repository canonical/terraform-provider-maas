package maas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasDevices() *schema.Resource {
	return &schema.Resource{
		Description: "Lists MAAS devices visible to the user.",
		ReadContext: dataSourceDevicesRead,

		Schema: map[string]*schema.Schema{
			"devices": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of devices visible to the user.",
				Elem: &schema.Resource{
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
							Computed:    true,
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
				},
			},
		},
	}
}

func dataSourceDevicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	devices, err := client.Devices.Get()
	if err != nil {
		return diag.FromErr(err)
	}

	items := []map[string]interface{}{}
	for _, device := range devices {
		item := map[string]interface{}{
			"description":        device.Description,
			"domain":             device.Domain.Name,
			"fqdn":               device.FQDN,
			"hostname":           device.Hostname,
			"ip_addresses":       []string{},
			"network_interfaces": []map[string]interface{}{},
			"owner":              device.Owner,
			"zone":               device.Zone.Name,
		}

		ipAddresses := make([]string, len(device.IPAddresses))
		for i, ipAddress := range device.IPAddresses {
			ipAddresses[i] = ipAddress.String()
		}
		item["ip_addresses"] = ipAddresses

		networkInterfaces := make([]map[string]interface{}, len(device.InterfaceSet))
		for i, networkInterface := range device.InterfaceSet {
			networkInterfaces[i] = map[string]interface{}{
				"id":          networkInterface.ID,
				"mac_address": networkInterface.MACAddress,
				"name":        networkInterface.Name,
			}
		}
		item["network_interfaces"] = networkInterfaces

		items = append(items, item)
	}

	if err := d.Set("devices", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(devices[0].SystemID)

	return nil
}
