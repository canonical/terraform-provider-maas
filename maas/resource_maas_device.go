package maas

import (
	"context"
	"fmt"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasDevice() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS devices.",
		CreateContext: resourceDeviceCreate,
		ReadContext:   resourceDeviceRead,
		UpdateContext: resourceDeviceUpdate,
		DeleteContext: resourceDeviceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				device, err := getDevice(client, d.Id())
				if err != nil {
					return nil, err
				}
				d.SetId(device.SystemID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the device.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
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
				Optional:    true,
				ForceNew:    true,
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
				Required:    true,
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
							Required:    true,
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
				Optional:    true,
				Computed:    true,
				Description: "The zone of the device.",
			},
		},
	}
}

func expandNetworkInterfacesItems(items []interface{}) []string {
	networkInterfacesItems := make([]string, 0)
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		networkInterfacesItems = append(networkInterfacesItems, itemMap["mac_address"].(string))
	}
	return networkInterfacesItems
}

func resourceDeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("Device creating"))
	client := meta.(*ClientConfig).Client

	deviceParams := entity.DeviceCreateParams{
		Description:  d.Get("description").(string),
		Domain:       d.Get("domain").(string),
		Hostname:     d.Get("hostname").(string),
		MacAddresses: expandNetworkInterfacesItems(d.Get("network_interfaces").(*schema.Set).List()),
	}

	device, err := client.Devices.Create(&deviceParams)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(device.SystemID)

	return resourceDeviceRead(ctx, d, meta)
}

func resourceDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("Device updating"))
	client := meta.(*ClientConfig).Client

	deviceParams := entity.DeviceUpdateParams{
		Description: d.Get("description").(string),
		Domain:      d.Get("domain").(string),
		Hostname:    d.Get("hostname").(string),
		Zone:        d.Get("zone").(string),
	}
	device, err := client.Device.Update(d.Id(), &deviceParams)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(device.SystemID)

	if d.HasChange("network_interfaces") {
		newInterfaces := d.Get("network_interfaces").(*schema.Set).List()

		// Get all existing interfaces from MAAS
		device, err := client.Device.Get(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
		existingInterfaces := device.InterfaceSet
		existingInterfacesMap := make(map[int]entity.NetworkInterface)
		for _, existingInterface := range existingInterfaces {
			tflog.Debug(ctx, fmt.Sprintf("Existing network interface %s with MAC address %s\n", existingInterface.Name, existingInterface.MACAddress))
			existingInterfacesMap[existingInterface.ID] = existingInterface
		}

		// Create a map to keep track of which interfaces to keep (update or create) or delete
		keepMap := make(map[int]bool)
		for _, existingInterface := range existingInterfaces {
			keepMap[existingInterface.ID] = false
		}
		for _, newInterface := range newInterfaces {
			keepMap[newInterface.(map[string]interface{})["id"].(int)] = true
		}

		// Create, update, or delete interfaces
		for id, keep := range keepMap {
			if !keep {
				tflog.Debug(ctx, fmt.Sprintf("Deleting network interface %d", id))
				err := client.NetworkInterface.Delete(d.Id(), id)
				if err != nil {
					return diag.FromErr(err)
				}
			} else {
				if networkInterface, ok := existingInterfacesMap[id]; ok {
					networkInterfaceParams := entity.NetworkInterfaceUpdateParams{
						MACAddress: networkInterface.MACAddress,
						Name:       networkInterface.Name,
					}
					tflog.Debug(ctx, fmt.Sprintf("Updating network interface %s with MAC address %s", networkInterface.Name, networkInterfaceParams.MACAddress))
					_, err := client.NetworkInterface.Update(d.Id(), id, &networkInterfaceParams)
					if err != nil {
						return diag.FromErr(err)
					}
				} else {
					networkInterfaceParams := entity.NetworkInterfacePhysicalParams{
						MACAddress: networkInterface.MACAddress,
						Name:       networkInterface.Name,
					}
					tflog.Debug(ctx, fmt.Sprintf("Creating network interface %s with MAC address %s", networkInterfaceParams.Name, networkInterfaceParams.MACAddress))
					_, err := client.NetworkInterfaces.CreatePhysical(d.Id(), &networkInterfaceParams)
					if err != nil {
						return diag.FromErr(err)
					}
				}
			}
		}
	}

	return resourceDeviceRead(ctx, d, meta)
}

func resourceDeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	return diag.FromErr(client.Device.Delete(d.Id()))
}

func resourceDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	device, err := getDevice(client, d.Id())
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
